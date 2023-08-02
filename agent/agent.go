package agent

import (
	"encoding/json"
	"strconv"
	"time"

	"jumbochain.org/bootnode"
	"jumbochain.org/filemanagement"
	"jumbochain.org/ldb"
)

const (
	//TODO
	//this is just a demo data
	// real minimum requirements has to be set

	//as you can see

	minCore            = 0
	minSsd             = 0
	minRam             = 0
	minCpuutilization  = 0
	minUptime          = 0
	minNodeUtilization = 0
	minStake           = 0
	minESG_score       = 0
)

type User_SystemInfo struct {
	SystemInfo    SystemInfo `json:"systemInfo"`
	WalletAddress string     `json:"walletAddress"`
	PeerID        string     `json:"peerID"`
	RandomNumber  float64    `json:"randomNumber"`
}

type validator_info struct {
	WalletAddress   string     `json:"walletAddress"`
	PeerID          string     `json:"peerID"`
	SystemInfo      SystemInfo `json:"systemInfo"`
	Uptime          time.Time  `json:"uptime"`
	NodeUtilization float64    `json:"nodeutilization"`
	Stake           uint64     `json:"staked_amount"`
	ESG_score       uint64     `json:"esg_score"`
	TotalScore      float64    `json:"totalscore"`
	Active          bool       `json:"active"`
}

func UpdateValidatorPool(User_SystemInfo_InBytes []byte) (float64, bool) {

	var user_SystemInfo User_SystemInfo
	json.Unmarshal(User_SystemInfo_InBytes, &user_SystemInfo)

	totalScore, validatorInfo := calculateScore(user_SystemInfo)

	//these both steps can be combined using same ldb.
	updateSystemAndScoreInfoOfValidatoroDB(validatorInfo)
	//yaa, this is the other one
	isInValidatorpool := checkMinimumRequirement(validatorInfo)

	if !isInValidatorpool {
		if user_SystemInfo.PeerID == string(bootnode.Bootnode1) {
			isInValidatorpool = true
		}
	}

	return totalScore, isInValidatorpool
}

func calculateScore(user_systemInfo User_SystemInfo) (float64, validator_info) {

	//this function returns hardcoded output, for now
	validator_info := getValidatorInfo(user_systemInfo)

	//this isb also hardcoded
	total_stakedAmount := []uint64{20000, 40000, 50000, 80000, 10000}

	//for now, I am using core insted of cpu utilization to stop score from cahnging so quiclkly
	totalScore := totalScore(validator_info.Stake, total_stakedAmount, validator_info.Uptime, user_systemInfo.SystemInfo.SSD, user_systemInfo.SystemInfo.RAM, uint64(user_systemInfo.SystemInfo.CpuUtilization), float64(validator_info.ESG_score), validator_info.NodeUtilization)

	totalScore = totalScore + user_systemInfo.RandomNumber

	validator_info.TotalScore = totalScore

	return totalScore, validator_info
}

func totalScore(stake uint64, stakedAmounts []uint64, prevUptime time.Time, ssdv float64, ramv float64, CPUUtilization uint64, sustainabilityScore float64, nodeUtilizationScore float64) float64 {
	stake_ := calculateStakeCoefficient(stake, stakedAmounts)
	uptime := calculateUptime(prevUptime)
	nodecapacity := nodeCapacityScore(ssdv, ramv, CPUUtilization)
	sustainability := sustainabilityScore
	nodeutilization := nodeUtilizationScore
	return (stake_ * 0.3) + (uptime * 0.2) + (nodecapacity * 0.15) + (sustainability * 0.25) + (nodeutilization * 0.1)
}

func calculateStakeCoefficient(S uint64, stakedAmounts []uint64) float64 {
	var TS uint64
	for _, amount := range stakedAmounts {
		TS += amount
	}
	return float64(S) / float64(TS)
}

func nodeUtilizationScore(totalBlocksMined float64) float64 {
	// compute USi
	nodeUtilizationScore := float64(100-totalBlocksMined) / 100

	// ensure USi has a lower bound of 0
	if nodeUtilizationScore < 0 {
		nodeUtilizationScore = 0
	}

	return nodeUtilizationScore
}

func calculateUptime(prevUptime time.Time) float64 {

	//this function has to be changed
	//go throught white paper

	currentTime := time.Now()
	// Calculate the difference between the previous uptime and the current time

	// uptimeDuration := currentTime.Sub(currentTime.Add(-prevUptime))

	uptimeDuration := currentTime.Sub(prevUptime)

	uptimeMinutes := uptimeDuration.Minutes()
	// If the uptime is greater than or equal to 15 minutes, return 0 uptime percentage
	if uptimeMinutes >= 15.0 {
		return 0.0
	}

	// Calculate the uptime percentage using the given formula
	uptimeScore := (15.0 - uptimeMinutes) / 15.0

	return uptimeScore
}

