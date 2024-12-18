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