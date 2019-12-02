package minedev

import (
	"github.com/pocethereum/pochain/common"
	"github.com/pocethereum/pochain/consensus/poc"
	"github.com/pocethereum/pochain/log"
	"github.com/cybergarage/go-net-upnp/net/upnp"
	"github.com/cybergarage/go-net-upnp/net/upnp/util"
	gofstab "github.com/deniswernert/go-fstab"
	"io/ioutil"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func getMacAndIp() (mac string, ip string, err error) {
	ifis, _ := util.GetAvailableInterfaces()
	log.Info("GetAvailableInterfaces", "ifis", ifis)

	mac = ifis[0].HardwareAddr.String()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic("get local network ip failed")
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && strings.HasPrefix(ipnet.IP.String(), "192") {
			if ipnet.IP.To4() != nil {
				log.Info("GetAvailableInterfaces", "ip", ipnet.IP.String())
				ip = ipnet.IP.String()
			}
		}
	}
	return
}

func getPlotSize() (plotsize uint64) {
	if gDev == nil {
		return 0
	}
	engine := gDev.ieth.Engine()
	if pocengine, ok := engine.(*poc.Poc); ok && pocengine != nil {
		plotsize = pocengine.GetSize()
	} else {
		plotsize = uint64(0)
	}
	return plotsize
}

// disk usage of path/disk
func diskUsage(path string) (disk upnp.DiskInfo) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}

func getMounts() (mounts gofstab.Mounts, err error) {
	switch runtime.GOOS {
	case "linux":
		mounts, err = gofstab.ParseSystem()
	case "darwin":
		if common.ISBINGDEBUG {
			m := &gofstab.Mount{
				File: "/Users/bing/Library/Poc/",
			}
			mounts = append(mounts, m)
		} else {
			cmd := exec.Command("mount", "-t", "hfs")
			if stdout, cmderr := cmd.StdoutPipe(); cmderr != nil {
				return mounts, cmderr
			} else {
				defer stdout.Close() // 保证关闭输出流
				if starterr := cmd.Start(); starterr != nil {
					return mounts, starterr
				}
				if result, readerr := ioutil.ReadAll(stdout); readerr != nil {
					return mounts, readerr
				} else {
					log.Info("mount -t hfs result:", "result", string(result))
				}
			}
			//TODO:bing
		}

	case "windows":

	}
	return
}

func fillDiskInfo(diskinfo *upnp.DiskInfo) {
	diskinfo.All = 0
	diskinfo.Count = 0
	diskinfo.Used = 0
	mounts, _ := getMounts()
	for _, val := range mounts {
		log.Info("get mounts filesystem size", "Path", val.File)
		if val.File == "swap" || val.File == "/dev/shm" || val.File == "/dev/pts" || val.File == "/proc" || val.File == "/sys" {
			continue
		}
		if (!common.ISBINGDEBUG) && val.File == "/" {
			continue
		}

		disk := diskUsage(val.File)
		if disk.All <= 1024 {
			continue
		}
		log.Info("get mounts filesystem size", "Disk", disk)
		diskinfo.Used += disk.Used
		diskinfo.All += disk.All
		diskinfo.Free += disk.Free
		diskinfo.Count += 1
	}
}

func getAvaliablePlotDiskInfo() (plots []Plot, err error) {
	mounts, _ := getMounts()
	for _, val := range mounts {
		log.Info("get mounts filesystem size", "Path", val.File)
		if val.File == "swap" || val.File == "/dev/shm" || val.File == "/dev/pts" || val.File == "/proc" || val.File == "/sys" {
			continue
		}
		if (!common.ISBINGDEBUG) && val.File == "/" {
			continue
		}

		disk := diskUsage(val.File)
		if disk.All <= 1024 {
			continue
		}
		log.Info("get mounts filesystem size", "Disk", disk)
		rowp := Plot{
			Id:       0,
			Name:     val.File,
			Path:     val.File,
			Uuid:     "",
			PlotSize: 0,
			DiskSize: disk.All,
			FreeSize: disk.Free,
			Status:   PLOT_STATUS_UNUSED,
		}
		plots = append(plots, rowp)
	}
	return plots, err
}

func GetSettingPlotdirs() (plotpaths string) {
	plotpatharray := []string{}
	plots, err := (&Plot{}).QueryAllPlotInfo()
	if err != nil {
		return
	}
	for _, plot := range plots {
		plotpath := plot.GetFullPlotPath()
		plotpatharray = append(plotpatharray, plotpath)
	}
	return strings.Join(plotpatharray, ",")
}