func nodeCapacityScore(ssdv float64, ramv float64, cpuutilv uint64) float64 {
	ssdScore := calculateSSDS(ssdv) * 0.2
	ramScore := calculateRAM(ramv) * 0.4
	cpuutilScore := calculateCORE(cpuutilv) * 0.4 // put cpuutilization here
	nodeCapacityScore := ssdScore + ramScore + cpuutilScore
	return nodeCapacityScore
}

func updateSystemAndScoreInfoOfValidatoroDB(validaotrInfo validator_info) {
	// var validator_info validator_info
	key := validaotrInfo.WalletAddress
	value_byte := ldb.GetInfoDB("validatordb", []byte(key))

	if len(value_byte) < 1 {
		//if new user, add user with default values, system info and total score
		validaotrInfo_bytes, _ := json.Marshal(validaotrInfo)
		ldb.AddInfoDB("validatordb", []byte(key), validaotrInfo_bytes)
	} else {
		//if user already exist, update system info and total score
		var validatorInfo_existing validator_info
		json.Unmarshal(value_byte, &validatorInfo_existing)

		validatorInfo_existing.SystemInfo = validaotrInfo.SystemInfo
		validatorInfo_existing.TotalScore = validaotrInfo.TotalScore

		validaotrInfo_bytes, _ := json.Marshal(validatorInfo_existing)
		ldb.AddInfoDB("validatordb", []byte(key), validaotrInfo_bytes)
	}

}

func checkMinimumRequirement(validaotrInfo_ validator_info) bool {

	//TODO
	//this is just a demo data
	// real minimum requirements has to be set
	//read white paper
	//ignoring uptime for now
	//mybe total score too has a minimum requiremnt

	key := validaotrInfo_.WalletAddress
	value_byte := ldb.GetInfoDB("validatordb", []byte(key))
	var validatorInfo validator_info
	json.Unmarshal(value_byte, &validatorInfo)

	// previousActiveStatus := validatorInfo.Active
	// key := validaotrInfo.WalletAddress

	if validatorInfo.ESG_score >= minESG_score &&
		validatorInfo.NodeUtilization >= minNodeUtilization &&
		validatorInfo.Stake >= minStake &&
		validatorInfo.SystemInfo.Core >= minCore &&
		validatorInfo.SystemInfo.RAM >= minRam &&
		validatorInfo.SystemInfo.SSD >= minSsd {

		//if all mininum requirements are full filed
		//update 'active' in database
		// //fmt.Println("minimum requiremnt is full filled")

		validatorInfo.Active = true

		//maybe-> look into it after words

		filemanagement.RemoveArrayFromFile("ValidatorPool.csv", key)
		totalScore := strconv.FormatFloat(validatorInfo.TotalScore, 'f', 6, 64)

		var validatorpool_info = []string{key, validatorInfo.PeerID, totalScore}
		filemanagement.AddArrayToFile("ValidatorPool.csv", validatorpool_info)

		// if previousActiveStatus == false {
		// 	//fmt.Println("previously false")
		// 	//Add to validator pool

		// 	totalScore := strconv.FormatFloat(validatorInfo.TotalScore, 'f', 6, 64)

		// 	var validatorpool_info = []string{key, validatorInfo.PeerID, totalScore}
		// 	filemanagement.AddArrayToFile("ValidatorPool.csv", validatorpool_info)

		// } else {
		// 	//fmt.Println("previously true")
		// 	filemanagement.RemoveArrayFromFile("ValidatorPool.csv", key)
		// 	totalScore := strconv.FormatFloat(validatorInfo.TotalScore, 'f', 6, 64)

		// 	var validatorpool_info = []string{key, validatorInfo.PeerID, totalScore}
		// 	filemanagement.AddArrayToFile("ValidatorPool.csv", validatorpool_info)

		// }

	} else {
		// //fmt.Println("minimum requiremnt is not  full filled")
		validatorInfo.Active = false

		//maybe-> look into it after words
		filemanagement.RemoveArrayFromFile("ValidatorPool.csv", key)

		// if previousActiveStatus == true {
		// 	//fmt.Println("previously true")

		// 	//remove from validator pool
		// 	filemanagement.RemoveArrayFromFile("ValidatorPool.csv", key)

		// }
	}

	validaotrInfo_bytes, _ := json.Marshal(validatorInfo)
	ldb.AddInfoDB("validatordb", []byte(key), validaotrInfo_bytes)

	return validatorInfo.Active

}

func getValidatorInfo(user_SystemInfo User_SystemInfo) validator_info {
	//this all is hardcoded for now
	var validator_info validator_info

	validator_info.WalletAddress = user_SystemInfo.WalletAddress
	validator_info.PeerID = user_SystemInfo.PeerID
	validator_info.SystemInfo = user_SystemInfo.SystemInfo

	validator_info.Stake = uint64(10000)
	validator_info.Uptime = time.Now()
	validator_info.ESG_score = uint64(80)

	//everything is hardcoded
	totalBlocksMined := float64(10)
	validator_info.NodeUtilization = nodeUtilizationScore(totalBlocksMined)

	validator_info.Active = false

	return validator_info

}
