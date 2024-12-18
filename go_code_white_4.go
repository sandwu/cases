import (
	"fmt"
	"k8s.io/klog/v2"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"xaasos/config"

	"github.com/spf13/cobra"
	"xaasos/pkg/util"
)
// 检测网卡速率
func getInterfaceSpeed(interfaceName string) (string, error) {
	cmd := exec.Command("ethtool", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Speed:") {
			speed := strings.TrimSpace(strings.Split(line, ":")[1])
			var speed_ string
			var mul int
			if strings.Contains(speed, "Kb") {
				speed_ = strings.TrimSpace(strings.Split(speed, "Kb")[0])
				mul = 1
			} else if strings.Contains(speed, "Mb") {
				speed_ = strings.TrimSpace(strings.Split(speed, "Mb")[0])
				mul = 1000
			} else if strings.Contains(speed, "Gb") {
				speed_ = strings.TrimSpace(strings.Split(speed, "Gb")[0])
				mul = 1000000
			} else {
				return "", nil
			}
			fmt.Printf("speed_ = %s\n", speed_)
			num, err := strconv.Atoi(speed_)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%d", num*mul), nil
		}
	}
	// 没有找到Speed字段，无法判断网卡速率，故跳过
	return "", nil
}