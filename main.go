package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"jumbochain.org/accounts"
	"jumbochain.org/agent"
	"jumbochain.org/block"
	"jumbochain.org/bootnode"
	"jumbochain.org/consensus"
	"jumbochain.org/enum"
	"jumbochain.org/filemanagement"
	"jumbochain.org/p2p"
	"jumbochain.org/temp"
	"jumbochain.org/transaction"
)

func main() {

	startNode := flag.Int("startNode", 0, "starts node")
	sourcePort := flag.Int("sourcePort", 3001, "source port number")
	Validator := flag.Int("Validator", 0, "is validator")
	validator_address := flag.String("address", "", "validator address")
	newAccount := flag.Int("newAccount", 0, "creats new Account")
	sendTrx := flag.Int("sendTrx", 0, "sends transaction")
	from := flag.String("from", "", "from address for transaction")
	to := flag.String("to", "", "to address for transaction")
	auth := flag.String("auth", "", "authorization string")
	value := flag.Int("value", 0, "value  for transaction")
	initGenesis := flag.Int("initGenesis", 0, "init genesis command")
	balance := flag.String("balance", "", "get balance of address")
	currentBlockNumber := flag.Int("currentBlockNumber", 0, "current block number")
	getBlockHashByNumber := flag.String("getBlockHashByNumber", "", "current block number")
	getBlockByHash := flag.String("getBlockByHash", "", "current block number")
	getTransactionByHash := flag.String("getTransactionByHash", "", "current block number")

	flag.Parse()

	if *getTransactionByHash != "" {
		blockHash := block.GetBlockByHash(*getTransactionByHash)
		fmt.Println("Transaction is :")
		fmt.Println(blockHash)
	}

	if *getBlockByHash != "" {
		block := block.GetBlockByHash(*getBlockByHash)
		fmt.Println("block is :")
		fmt.Println(block)
	}

	if *getBlockHashByNumber != "" {
		blockHash := block.GetBlockHashByNumber(*getBlockHashByNumber)
		fmt.Println("block hash is :")
		fmt.Println(blockHash)
	}

	if *currentBlockNumber != 0 {
		currentBlockNumber := block.GetCurrentBlockNumber()
		fmt.Println("current block number is :")
		fmt.Println(currentBlockNumber)

	}

	if *balance != "" {
		balanceOfAddress := block.GetBalance(*balance)
		fmt.Println("balance of address is :")
		fmt.Println(balanceOfAddress)
	}

	if *sendTrx == 1 {
		transaction.SendTxx2(*from, *to, *value, *auth)
	}

	if *newAccount == 1 {
		accounts.NewAccount()
	}

	if *initGenesis == 1 {
		block.InitGenesis()
		fmt.Println("Genesis has been loaded successfully")
	}

	if *startNode == 1 {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// jumborpc.RunRpcServer()

		var node_multiAddr string
		r := rand.Reader

		h, err := p2p.MakeHost(*sourcePort, r)
		if err != nil {
			log.Fatal(err)
		}

		node_multiAddr = p2p.GetHostAddress(h)
		p2p.Node_multiAddr = node_multiAddr
		p2p.Host = h

		filemanagement.AppendTofile(string(enum.Peerlist), node_multiAddr)

		p2p.StartListener(ctx, h, *sourcePort, p2p.HandleStream, p2p.HandleStream2, p2p.HandleStream3)

		//set true for bootNode
		var isBootNode bool = true

		if !isBootNode {
			recordlength := filemanagement.GetNumberOfRecords(string(enum.Peerlist))

			if recordlength == 1 {
				target := string(bootnode.Bootnode1)
				rw, err := p2p.ConnectionForPeerlist(ctx, h, target)
				if err != nil {
					//log.Println(err)
					return
				}

				// Create a thread to read and write data.
				go p2p.WriteData2(rw, node_multiAddr)
				go p2p.ReadData2(rw)

			}

		} else {
			consensus.IsInValidatorPool = true
		}

		//uncomment it

		// updating peerlist
		ticker := time.NewTicker(10 * time.Second)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					////fmt.Println("- + - + - + -----updating peerlist")
					peerlist := temp.ReadCsv("peerlist.csv")
					for i := len(peerlist) - 1; i > -1; i-- {
						// ////fmt.Println(peerlist[i][0])
						target := peerlist[i][0]
						if target != node_multiAddr {
							rw, err := p2p.ConnectionForPeerlist(ctx, h, target)
							if err != nil {
								////log.Println(err)
								return
							}

							// Create a thread to read and write data.
							go p2p.WriteData2(rw, "updatePeerlist")
							go p2p.ReadData2(rw)
						}

					}

				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()

		// //-------------------------------------------------------------------------------------

		// //GO routine for updating validator peerlist

		ticker1 := time.NewTicker(6 * 10 * time.Second)
		quit1 := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker1.C:
					////fmt.Println("- + - + - + -----updating  validator peerlist")
					peerlist := temp.ReadCsv("peerlist.csv")
					for i := 0; i < len(peerlist); i++ {
						// ////fmt.Println(peerlist[i][0])
						target := peerlist[i][0]
						if target != node_multiAddr {
							rw, err := p2p.ConnectionForValidatorPeerlist(ctx, h, target)
							if err != nil {
								////log.Println(err)
								return
							}

							// Create a thread to read and write data.
							go p2p.WriteData4(rw)
							go p2p.ReadData4(rw)
						}

					}
				case <-quit1:
					ticker1.Stop()
					return
				}
			}
		}()

		//NODE SYNCING

		ticker_nodeSyncing := time.NewTicker(25 * time.Second)
		quit_nodeSyncing := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker_nodeSyncing.C:

					//fmt.Println("++++++++++++")
					//fmt.Println("try node syncing")
					//fmt.Println(p2p.NodeSyncingInProcess)
					//fmt.Println("++++++++++++")
					if !p2p.NodeSyncingInProcess {
						//fmt.Println("----------------------")
						//fmt.Println("Node Syncing in is progress")
						//fmt.Println("----------------------")
						//changhe this with filemanagemenrt
						peerlist := temp.ReadCsv("peerlist.csv")
						for i := len(peerlist) - 1; i > -1; i-- {
							target := peerlist[i][0]
							if target != node_multiAddr {
								rw, err := p2p.StartPeerAndConnect_NodeSyncing(ctx, h, target)
								if err == nil {
									//fmt.Println("----------------------")
									//fmt.Println("its TRUE")
									//fmt.Println("----------------------")
									p2p.NodeSyncingInProcess = true

									blockHeight := block.GetCurrentBlockNumber()
									p2p.NodeSyncingRequestWrite(rw, blockHeight)
									go p2p.NodeSyncingRequestRead(rw)

									break

								} else {
									continue
								}

							}
						}
					}

				case <-quit_nodeSyncing:
					ticker_nodeSyncing.Stop()
					return
				}
			}
		}()

		//NODE SYNCING

		//blockExtended sharding execution

		var transactionFetchinInprocess bool = false
		ticker_sharding := time.NewTicker(15 * time.Second)
		quit_sharding := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker_sharding.C:
					if !transactionFetchinInprocess {

						currentBlockExtendedNumber_string := block.GetCurrentBlockExtendedNumber()
						currentBlockExtendedNumber, _ := strconv.Atoi(currentBlockExtendedNumber_string)

						executedBlockExtendedNumber_string := block.GetExecutedBlockExtendedNumber()
						executedBlockExtendedNumber, _ := strconv.Atoi(executedBlockExtendedNumber_string)

						if executedBlockExtendedNumber < currentBlockExtendedNumber {
							transactionFetchinInprocess = true

							for blockExtendedNumber := executedBlockExtendedNumber + 1; blockExtendedNumber < currentBlockExtendedNumber; blockExtendedNumber++ {

								involvedAddresses, involvedPeers := block.GetBlockExtendedInfo(blockExtendedNumber)

								keystoreAddresses := filemanagement.GetAllRecords(string(enum.Keystore))

								m := make(map[string]bool)
								var commonAddresses []string

								// Store elements from the first column of arr1 in the map
								for _, row := range keystoreAddresses {
									m[row[0]] = true
								}

								// Check elements from arr2 against the map
								for _, str := range involvedAddresses {
									if m[str] {
										commonAddresses = append(commonAddresses, str)
									}
								}

								if len(commonAddresses) > 0 {

									for i := 0; i < len(involvedPeers); i++ {
										target := involvedPeers[i]
										rw, err := p2p.Validator_Trxs_StartPeerAndConnect(p2p.Host, target)
										if err == nil {
											p2p.TransactionRequestWrite(rw, blockExtendedNumber)

											break
										} else {
											continue
										}

									}

								}
								//delete blockExtendedNumber's information
								block.DeleteBlockExtendedData(blockExtendedNumber)

							}
							transactionFetchinInprocess = false

						}

					}

				case <-quit_sharding:
					ticker_sharding.Stop()
					return
				}
			}
		}()

		//send transactions to validator
		ticker3 := time.NewTicker(10 * time.Second)
		quit3 := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker3.C:
					validatorpeerlist := filemanagement.GetAllRecordsPeerlist(string(enum.ValidatorPeerlist))

					for i := 0; i < len(validatorpeerlist); i++ {
						////fmt.Println(validatorpeerlist[i][0])
						target := validatorpeerlist[i][1]
						info := p2p.RunSender(h, target)

						p2p.SendStream(h, info)

					}
					if err := os.Truncate("TrxMemPool.csv", 0); err != nil {
						//log.Printf("Failed to truncate: %v", err)
					}
				case <-quit3:
					ticker3.Stop()
					return
				}
			}
		}()

		if *Validator == 1 {
			//TODO
			//if validator shut down the old score whould be assumed soooo
			//do delete nodes that are not active

			//ALSO, Find a way to tell, which nodes are active! (atlease to node with database sync)

			//

			//this has to be improved a lot, check validatity of adddress and OL

			//1

			if *validator_address == "" {
				//TODO check validator adress from keystore
				//fmt.Println("enter a valid validator address")

			}

			//agent

			filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"0", time.Now().Format("2006-01-02 15:04:05"), "Agent Started"})
			ticker_agent := time.NewTicker(1 * 40 * time.Second)
			quit_agent := make(chan struct{})
			go func() {
				for {
					select {
					case <-ticker_agent.C:

						filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"1", time.Now().Format("2006-01-02 15:04:05"), "agent process repeats"})
						body_InBytes := agent.FetchSystemRequirements(*validator_address, node_multiAddr)

						validatorpool_peerlist := filemanagement.GetAllRecordsPeerlist("ValidatorPool.csv")
						filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"2", time.Now().Format("2006-01-02 15:04:05"), "Fetched System Requiremnts", "len(validatropool)-> ", strconv.Itoa(len(validatorpool_peerlist))})

						for i := len(validatorpool_peerlist) - 1; i > -1; i-- {

							target := validatorpool_peerlist[i][1]

							if target != node_multiAddr {
								filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"3", time.Now().Format("2006-01-02 15:04:05"), "Current Validator is-> ", validatorpool_peerlist[i][0], target})

								rw, err := p2p.ValidatorPool_StartPeerAndConnect(ctx, h, target)
								if err == nil {
									filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"4", time.Now().Format("2006-01-02 15:04:05"), "Write and Read Started"})
									go p2p.WriteData7(rw, body_InBytes)
									go p2p.ReadData7(rw)

								} else {
									filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"ERROR: ", time.Now().Format("2006-01-02 15:04:05"), err.Error()})
									continue
								}
							}
							// //fmt.Println("i: ", i)
							time.Sleep(10 * time.Second)
						}
						filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"7", time.Now().Format("2006-01-02 15:04:05"), "Finished sending all validators"})

						totalScore, isInValidatorpool := agent.UpdateValidatorPool(body_InBytes)

						if !isInValidatorpool {
							if isBootNode {
								isInValidatorpool = true
							}
						}

						consensus.SetValidatorInfo(totalScore, isInValidatorpool)

						filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"8", time.Now().Format("2006-01-02 15:04:05"), "isInValidatorpool and total score of this node is saved as ->", strconv.FormatBool(isInValidatorpool), strconv.FormatFloat(totalScore, 'f', -1, 64)})

					case <-quit_agent:
						ticker_agent.Stop()
						return
					}
				}
			}()

			//N_account_sign_sendTrx_VN_agent_validatorPool_verifyTrx_validator_consensus_createBlock

			//this should be done when minimum criteria is checked after agent response
			//and this is temporary

			//TODO
			//validator node will respond with true and false, save that in a variable

			//this validity check should be repeated every x seconds

			filemanagement.AddArrayToFile(string(enum.ValidatorPool_logFile), []string{"0", time.Now().Format("2006-01-02 15:04:05"), "Consensus Started"})
			ticker_isInPool := time.NewTicker(20 * time.Second)
			defer ticker_isInPool.Stop()

			for {
				select {
				case <-ticker_isInPool.C:
					filemanagement.AddArrayToFile(string(enum.ValidatorPool_logFile), []string{"1", time.Now().Format("2006-01-02 15:04:05"), "Consensus process repeats", "is In Pool:", strconv.FormatBool(consensus.IsInValidatorPool)})

					if consensus.IsInValidatorPool {
						// //fmt.Println("hi")

						var validatorPeerInfo []string
						validatorPeerInfo = append(validatorPeerInfo, *validator_address)
						validatorPeerInfo = append(validatorPeerInfo, node_multiAddr)
						totalScore := strconv.FormatFloat(consensus.TotalScore, 'f', 6, 64)

						validatorPeerInfo = append(validatorPeerInfo, totalScore)

						arrayTOAdd := []string{"2", time.Now().Format("2006-01-02 15:04:05"), "ValidatroInfo is: "}
						arrayTOAdd = append(arrayTOAdd, validatorPeerInfo[0], validatorPeerInfo[1], validatorPeerInfo[2])
						filemanagement.AddArrayToFile(string(enum.ValidatorPool_logFile), arrayTOAdd)

						//fmt.Println("consensus will run!")
						//fmt.Println("every 20 sec!")

						//this is listner for all the members in validator pool

						p2p.ValidatorPool_runListner(ctx, h, *sourcePort, p2p.ValidatorPool_HandleStream, p2p.ValidatorPool_Trxs_HandleStream)

						filemanagement.AddArrayToFile(string(enum.ValidatorPool_logFile), []string{"3", time.Now().Format("2006-01-02 15:04:05"), "Listner started"})
						IsAlreadyInFile := filemanagement.IsAlreadyInValidatorPool(string(enum.ValidatorPeerlist), validatorPeerInfo)
						////fmt.Println(IsAlreadyInFile)
						if !IsAlreadyInFile {
							filemanagement.AddArrayToFile(string(enum.ValidatorPeerlist), validatorPeerInfo)
						}

						filemanagement.AddArrayToFile(string(enum.ValidatorPool_logFile), []string{"4", time.Now().Format("2006-01-02 15:04:05"), "iIsAlreadyInFile:", strconv.FormatBool(IsAlreadyInFile)})
						//verify transaction

						filemanagement.AddArrayToFile(string(enum.ValidatorPool_logFile), []string{"5", time.Now().Format("2006-01-02 15:04:05"), "Verification started"})
						ticker := time.NewTicker(6 * 10 * time.Second)
						quit := make(chan struct{})
						go func() {
							for {
								//fmt.Println("in verify")
								select {
								case <-ticker.C:
									// ////fmt.Println("every 5 sec!")
									transaction.Temp_VerifyTransaction()
								case <-quit:
									ticker.Stop()
									return
								}
							}
						}()

						//consessus

						filemanagement.AddArrayToFile(string(enum.ValidatorPool_logFile), []string{"6", time.Now().Format("2006-01-02 15:04:05"), "Consessus started"})
						filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"0", time.Now().Format("2006-01-02 15:04:05"), "Consessus started"})
						ticker_consessus := time.NewTicker(3 * 60 * time.Second)
						quit_consessus := make(chan struct{})
						go func() {
							for {
								select {
								case <-ticker_consessus.C:

									filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"1", time.Now().Format("2006-01-02 15:04:05"), "Consessus repeated again"})
									isValidator, isLeader := consensus.SelectValidators(node_multiAddr)

									filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"2", time.Now().Format("2006-01-02 15:04:05"), "isValidator? ", strconv.FormatBool(isValidator), "isLeader?", strconv.FormatBool(isLeader)})

									if isValidator {

										//this is the listner for the member of current validation

										//TODO: if node is not a validator now, then it should stop this listning channel

										p2p.Validator_runListner(ctx, h, *sourcePort, p2p.Validator_HandleStream, p2p.Validator_SaveTrx_HandleStream)

										filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"3", time.Now().Format("2006-01-02 15:04:05"), "Listner Started"})
										if isLeader {
											//fmt.Println("I am the leader")

											//verify block hash somewhere

											// /------------------------------------------------------
											transactions_, _ := consensus.FetchVerifiedTransactions()

											filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"4", time.Now().Format("2006-01-02 15:04:05"), "number of transaction is verified memory pool are: ", strconv.Itoa(len(transactions_))})
											consensus.AddVerificationConcent()
											//store transaction hashes in a variable as well

											//---send full transactions to validators
											//TODO: dont send to self
											//TODO: previous leader should not send through this channel
											transactions := consensus.FetchTransactions(transactions_)

											arrayToBeAdded := []string{"5", time.Now().Format("2006-01-02 15:04:05"), "transactions that will be insrted in a block: "}

											for i := 0; i < len(transactions); i++ {
												arrayToBeAdded = append(arrayToBeAdded, transactions[i])
											}
											filemanagement.AddArrayToFile(string(enum.Consensus_logFile), arrayToBeAdded)

											validator_peerlist := temp.ReadCsv("Validators.csv")
											consensus.Validator_peerlist = validator_peerlist

											filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"6", time.Now().Format("2006-01-02 15:04:05"), "number of members that are validator: ", strconv.Itoa(len(validator_peerlist))})
											//fmt.Println(" start")
											for i := len(validator_peerlist) - 1; i > -1; i-- {
												validator_info := validator_peerlist[i][0]
												filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"7", time.Now().Format("2006-01-02 15:04:05"), "sending transactions to this validator: ", validator_info})
												var validatorInfo consensus.ValidatorInfo
												json.Unmarshal([]byte(validator_info), &validatorInfo)

												target := validatorInfo.MultiAddr
												if target != node_multiAddr {

													filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"8", time.Now().Format("2006-01-02 15:04:05"), "target multiAdd", target})

													rw, err := p2p.Validator_StartPeerAndConnect(h, target)

													if err == nil {
														filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"9", time.Now().Format("2006-01-02 15:04:05"), "Read and Write Started"})
														go p2p.LeaderWrite(rw, transactions)
														go p2p.LeaderRead(rw)

													} else {
														filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"ERROR", time.Now().Format("2006-01-02 15:04:05"), err.Error()})
														continue
													}

													//go leaderread
													//this will count the concent and when 51% is crossed create block
												}
											}

											// /------------------------------------------------------

											////fmt.Println(transactions, transactions)

											//send all hashes to remaning 28 validator and wait for response (stream will either be 'ok' or 'not ok')

											//update verified accourdingly
											//check if 51 % is crossed
											//if crossed -> create block
											// block.CreateBlock()

										} else {
											//start listner
											//fetch all the trx hases in a varaible
											//fetch transactions from TrxMemoryPoolValidator store it in varaible and empty csv file
											//loop arround all the transactions create its json body and check its hash with the hash list
											//if not in hash list push back in csv file
											//if in hash list -> verify it -> and keep a count of all the verified hashes
											//if last hash from the list is verified --> send stream 'ok'
											//if any one of the hash is not verified --> send stream 'not ok'

										}

									}
									// isNodeValidator := consensus.IsNodeValidator
									//select top validator
								case <-quit_consessus:
									ticker_consessus.Stop()
									return
								}
							}
						}()

					}
				}
			}

		}
		<-ctx.Done()

	}

}
