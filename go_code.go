import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/klog/v2"
	"net/http"
	"sync"
	"time"
)
// @param task 任务// @param ctx 上下文// @description 上报任务状态
func reportStatus(ctx TypeHandlerContext, task *ApplicationTask, wg *sync.WaitGroup) {
	defer wg.Done()

	// 首次单独上报，如果出错可以忽略
	response, err := updateToCluster(ctx.Context, task)
	if err != nil {
		klog.Errorf("task %+v first report failed %+v -> %v", task, err, response)
		task.ReportRuntime = fmt.Sprintf("first report failed %v", err)
		update(task)
	}

	// 后续定时上报
	ticker := time.NewTicker(time.Duration(task.ReportInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Context.Done():
			// 超时了
			klog.Infof("timeout, task %+v report cancle", task)
			return
		case <-ticker.C:
			_, err := updateToCluster(ctx.Context, task)
			// 上报不成功一直重试（直接整个任务超时了）
			if err != nil {
				klog.Errorf("task %+v err:%v report failed and retry", task, err)
				task.ReportRuntime = fmt.Sprintf("report failed and retry %v", err)
				update(task)
				continue
			}

			// 任务结束且上报成功可以停止
			if task.Status == Reporting {
				klog.Infof("task %+v report success", task)
				task.ReportRuntime = fmt.Sprintf("report success")
				task.Status = Completed

				typeHandler, ok := GetTaskHandler(task.Type)
				if ok {
					klog.Infof("task %+v call HookComplete after reported", task)
					err := typeHandler.HookComplete(ctx)
					if err != nil {
						formatRuntime(ctx, nil, fmt.Sprintf("HookComplete err: %v", err.Error()), task)
					}
				}

				update(task)

				return
			}
		}

	}
}

import (
	"fmt"
	"k8s.io/klog/v2"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"xaasos/config"

	"github.com/spf13/cobra"
	"xaasos/pkg/util"
)
func checkNetCard(ip string) bool {
	interfaceName, err := getInterfaceName(ip)
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err.Error())
		return false
	}

	speed, err := getInterfaceSpeed(interfaceName)
	if err != nil {
		log.Fatal(err)
	}
	if speed == "" {
		fmt.Println("[WARNING]Unable to detect network card speed")
		return true
	}
	if speed <= "100000" {
		fmt.Println("netcard is lower than 100Mb/s")
		return false
	}
	fmt.Printf("Interface: %s\nSpeed: %s\n", interfaceName, speed)
	return true
}

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"upgrademan/pkg/service/scale"

	"code.sangfor.org/imp/xaas-os-platform/xos-common-libs/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"upgrademan/db"
	conf "upgrademan/pkg/config"
	"upgrademan/pkg/helm"
	"upgrademan/pkg/install"
	"upgrademan/pkg/install/config"
	"upgrademan/pkg/task"
)
// 获取默认值
func getDefaultInt(v int, defaultValue int) int {
	if v == 0 {
		v = defaultValue
	}
	return defaultValue
}

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
func cloudAccountCache() (map[string]*manage.CloudAccountWithTags, error) {
	/* 补充云账户信息 */
	cloudAccounts, err := manage.FindAccountDetail()
	if err != nil {
		assetLogger.Logger.Errorf("FindAccountDetail failed!!err msg:%s", err.Error())
		return nil, err
	}
	// find cloudAccount info by ids
	cloudAccountMap := make(map[string]*manage.CloudAccountWithTags)
	for _, v := range cloudAccounts {
		account := v
		var tagList []string
		rows, tagErr := manage.FindTagListByAccountId(account.ID)
		if tagErr != nil {
			assetLogger.Logger.Errorf("cannot find tag List by accountId:%s,err msg:%v", account.ID, tagErr)
			return nil, err
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				assetLogger.Logger.Errorf("close row failed,err: %v", err)
			}
		}(rows)
		for rows.Next() {
			var code int64
			var name string
			if scanErr := rows.Scan(&code, &name); err != nil {
				assetLogger.Logger.Errorf("scan tag data occur error, error msg:%s", err.Error())
				return nil, scanErr
			}
			tagList = append(tagList, name)
		}
		cloudAccountMap[account.ID] = &manage.CloudAccountWithTags{
			CloudAccountSimpleInfo: *account,
			Tags:                   tagList,
		}
	}
	return cloudAccountMap, nil
}

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

