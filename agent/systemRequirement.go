package agent

import (
	"encoding/json"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/cpu"
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGetSystemInfo = modkernel32.NewProc("GetSystemInfo")
	procGetTickCount  = modkernel32.NewProc("GetTickCount")
)

type SYSTEM_INFO struct {
	wProcessorArchitecture      uint16
	wReserved                   uint16
	dwPageSize                  uint32
	lpMinimumApplicationAddress *byte
	lpMaximumApplicationAddress *byte
	dwActiveProcessorMask       uintptr
	dwNumberOfProcessors        uint32
	dwProcessorType             uint32
	dwAllocationGranularity     uint32
	wProcessorLevel             uint16
	wProcessorRevision          uint16
}

type SystemInfo struct {
	Core           uint32  `json:"core"`
	SSD            float64 `json:"ssd"`
	RAM            float64 `json:"ram"`
	CpuUtilization float64 `json:"cpuutilization"`
}

func getWindowSystemInfo() SystemInfo {
	var sysInfo SYSTEM_INFO

	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&sysInfo)))

	Core := sysInfo.dwNumberOfProcessors

	// Solid State Drive
	out, err := exec.Command("wmic", "LogicalDisk", "Where", "DriveType=3", "get", "Size,FreeSpace").Output()
	if err != nil {
		// return WalletSystemInfo{}, err
	}
	lines := strings.Split(string(out), "\n")
	var free int64
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 1 {
			v, _ := strconv.ParseInt(fields[0], 10, 64)
			free = v
		}
	}
	ssd := float64(free) / (1024 * 1024 * 1024)

	// RAM
	out, err = exec.Command("wmic", "ComputerSystem", "get", "TotalPhysicalMemory").Output()
	if err != nil {
		// return WalletSystemInfo{}, err
	}
	fields := strings.Fields(string(out))
	ram, _ := strconv.ParseInt(fields[1], 10, 64)
	ramGB := float64(ram) / (1024 * 1024 * 1024)

	// CPUUtilization
	cpuUtilization_percentages, err := cpu.Percent(time.Second, true)
	if err != nil {
		// return WalletSystemInfo{}, err
	}

	cpuUtilization := cpuUtilization_percentages[0]

	systemInfo := SystemInfo{
		Core:           Core,
		SSD:            ssd,
		RAM:            ramGB,
		CpuUtilization: cpuUtilization,
	}

	////fmt.Println("core", Core)
	////fmt.Println("ssd", ssd)
	////fmt.Println("ram", ramGB)
	////fmt.Println("cpu utilization", cpuUtilization)

	////fmt.Println("")

	return systemInfo

}

func FetchSystemRequirements(walletAddress string, peerId string) []byte {
	systemInfo := getWindowSystemInfo()
	//this is temporary just for testing, until all info nis real like node utilization, uptime and all
	rand.Seed(time.Now().UnixNano())
	// Generate a random integer between 0 and 100
	randomNum := rand.Float64()
	user_SystemInfo := User_SystemInfo{
		SystemInfo:    systemInfo,
		WalletAddress: walletAddress,
		PeerID:        peerId,
		RandomNumber:  randomNum,
	}
	user_SystemInfo_bytes, err := json.MarshalIndent(user_SystemInfo, "", " ")
	if err != nil {
		// handle error
		panic(err)
	}

	return user_SystemInfo_bytes
}

func calculateSSDS(solidStateDriveCapacity float64) float64 {
	// Apply constraints for minimum and maximum scores
	var ssds float64
	switch {
	case solidStateDriveCapacity <= 2048:
		ssds = 0.0
	case solidStateDriveCapacity >= 16384:
		ssds = 1.0
	default:
		ssds = (solidStateDriveCapacity / (1024 * 14.0)) - (1.0 / 7.0)
	}

	return ssds
}

func calculateRAM(ramInstalled float64) float64 {
	// Apply constraints for minimum and maximum score
	var ram float64
	switch {
	case ramInstalled <= 4:
		ram = 00
	case ramInstalled >= 32:
		ram = 1.0
	default:
		ram = (ramInstalled / 28) - (1.0 / 7.0)
	}

	return ram
}

func calculateCORE(coresInstalled uint64) float64 {
	// Apply constraints for minimum and maximum scores
	var core float64

	switch {
	case coresInstalled <= 4:
		core = 0.0
	case coresInstalled >= 16:
		core = 1.0
	default:
		core = (float64(coresInstalled) / 12) - (1 / 3)
	}

	return core
}
