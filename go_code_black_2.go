import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"upgrademan/pkg/service/scale"

	"code.sangfor.org/imp/xaas-os-platform/xos-common-libs/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"upgrademan/db"
	conf "upgrademan/pkg/config"
	"upgrademan/pkg/helm"
	"upgrademan/pkg/install"
	"upgrademan/pkg/install/config"
	"upgrademan/pkg/task"
)
// 获取默认值
func getDefaultInt(v int, defaultValue int) int {
	if v == 0 {
		v = defaultValue
	}
	return defaultValue
}