import (
	"code.sangfor.org/xaasos/clusterman/server/internal/types"
	"code.sangfor.org/xaasos/clusterman/server/pkg/sdk/db"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)
func GetXOSAllNodeIP(ctx context.Context) ([]string, error) {
	ClusterMgr := Cluster()
	nodeIPs := make([]string, 0)
	opt := types.XosClusterNodeOpt{
		Search:    "",
		PageSize:  100,
		PageIndex: 1,
	}
	data, err := ClusterMgr.GetClusterNodes(ctx, opt)
	if err != nil {
		return nodeIPs, err
	}
	indexs := data.Total / int64(opt.PageSize)

	for ; indexs >= 0; indexs-- {
		data, err := ClusterMgr.GetClusterNodes(ctx, opt)
		if err != nil {
			logx.Errorf("get ListNodes node failed: %s", err.Error())
			return nodeIPs, err
		}
		if len(data.Node) == 0 {
			break
		}
		for _, node := range data.Node {
			nodeIPs = append(nodeIPs, node.Ip)
		}
		opt.PageIndex++
	}
	return nodeIPs, nil
}

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
	"sort"
	"strings"
	"time"
	"upgrademan/pkg/helm"
	"upgrademan/pkg/install"
	"upgrademan/pkg/install/config"
	"upgrademan/pkg/task"
)
func (h *TypeHandlerInstall) Handle(ctx task.TypeHandlerContext, req *helm.DeployRequest) error {
	appInfo, ok := req.Extend.(config.AppInfo)
	if !ok {
		return fmt.Errorf("get app info err: %v", req.Extend)
	}

	var hookExecRecord map[int]bool
	v, ok := ctx.Extra["hook_exec_record"]
	if ok {
		hookExecRecord = v.(map[int]bool)
	} else {
		hookExecRecord = make(map[int]bool)
		ctx.Extra["hook_exec_record"] = hookExecRecord
	}

	// 执行脚本, 获取这个级别的应用的前置配置
	if ok := hookExecRecord[appInfo.Priority]; ok {
		// 应用初始化脚本
		config.ArgConfigInit()
		config.GetPreArgs(appInfo.Priority, INSTALL)

		// 应用初始脚本
		config.AppDeployInit(req.ChartName, req.Namespace)

		// 获取应用初始化脚本参数
		config.GetAppPreArgs()
	}

	// 合并命令行传参和声明参数
	arg := appInfo.Args
	cmdArg, ok := config.AppArgMap[strings.ToLower(req.AppType)+"."+appInfo.Name]
	if ok {
		arg = append(arg, cmdArg)
	}
	klog.V(5).Info(appInfo.Name, " ", arg)
	req.Values = install.ArgsToYAML(arg)
	req.Wait = true
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	// 部署helm应用
	helmClient := helm.NewClient()
	err := helmClient.InstallChart(req)
	if err != nil {
		klog.Errorf("intall failed %v", err)
		return err
	}

	//  执行脚本, 获取这个级别的应用的后置配置或操作
	if ok := hookExecRecord[appInfo.Priority]; ok {
		config.PostHookShell(appInfo.Priority, INSTALL)
	}
	hookExecRecord[appInfo.Priority] = true
	return nil
}

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
)
func (vpn VpnGateway) SetNodeInfo(node CommonAsset) Node {
	vpn.CommonAsset = node
	var err error
	var byteObj []byte
	byteObj, err = bson.Marshal(node.Data)
	if err != nil {
		panic("fail to create vpn node: " + err.Error())
	}
	err = bson.Unmarshal(byteObj, vpn.Data)
	if err != nil {
		panic("fail to create vpn node: " + err.Error())
	}
	return vpn
}

