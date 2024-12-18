// subnetStatistic 同一个云账号下的子网风险统计
func subnetStatistic(ctx context.Context, cloudAccountId string, subnets []*asset.SubnetEntityV2,
	subnetAssetCount, subnetRiskAssetCount map[string]map[int][]string) errors2.RestV1Error {
	claim, err := utils.GetClaimFromContext(ctx)
	if err != nil {
		assetLogger.Logger.Errorf("cannot get claim from context!!err msg:%s, clain:%v", err.Error(), claim)
		return errors2.NewI18nError(i18n.GetAdminInfoFailed)
	}
	if claim.AdminId == 0 {
		assetLogger.Logger.Errorf("cannot get claim from context!! clain:%v", claim)
		return errors2.NewI18nError(i18n.TokenInvalid)
	}
	if subnetAssetCount == nil || subnetRiskAssetCount == nil {
		return errors2.NewI18nError(i18n.ReadAssetInfoFailed)
	}
	allSubnetId := make([]string, len(subnets)) // 所有的子网id
	for i, subnet := range subnets {
		allSubnetId[i] = subnet.AssetId
	}
	if assetIdShowsMap, err := getSubnetsIdShows(ctx, allSubnetId); err != nil { // map["资产原始id"]{"资产类型", "子网id"}
		return errors2.NewI18nError(i18n.ReadAssetInfoFailed)
	} else {
		var allAssetId []string // 子网下的所有原始资产id
		for k, v := range assetIdShowsMap {
			allAssetId = append(allAssetId, k)
			// 记录子网下的各类别的资产id
			if _, ok := subnetAssetCount[v.SubnetRelationId]; ok == false {
				subnetAssetCount[v.SubnetRelationId] = make(map[int][]string)
			}
			subnetAssetCount[v.SubnetRelationId][v.AssetType] = append(subnetAssetCount[v.SubnetRelationId][v.AssetType], k)
		}

		if len(allAssetId) == 0 {
			return nil
		}
		if risk, err := getAllTypeRiskAssetIdList(claim.AdminId, cloudAccountId, allAssetId); err != nil { // [{"资产原始id", "资产类型"}]
			return errors2.NewI18nError(i18n.ReadAssetInfoFailed)
		} else {
			riskAsset, err := GetOriginAsset(risk)
			if err != nil {
				assetLogger.Logger.Errorf("failed to query! err:%s", err.Error())
				return errors2.NewI18nError(i18n.ReadAssetInfoFailed)
			}
			for _, v := range riskAsset {
				assetId, assetType := v.AssetId, v.AssetType
				subnetId := assetIdShowsMap[assetId].SubnetRelationId
				if _, ok := subnetRiskAssetCount[subnetId]; ok == false {
					subnetRiskAssetCount[subnetId] = make(map[int][]string)
				}
				// 记录子网下的各类别的风险资产id
				subnetRiskAssetCount[subnetId][assetType] = append(subnetRiskAssetCount[subnetId][assetType], assetId)
			}
		}
	}
	return nil
}