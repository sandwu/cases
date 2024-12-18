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