import (
	"sangfor.com/csc/csc/internal/app/cscasset/module/common"
	commonConstant "sangfor.com/csc/csc/internal/app/csctopo/module/constant"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/chain"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/constant"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/params"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/pool"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/util"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/vertex"
	"strconv"
)
func generateCenOrGwEdge(sceneParams *params.SceneParams, assetNodeChain chain.AssetNodeChain) {
	// 源目节点
	startNode := assetNodeChain.NodePath[0]
	endNode := assetNodeChain.NodePath[len(assetNodeChain.NodePath)-1]

	// 离起始、目的节点最近的云企业网或网关
	startNodeCenOrGw := assetNodeChain.NodePath[4]
	endNodeCenOrGw := assetNodeChain.NodePath[len(assetNodeChain.NodePath)-5]

	// 如存在可访问关系,预生成图数据库的边
	// asset1 -> asset2: accessible
	sceneParams.ResultsChan <- map[string]string{commonConstant.EdgeTypeAccessible: util.GenerateNebulaEdge(startNode.GetId(), endNode.GetId())}
	// vpc1 -> vpc2: accessible
	sceneParams.ResultsChan <- map[string]string{commonConstant.EdgeTypeAccessible: util.GenerateNebulaEdge(startNode.GetVpcId(), endNode.GetVpcId())}
	// 云账号1 -> 云账号2: accessible
	if startNode.GetCloudAccountId() != endNode.GetCloudAccountId() {
		sceneParams.ResultsChan <- map[string]string{commonConstant.EdgeTypeAccessible: util.GenerateNebulaEdge(startNode.GetCloudAccountId(), endNode.GetCloudAccountId())}
	}
	// 云平台1 -> 云平台2: accessible
	if startNode.GetCloudSource() != endNode.GetCloudSource() {
		sceneParams.ResultsChan <- map[string]string{commonConstant.EdgeTypeAccessible: util.GenerateNebulaEdge(strconv.FormatInt(startNode.GetCloudSource(), 10), strconv.FormatInt(endNode.GetCloudSource(), 10))}
	}

	// asset1 -> 离asset1最近的云企业网或网关: forward
	sceneParams.ResultsChan <- map[string]string{commonConstant.EdgeTypeForward: util.GenerateNebulaEdge(startNode.GetId(), startNodeCenOrGw.GetId())}
	// 离asset2最近的云企业网或网关 -> asset2: forward
	sceneParams.ResultsChan <- map[string]string{commonConstant.EdgeTypeForward: util.GenerateNebulaEdge(endNodeCenOrGw.GetId(), endNode.GetId())}
	// 离asset1最近的云企业网或网关 -> ... -> 离asset2最近的云企业网或网关: forward
	for i := 4; i < len(assetNodeChain.NodePath)-5; i++ {
		sceneParams.ResultsChan <- map[string]string{commonConstant.EdgeTypeForward: util.GenerateNebulaEdge(assetNodeChain.NodePath[i].GetId(), assetNodeChain.NodePath[i+1].GetId())}
	}
}

import (
	"sangfor.com/csc/csc/internal/app/cscasset/module/common"
	commonConstant "sangfor.com/csc/csc/internal/app/csctopo/module/constant"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/chain"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/constant"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/params"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/pool"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/util"
	"sangfor.com/csc/csc/internal/app/csctopo/module/rule_calculator/vertex"
	"strconv"
)
func processCenOrGwChainAndGenerateEdge(assetNodeChain chain.AssetNodeChain, sceneParams *params.SceneParams) {
	lenPath := len(assetNodeChain.NodePath)

	// 由于源目节点需要在计算时修改ip，为避免并发改，这里将源目节点重新实例化
	endpoint1, ok := sceneParams.Storage.GetAssetByID(assetNodeChain.NodePath[0].GetId())
	if !ok {
		return
	}
	endpoint2, ok := sceneParams.Storage.GetAssetByID(assetNodeChain.NodePath[lenPath-1].GetId())
	if !ok {
		return
	}
	assetNodeChain.NodePath[0] = assetNodeFactory.CreateNode(endpoint1)
	assetNodeChain.NodePath[lenPath-1] = assetNodeFactory.CreateNode(endpoint2)

	// 源目节点
	startNode := assetNodeChain.NodePath[0]
	endNode := assetNodeChain.NodePath[lenPath-1]

	// 通过网卡上的ip进行遍历
	startEthNode := assetNodeChain.NodePath[1]
	endEthNode := assetNodeChain.NodePath[lenPath-2]

	startNodePrivateIp := startEthNode.GetPrivateIp()
	endNodePrivateIp := endEthNode.GetPrivateIp()

	for _, startIp := range startNodePrivateIp {
		startNode.SetCurIp(startIp)
		for _, endIp := range endNodePrivateIp {
			endNode.SetCurIp(endIp)
			accessible := util.CalculateAccessibleRelationship(assetNodeChain)
			if accessible {
				generateCenOrGwEdge(sceneParams, assetNodeChain)
				return
			}
		}
	}
	return
}

