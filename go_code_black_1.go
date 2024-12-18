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
func checkNetCard(ip string) bool {
	interfaceName, err := getInterfaceName(ip)
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err.Error())
		return false
	}

	speed, err := getInterfaceSpeed(interfaceName)
	if err != nil {
		log.Fatal(err)
	}
	if speed == "" {
		fmt.Println("[WARNING]Unable to detect network card speed")
		return true
	}
	if speed <= "100000" {
		fmt.Println("netcard is lower than 100Mb/s")
		return false
	}
	fmt.Printf("Interface: %s\nSpeed: %s\n", interfaceName, speed)
	return true
}