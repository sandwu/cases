import (
	"context"
	"errors"
	mongodrv "go.mongodb.org/mongo-driver/mongo"
	"sangfor.com/csc/csc/internal/app/cscborder/constants"
	"sangfor.com/csc/csc/internal/app/cscborder/logger"
	"sangfor.com/csc/csc/internal/app/cscborder/module/border/model/table"
	mongopo "sangfor.com/csc/csc/internal/app/cscborder/module/border/mongo_repo"
	pgrepo "sangfor.com/csc/csc/internal/app/cscborder/module/border/repo"
	cr "sangfor.com/csc/csc/internal/pkg/common-request"
	"sangfor.com/csc/csc/internal/pkg/model"
)
func UpdateAccessibleRelationSubTask() {
	// 从t_border_device_region_relationship取出region_source为1的数据，取出region_id，region_name，cloud_name，delete
	repo := pgrepo.NewBorderRepositoryInstance()

	// regionsList 是准备要进行处理的数据
	regionsList, err := repo.SelectAccessRelationshipByRegionSource(constants.REGION_SOURCE_CLOUD)
	if err != nil {
		logger.Logger.Errorf("Update accessible relation task fail, failed to query access relationship from pg! err: %s", err.Error())
		return
	}
	if len(regionsList) == 0 {
		return
	}

	// uniqueRegions regionsList去重后的数据
	uniqueRegionIds := make(map[string]struct{})
	uniqueGatewayIds := make(map[string]struct{})
	var regionIds []string
	var gatewayIds []string
	for _, region := range regionsList {
		if _, exists := uniqueRegionIds[region.RegionId]; !exists {
			uniqueRegionIds[region.RegionId] = struct{}{}
			regionIds = append(regionIds, region.RegionId)
		}
		if region.GatewaySource == constants.CLOUD_DEVICE {
			if _, exists := uniqueGatewayIds[region.GatewayId]; !exists {
				uniqueGatewayIds[region.GatewayId] = struct{}{}
				gatewayIds = append(gatewayIds, region.GatewayId)
			}
		}
	}

	// 用regionIds数组从RuleAsset中查出数据
	ruleAssetRepo := mongopo.NewRuleAssetRepoInstance()
	entity := model.RuleAssetQueryEntity{
		AssetIds: regionIds,
	}
	ruleAssetRegionList, err := ruleAssetRepo.QueryRuleAssets(entity)
	if err != nil {
		if !errors.Is(err, mongodrv.ErrNoDocuments) {
			logger.Logger.Errorf("Update accessible relation task fail, failed to query regions from mongo! err: %s", err.Error())
			return
		}
	}
	// ruleAssetMap 从RuleAsset中查出的数据，进行与uniqueRegions比较时只需要用到资产名称,这个数据不重复
	ruleAssetRegionMap := make(map[string]table.RuleAsset)
	uniqueCloudAccountIds := make(map[string]struct{})
	var cloudAccountIds []string
	for _, ruleAssetRegion := range ruleAssetRegionList {
		ruleAssetRegionMap[ruleAssetRegion.AssetId] = ruleAssetRegion
		if _, exists := uniqueCloudAccountIds[ruleAssetRegion.CloudAccountId]; !exists {
			uniqueCloudAccountIds[ruleAssetRegion.CloudAccountId] = struct{}{}
			cloudAccountIds = append(cloudAccountIds, ruleAssetRegion.CloudAccountId)
		}
	}

	// 用cloudAccountIds去t_cloud_account查出云环境名称
	cloudAccountList, err := repo.SelectPlatformNameByAccountId(cloudAccountIds)
	if err != nil {
		logger.Logger.Errorf("Update accessible relation task fail, failed to query cloud account from pg! err: %s", err.Error())
		return
	}
	// cloudAccountMap 不重复的云环境名称
	cloudAccountMap := make(map[string]string)
	for _, cloudAccount := range cloudAccountList {
		cloudAccountMap[cloudAccount.Id] = cloudAccount.PlatformName
	}

	// 用gatewayIds数组从RuleAsset中查出数据
	entity = model.RuleAssetQueryEntity{
		AssetIds: gatewayIds,
	}
	ruleAssetGatewayList, err := ruleAssetRepo.QueryRuleAssets(entity)
	if err != nil {
		if !errors.Is(err, mongodrv.ErrNoDocuments) {
			logger.Logger.Errorf("Update accessible relation task fail, failed to query gateway from mongo! err: %s", err.Error())
			return
		}
	}
	// ruleAssetGatewayMap 云上网关资产数据
	ruleAssetGatewayMap := make(map[string]table.RuleAsset)
	for _, ruleAssetGateway := range ruleAssetGatewayList {
		ruleAssetGatewayMap[ruleAssetGateway.AssetId] = ruleAssetGateway
	}

	// 循环regionsList
	for _, region := range regionsList {
		updateValue := make(map[string]interface{})
		if region.GatewaySource == constants.CLOUD_DEVICE {
			// 判断GatewayId是否在ruleAssetGatewayMap中是否存在
			if gatewayInMongo, exists := ruleAssetGatewayMap[region.GatewayId]; !exists {
				// 不存在，说明网关设备已被删除
				updateValue["gateway_delete"] = constants.DELETED
			} else {
				updateValue["gateway_delete"] = constants.UNDELETED
				// 存在，网关设备没删除
				updateValue["gateway_name"] = gatewayInMongo.AssetName
			}
		}

		// 判断regionId是否在ruleAssetRegionMap中是否存在
		if regionInMongo, exists := ruleAssetRegionMap[region.RegionId]; !exists {
			// 不存在，说明区域已被删除
			updateValue["region_delete"] = constants.DELETED
		} else {
			// 存在，区域没删除
			updateValue["region_delete"] = constants.UNDELETED

			// 判断云环境名称是否被修改
			if platformName, cloudAccountExists := cloudAccountMap[regionInMongo.CloudAccountId]; !cloudAccountExists {
				updateValue["region_delete"] = constants.DELETED
			} else {
				updateValue["region_name"] = platformName + "-" + regionInMongo.AssetName
				updateValue["cloud_name"] = platformName
			}
		}

		if len(updateValue) > 0 {
			updateErr := repo.UpdateAccessRelationshipById(region.Id, updateValue)
			if updateErr != nil {
				logger.Logger.Errorf("Update accessible relation task fail, failed to update accessible relation from pg! err: %s", updateErr.Error())
				continue
			}
		}
	}
}