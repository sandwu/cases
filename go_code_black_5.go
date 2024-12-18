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