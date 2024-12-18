import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/klog/v2"
	"net/http"
	"sync"
	"time"
)
// @param task 任务// @param ctx 上下文// @description 上报任务状态
func reportStatus(ctx TypeHandlerContext, task *ApplicationTask, wg *sync.WaitGroup) {
	defer wg.Done()

	// 首次单独上报，如果出错可以忽略
	response, err := updateToCluster(ctx.Context, task)
	if err != nil {
		klog.Errorf("task %+v first report failed %+v -> %v", task, err, response)
		task.ReportRuntime = fmt.Sprintf("first report failed %v", err)
		update(task)
	}

	// 后续定时上报
	ticker := time.NewTicker(time.Duration(task.ReportInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Context.Done():
			// 超时了
			klog.Infof("timeout, task %+v report cancle", task)
			return
		case <-ticker.C:
			_, err := updateToCluster(ctx.Context, task)
			// 上报不成功一直重试（直接整个任务超时了）
			if err != nil {
				klog.Errorf("task %+v err:%v report failed and retry", task, err)
				task.ReportRuntime = fmt.Sprintf("report failed and retry %v", err)
				update(task)
				continue
			}

			// 任务结束且上报成功可以停止
			if task.Status == Reporting {
				klog.Infof("task %+v report success", task)
				task.ReportRuntime = fmt.Sprintf("report success")
				task.Status = Completed

				typeHandler, ok := GetTaskHandler(task.Type)
				if ok {
					klog.Infof("task %+v call HookComplete after reported", task)
					err := typeHandler.HookComplete(ctx)
					if err != nil {
						formatRuntime(ctx, nil, fmt.Sprintf("HookComplete err: %v", err.Error()), task)
					}
				}

				update(task)

				return
			}
		}

	}
}