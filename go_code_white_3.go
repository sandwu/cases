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