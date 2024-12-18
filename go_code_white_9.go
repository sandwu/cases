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