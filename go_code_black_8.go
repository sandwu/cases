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