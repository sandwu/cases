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