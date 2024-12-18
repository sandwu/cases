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