package enum

type log string

const (
	Agent_logFile         log = "logs/agent.csv"
	ValidatorPool_logFile log = "logs/validatodpool.csv"
	Consensus_logFile     log = "logs/consensus.csv"
)
