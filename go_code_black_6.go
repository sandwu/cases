import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
	"sort"
	"strings"
	"time"
	"upgrademan/pkg/helm"
	"upgrademan/pkg/install"
	"upgrademan/pkg/install/config"
	"upgrademan/pkg/task"
)
func (h *TypeHandlerInstall) Handle(ctx task.TypeHandlerContext, req *helm.DeployRequest) error {
	appInfo, ok := req.Extend.(config.AppInfo)
	if !ok {
		return fmt.Errorf("get app info err: %v", req.Extend)
	}

	var hookExecRecord map[int]bool
	v, ok := ctx.Extra["hook_exec_record"]
	if ok {
		hookExecRecord = v.(map[int]bool)
	} else {
		hookExecRecord = make(map[int]bool)
		ctx.Extra["hook_exec_record"] = hookExecRecord
	}

	// 执行脚本, 获取这个级别的应用的前置配置
	if ok := hookExecRecord[appInfo.Priority]; ok {
		// 应用初始化脚本
		config.ArgConfigInit()
		config.GetPreArgs(appInfo.Priority, INSTALL)

		// 应用初始脚本
		config.AppDeployInit(req.ChartName, req.Namespace)

		// 获取应用初始化脚本参数
		config.GetAppPreArgs()
	}

	// 合并命令行传参和声明参数
	arg := appInfo.Args
	cmdArg, ok := config.AppArgMap[strings.ToLower(req.AppType)+"."+appInfo.Name]
	if ok {
		arg = append(arg, cmdArg)
	}
	klog.V(5).Info(appInfo.Name, " ", arg)
	req.Values = install.ArgsToYAML(arg)
	req.Wait = true
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	// 部署helm应用
	helmClient := helm.NewClient()
	err := helmClient.InstallChart(req)
	if err != nil {
		klog.Errorf("intall failed %v", err)
		return err
	}

	//  执行脚本, 获取这个级别的应用的后置配置或操作
	if ok := hookExecRecord[appInfo.Priority]; ok {
		config.PostHookShell(appInfo.Priority, INSTALL)
	}
	hookExecRecord[appInfo.Priority] = true
	return nil
}