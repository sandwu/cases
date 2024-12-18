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