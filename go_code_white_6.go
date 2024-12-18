import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"os"
	"time"

	"upgrademan/pkg/model"
	"upgrademan/pkg/tasksys"
	util "upgrademan/pkg/utils"
	"upgrademan/pkg/utils/httpcommon"
)
func GetCurrentUpgradeStatus() (*CurrentStaus, time.Time) {
	cStatus := &CurrentStaus{
		status: model.StatusUpgradeSuccess,
		utype:  0,
		reason: "",
	}

	// 有正在执行的系统升级任务
	if CheckSysRunningTasks() {
		cStatus.status = model.StatusUpgrading
		cStatus.utype = 1
		return cStatus, time.Time{}
	}

	// 检查最后一次系统升级任务的状态
	task, err := GetLatestSysTask()
	if err != nil {
		klog.Warning(err)
		return cStatus, time.Time{}
	}

	upgradeResult, err := task.GetStatus("")
	if err != nil {
		klog.Error(err)
		return cStatus, time.Time{}
	}

	failVersion, err := task.GetVersion()
	if err != nil {
		klog.Error(err)
	} else {
		cStatus.version = failVersion
	}

	endTime := time.Time{}
	klog.Infof("upgrade status is %v", upgradeResult)
	if upgradeResult != "Completed" {
		cStatus.status = model.StatusUpgradeFail

		if cStatus.status == model.StatusUpgradeFail {
			cStatus.utype = 0
		}
		reason, err := task.GetReason()
		if err != nil {
			klog.Error(err)
		} else {
			cStatus.reason = reason
		}

		endTime, err = task.GetEndTime()
		if err != nil {
			klog.Error(err)
		}
	}

	return cStatus, endTime
}