import (
	"context"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"go.mongodb.org/mongo-driver/bson"
	"sangfor.com/csc/csc/internal/app/cscasset/module/asset"
	assetpb "sangfor.com/csc/csc/internal/app/cscasset/module/asset/pb"
	"sangfor.com/csc/csc/internal/app/cscasset/module/common"
	assetapi "sangfor.com/csc/csc/internal/app/cscasset/openapi"
	managepb "sangfor.com/csc/csc/internal/app/cscmanage/module/access/pb"
	"sangfor.com/csc/csc/internal/app/cscmanage/openapi"
	common2 "sangfor.com/csc/csc/internal/app/csctopo/module/common"
	"sangfor.com/csc/csc/internal/pkg/constants"
	"sangfor.com/csc/csc/internal/pkg/errors2"
	"sangfor.com/csc/csc/internal/pkg/i18n"
	"sangfor.com/csc/csc/internal/pkg/utils/collections"
	"sangfor.com/csc/csc/internal/pkg/utils/strutil"
	"sort"
	"strconv"
)
// directionOut为true时查询出子网的方向// 根据访问方向查询子网拓扑
func querySubnetTopologyByDirection(ctx context.Context, vpcAssetId string, directionOut bool) ([]*Edge,
	map[string]*GetSubnetTopologyResponseDataNode, errors2.RestV1Error) {
	internetNgql := `MATCH (vpc:vpc)-[:link]->(subnet:subnet)-[:relate]->(e:endpoint)-[:forward]->(t1:transfer)
-[:forward*0..1]->(t2:transfer)-[:forward*0..1]->(internet:internet)
  WHERE id(vpc) == $vpcId  return DISTINCT subnet, t1, t2, internet`
	// 跨vpc通用查询语句，包括对等连接，VPN网关，专线网关，云企业网（同云充当对等连接的场景）
	// todo: 530不考虑云企业网和网关相连
	vpcPeerNgql := `MATCH (vpc:vpc)-[:link]->(subnet:subnet)-[:relate]->(:endpoint)-[:forward]->(t1:transfer)-[:forward*0..3]-(t2:transfer)
  WHERE id(vpc) == $vpcId AND any(t IN ["peer","vpn_gateway","ecr","cen"] WHERE t IN LABELS(t1))
MATCH (vpc2:vpc)-[:link]->(:subnet)-[:relate]->(:endpoint)<-[:forward]-(t2)
  WHERE id(vpc)!=id(vpc2) AND (vpc:vpc)-[:accessible]-(vpc2:vpc)
return distinct subnet, t1, vpc2`

	if !directionOut {
		internetNgql = `MATCH (vpc:vpc)-[:link]->(subnet:subnet)-[:relate]->(e:endpoint)<-[:forward]-(t1:transfer)
<-[:forward*0..1]-(t2:transfer)<-[:forward*0..1]-(internet:internet)
  WHERE id(vpc) == $vpcId  return distinct subnet, t1, t2, internet`
		// 跨vpc通用查询语句，包括对等连接，VPN网关，专线网关，云企业网（同云充当对等连接的场景）
		vpcPeerNgql = `MATCH (vpc:vpc)-[:link]->(subnet:subnet)-[:relate]->(:endpoint)<-[:forward]-(t1:transfer)-[:forward*0..3]-(t2:transfer)
  WHERE id(vpc) == $vpcId AND any(t IN ["peer","vpn_gateway","ecr","cen"] WHERE t IN LABELS(t1))
MATCH (vpc2:vpc)-[:link]->(:subnet)-[:relate]->(:endpoint)-[:forward]->(t2) 
  WHERE id(vpc)!=id(vpc2) AND (vpc:vpc)-[:accessible]-(vpc2:vpc)
return distinct subnet, t1, vpc2`
	}

	// 互联网与子网之间拓扑
	edges1, nodeMap1, err := executeSubnetNgql(ctx, vpcAssetId, directionOut, internetNgql)
	if err != nil {
		return nil, nil, err
	}
	// 其他vpc与子网之间拓扑
	edges2, nodeMap2, err := executeSubnetNgql(ctx, vpcAssetId, directionOut, vpcPeerNgql)
	if err != nil {
		return nil, nil, err
	}
	edges, nodeMap := _mergeNodeEdge(nodeMap1, edges1, nodeMap2, edges2)
	return edges, nodeMap, nil
}

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

