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