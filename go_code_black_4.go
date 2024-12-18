import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	pulsarClient "github.com/apache/pulsar-client-go/pulsar"
	"github.com/hamba/avro"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	assetLogger "sangfor.com/csc/csc/internal/app/cscasset/logger"
	"sangfor.com/csc/csc/internal/app/cscasset/module/asset"
	"sangfor.com/csc/csc/internal/app/cscasset/module/common"
	"sangfor.com/csc/csc/internal/app/cscasset/module/manage"
	"sangfor.com/csc/csc/internal/app/cscdatalink/db"
	"sangfor.com/csc/csc/internal/app/cscxdata/resource/cloudresource"
	cr "sangfor.com/csc/csc/internal/pkg/common-request"
	"sangfor.com/csc/csc/internal/pkg/mypulsar"
	"strings"
	"time"
)
func fromDbToXDR(ctx context.Context, assetDataList *[]asset.AssetEntityXDR,
	accountMap map[string]*manage.CloudAccountWithTags) []asset.ForXDRAsset {
	var assetResponseList []asset.ForXDRAsset

	var assetIds []string
	var tmp asset.ForXDRAsset // 全覆盖赋值，提升初始化性能
	for _, assetData := range *assetDataList {
		tmp.UploadTimestamp = time.Now().Unix()
		tmp.OriginAssetId = assetData.AssetId
		tmp.AssetId = assetData.AssetIdShow
		tmp.Device.Name = assetData.AssetName
		tmp.Device.ClassifyId = common.XdrClassifyld
		tmp.Device.Classify1Id = common.XdrClassify1ld
		tmp.Device.OnlineStatus = 101

		// 依据网卡id补充ip和mac字段
		if assetData.Data.NetworkInterfaces != nil {
			relatedInterfaceIds := make([]string, len(assetData.Data.NetworkInterfaces))
			for index, ni := range assetData.Data.NetworkInterfaces {
				relatedInterfaceIds[index] = ni.RelationNetworkInterfaceId
			}
			// 查询ip 和 mac
			privateIpsResult, err := asset.FindIpAndMac(ctx, common.CollectionRuleAsset, relatedInterfaceIds)
			if err != nil {
				assetLogger.Logger.Errorf(fmt.Sprintf("find ip and mac  failed, err is: %v", err))
				continue
			} else {
				tmp.Ips = make([]asset.IpMac, 0)
				for index, _ := range *privateIpsResult {
					tmp.Ips = append(tmp.Ips, (*privateIpsResult)[index].PrivateIps...)
				}
				tmp.Cloud.InternetIps = make([]asset.IpMac, 0)
				for index, _ := range assetData.Data.IpInfo.PublicIp {
					tmp.Cloud.InternetIps = append(tmp.Cloud.InternetIps, asset.IpMac{Addr: assetData.Data.IpInfo.PublicIp[index]})
				}
			}
		} else {
			tmp.Ips = nil
			tmp.Cloud.InternetIps = nil
		}
		tmp.Os.Type = common.OsTypeMapXdr[common.Unknown]
		tmp.Adapter.SId = assetData.Data.Vpc.Id
		tmp.Adapter.Vendor = "-"
		tmp.Adapter.ProductVer = "-"
		tmp.Adapter.ProductType = common.Version2CloudSourceMap[assetData.CloudSource]
		tmp.Cloud.Type = int(assetData.CloudSource)
		tmp.Cloud.Tags = make([]string, 0)
		tmp.Cloud.Vpc.Id = assetData.Data.Vpc.Id
		tmp.Cloud.Vpc.Name = assetData.Data.Vpc.Name
		tmp.Cloud.Region.Id = assetData.Data.Region.Id
		tmp.Cloud.Region.Name = assetData.Data.Region.Name
		tmp.Cloud.Accounts = fillAccountInfo(assetData.CloudAccountId, accountMap)
		assetIds = append(assetIds, assetData.AssetId)
		assetResponseList = append(assetResponseList, tmp)

	}
	/* 从旧Asset表中补充osType CloudTags字段 */
	extraAssetResp, err := asset.FindAssetsByFilterForce(ctx, common.CollectionAsset, []bson.M{{"$match": bson.M{"assetId": bson.M{"$in": assetIds}}}})
	if err != nil {
		assetLogger.Logger.Errorf("FindAssetsByFilterForce failed!!err msg:%s", err.Error())
		return assetResponseList
	}
	extraAssetMap := make(map[string]*asset.AssetEntity)
	for _, v := range *extraAssetResp {
		extraAsset := v
		extraAssetMap[extraAsset.AssetId] = &extraAsset
	}

	for i, _ := range assetResponseList {
		data := &assetResponseList[i]
		if value, ok := extraAssetMap[data.OriginAssetId]; ok {
			data.Os.Type = common.OsTypeMapXdr[value.OsType]
			data.Cloud.Tags = value.CloudTag
			//data.Device.OnlineStatus = common.StatusMapXdr[value.Status]
		}
	}
	return assetResponseList
}