func GenerateChart_RedisSingle_BaseResource() {

	// TODO add by zlr, 需要修改， 需要国际化
	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"cpu",
		"CPU使用率", "节点的CPU使用率", "%", []string{
			`avg(rate(container_cpu_usage_seconds_total{job="cadvisor", namespace="$namespace", container="redis"}[1m])*100)`,
		},
		[]string{
			"CPU使用率",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"memory",
		"内存使用量", "节点的内存使用量", "MB", []string{
			`sum(container_memory_usage_bytes{job="cadvisor", namespace="$namespace", container="redis"}/(1024*1024))`,
		},
		[]string{
			"内存使用量",
		})
	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"memory",
		"内存使用率", "redis中间件的内存使用率待修改", "%", []string{
			`100*avg(container_memory_usage_bytes{job="cadvisor", namespace="$namespace", container="redis"}/container_spec_memory_limit_bytes{job="cadvisor", namespace="$namespace", container="redis"})`,
		},
		[]string{
			"内存使用率",
		})
	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘写入IOPS", "节点的数据盘写入IOPS", "times/s", []string{
			`sum(rate(container_fs_writes_total{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}[1m]))`,
		},
		[]string{
			"数据盘写入IOPS",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘读取IOPS", "节点的数据盘读取IOPS", "times/s", []string{
			`sum(rate(container_fs_reads_total{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}[1m]))`,
		},
		[]string{
			"数据盘读取IOPS",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘写入吞吐量", "节点的数据盘写入吞吐量", "MB/s", []string{
			`sum(rate(container_fs_writes_bytes_total{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}/(1024*1024)[1m]))`,
		},
		[]string{
			"数据盘写入吞吐量",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘读取吞吐量", "节点的数据盘读取吞吐量", "MB/s", []string{
			`sum(rate(container_fs_reads_bytes_total{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}/(1024*1024)[1m]))`,
		},
		[]string{
			"数据盘读取吞吐量",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘使用量", "节点的数据盘使用量", "GB", []string{
			`sum(container_fs_usage_bytes{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}/(1024*1024*1024))`,
		},
		[]string{
			"数据盘使用量",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘使用率", "节点的数据盘使用率", "%", []string{
			`avg(100*container_fs_usage_bytes{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}/container_fs_limit_bytes{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"})`,
		},
		[]string{
			"数据盘使用率",
		})
}

//	@return error//	@return vertex.Node//	@param dstId 目的节点id//	@param srcId 源节点id//	@Description: 查询源目节点信息//// getSrcAndDstNode
func getSrcAndDstNode(srcId, dstId string) (vertex.Node, vertex.Node, error) {
	// 目的节点不能为互联网节点
	if dstId == constants.InternetNodeId {
		return nil, nil, constant.DestNodeTypeError
	}

	var srcNode, dstNode vertex.Node
	var err error
	if srcId == constants.InternetNodeId {
		srcNode = vertex.Internet{}
	} else {
		srcNode, err = nodeId2Node(srcId)
		if err != nil {
			attackLogger.Logger.Error(fmt.Sprintf("getSrcAndDstNode, get srcnode detail failed, err is %v, srcId is %v", err, srcId))
			return nil, nil, err
		}
		// 检查是否是db或ecs
		if int(srcNode.GetType()) != common.AssetTypeCloudServer && !tool.InIntSlice(constant.DbAssetType, int(srcNode.GetType())) {
			return nil, nil, SrcOrDstNodeTypeError
		}
	}

	dstNode, err = nodeId2Node(dstId)
	if err != nil {
		attackLogger.Logger.Error(fmt.Sprintf("getSrcAndDstNode, get dstnode detail failed, err is %v, dstId is %v", err, dstId))
		return nil, nil, err
	}
	// 检查是否是db或ecs
	if int(dstNode.GetType()) != common.AssetTypeCloudServer && !tool.InIntSlice(constant.DbAssetType, int(dstNode.GetType())) {
		return nil, nil, SrcOrDstNodeTypeError
	}

	return srcNode, dstNode, nil
}

import (
	"fmt"
	"k8s.io/klog/v2"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"xaasos/config"

	"github.com/spf13/cobra"
	"xaasos/pkg/util"
)
// 检测网卡速率
func getInterfaceSpeed(interfaceName string) (string, error) {
	cmd := exec.Command("ethtool", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Speed:") {
			speed := strings.TrimSpace(strings.Split(line, ":")[1])
			var speed_ string
			var mul int
			if strings.Contains(speed, "Kb") {
				speed_ = strings.TrimSpace(strings.Split(speed, "Kb")[0])
				mul = 1
			} else if strings.Contains(speed, "Mb") {
				speed_ = strings.TrimSpace(strings.Split(speed, "Mb")[0])
				mul = 1000
			} else if strings.Contains(speed, "Gb") {
				speed_ = strings.TrimSpace(strings.Split(speed, "Gb")[0])
				mul = 1000000
			} else {
				return "", nil
			}
			fmt.Printf("speed_ = %s\n", speed_)
			num, err := strconv.Atoi(speed_)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%d", num*mul), nil
		}
	}
	// 没有找到Speed字段，无法判断网卡速率，故跳过
	return "", nil
}

//	@return []int64 可利用端口//	@return bool 连通性结果//	@param xdrPorts 待检查端口//	@param nodeIdOfPath 路径节点id//	@param dstNode 目的节点//	@param srcNode 源节点//	@Description: 确认一组源目节点的ip后,检查该组ip在路径节点上的端口连通性//// checkOneIpPath
func checkOneIpPath(srcNode, dstNode vertex.Node, nodeIdOfPath []string, xdrPorts []int64) (bool, []int64) {
	for i := 0; i < len(nodeIdOfPath); i++ {
		curNode, err := nodeId2Node(nodeIdOfPath[i])
		if err != nil {
			logger.Logger.Errorf("exec nodeId2Node fail, err is: %v ", err)
			return false, []int64{}
		}

		if tool.InIntSlice(constant.SupportPortDenoiseAssetType, int(curNode.GetType())) {
			isAvailable, usePorts, err := curNode.ComputePort(srcNode, dstNode, nil, nil, xdrPorts)
			if err != nil || !isAvailable {
				logger.Logger.Errorf("exec computePort fail, srcNode: %v, dstNode: %v, err: %v ", srcNode, curNode, err)
				return false, []int64{}
			}
			xdrPorts = usePorts
		}
	}
	return true, xdrPorts
}

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"os"
	"time"

	"upgrademan/pkg/model"
	"upgrademan/pkg/tasksys"
	util "upgrademan/pkg/utils"
	"upgrademan/pkg/utils/httpcommon"
)
func GetCurrentUpgradeStatus() (*CurrentStaus, time.Time) {
	cStatus := &CurrentStaus{
		status: model.StatusUpgradeSuccess,
		utype:  0,
		reason: "",
	}

	// 有正在执行的系统升级任务
	if CheckSysRunningTasks() {
		cStatus.status = model.StatusUpgrading
		cStatus.utype = 1
		return cStatus, time.Time{}
	}

	// 检查最后一次系统升级任务的状态
	task, err := GetLatestSysTask()
	if err != nil {
		klog.Warning(err)
		return cStatus, time.Time{}
	}

	upgradeResult, err := task.GetStatus("")
	if err != nil {
		klog.Error(err)
		return cStatus, time.Time{}
	}

	failVersion, err := task.GetVersion()
	if err != nil {
		klog.Error(err)
	} else {
		cStatus.version = failVersion
	}

	endTime := time.Time{}
	klog.Infof("upgrade status is %v", upgradeResult)
	if upgradeResult != "Completed" {
		cStatus.status = model.StatusUpgradeFail

		if cStatus.status == model.StatusUpgradeFail {
			cStatus.utype = 0
		}
		reason, err := task.GetReason()
		if err != nil {
			klog.Error(err)
		} else {
			cStatus.reason = reason
		}

		endTime, err = task.GetEndTime()
		if err != nil {
			klog.Error(err)
		}
	}

	return cStatus, endTime
}

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	qm_options "github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gorm.io/gorm"
	"sangfor.com/csc/csc/internal/app/cscasset/module/common"
	"sangfor.com/csc/csc/internal/app/cscasset/module/internal/model/asset"
	progress2 "sangfor.com/csc/csc/internal/app/cscasset/module/internal/progress"
	ctYunAuth "sangfor.com/csc/csc/internal/app/cscasset/module/internal/sdk/ctyun-sdk-web/auth"
	cuCloudAuth "sangfor.com/csc/csc/internal/app/cscasset/module/internal/sdk/cucloud-sdk-web/auth"
	manageClient "sangfor.com/csc/csc/internal/app/cscasset/module/manage"
	"sangfor.com/csc/csc/internal/app/cscasset/module/manage/ctyun"
	"sangfor.com/csc/csc/internal/app/cscasset/module/manage/cucloud"
	pb2 "sangfor.com/csc/csc/internal/app/cscasset/module/manage/pb"
	"sangfor.com/csc/csc/internal/app/cscasset/module/sync/pb"
	privateAli "sangfor.com/csc/csc/internal/app/cscasset/module/sync/privatecloud/ali"
	privateHuawei "sangfor.com/csc/csc/internal/app/cscasset/module/sync/privatecloud/huawei"
	"sangfor.com/csc/csc/internal/app/cscasset/module/sync/privatecloud/vmware"
	ali2 "sangfor.com/csc/csc/internal/app/cscasset/module/sync/publiccloud/ali"
	"sangfor.com/csc/csc/internal/app/cscasset/module/sync/publiccloud/huawei"
	tencent2 "sangfor.com/csc/csc/internal/app/cscasset/module/sync/publiccloud/tencent"
	"sangfor.com/csc/csc/internal/app/cscasset/module/sync/task"
	taskCommon "sangfor.com/csc/csc/internal/app/cscasset/module/sync/task/common"
	"sangfor.com/csc/csc/internal/app/cscdatalink/db/akskdb/models"
	asset2 "sangfor.com/csc/csc/internal/app/cscdatalink/sync/aes"
	managecommon "sangfor.com/csc/csc/internal/app/cscmanage/common"
	"sangfor.com/csc/csc/internal/app/cscmanage/module/permission"
	"sangfor.com/csc/csc/internal/pkg/logger"
	mongo "sangfor.com/csc/csc/internal/pkg/mymongo"
	db "sangfor.com/csc/csc/internal/pkg/mypostgresql"
	"sangfor.com/csc/csc/internal/pkg/myredis"
	security "sangfor.com/csc/csc/internal/pkg/security"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)
func connectivityTestByCloudAccounts(accountInfos []*AccountInfo) []*task.CloudAccount {
	var res []*task.CloudAccount
	for _, v := range accountInfos {
		tenantId, err := strconv.Atoi(v.AdminId)
		if err != nil {
			Log.Errorf("string transfer int failed, err msg: %v, admin: %s", err, v.AdminId)
			continue
		}
		certificateJson, err := json.Marshal(v.AccessKey)
		if err != nil {
			Log.Errorf("cannot covert certificate struct to json,err msg: %v", err)
			continue
		}
		if v.Type == 1 {
			res = append(res, &task.CloudAccount{
				Id:           v.CloudAccountId,
				Adminid:      int64(tenantId),
				Creator:      v.Creator,
				Version:      common.CloudSourceStrMap[v.CloudSource],
				Certificate:  string(certificateJson),
				Address:      v.Address,
				CloudVersion: v.CloudVersion,
			})
		} else if v.Type == 2 {
			//如果是测试连通性不通过就不发送任务，否则会出现大量错误日志
			temp := v.AccessKey.AccessKeySecret
			if err := decryptAk(v.AccessKey.AccessKeyId,
				&temp); err != nil {
				Log.Errorf("DecryptAk, err: %s", err.Error())
				continue
			}
			status := PrivateCloudConnect(v.CloudSource, v.AccessKey.AccessKeyId,
				temp, v.Address)
			if status == common.ConnectNormal {
				res = append(res, &task.CloudAccount{
					Id:           v.CloudAccountId,
					Adminid:      int64(tenantId),
					Version:      common.CloudSourceStrMap[v.CloudSource],
					Certificate:  string(certificateJson),
					Address:      v.Address,
					CloudVersion: v.CloudVersion,
				})
			}
		}
	}
	return res
}

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
func SyncCloudRegionCron(ctx context.Context, info *cr.CommonRequest) (*cr.CommonResponse, error) {
	UpdateDeviceRelationSubTask()
	UpdateAccessibleRelationSubTask()
	return &cr.CommonResponse{}, nil
}