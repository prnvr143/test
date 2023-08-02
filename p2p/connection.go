package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/multiformats/go-multiaddr"
	ma "github.com/multiformats/go-multiaddr"

	"jumbochain.org/agent"
	"jumbochain.org/consensus"
	"jumbochain.org/enum"
	"jumbochain.org/filemanagement"
	"jumbochain.org/temp"

	"jumbochain.org/block"
	"jumbochain.org/transaction"
)

// toTempFile = false

var TrueFalse = false

var NodeSyncingInProcess bool = false

var mutex sync.Mutex

var Node_multiAddr string

var Host host.Host

func MakeHost(port int, randomness io.Reader) (host.Host, error) {
	// Creates a new RSA key pair for this host.

	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.Secp256k1, 2048, randomness)
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	// 0.0.0.0 will listen on any interface device.
	// ipAddress := GetOutboundIP()

	//127.0.0.1
	sourceMultiAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	return libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
}

func StartListener(ctx context.Context, ha host.Host, listenPort int, streamHandler network.StreamHandler, streamHandler2 network.StreamHandler, streamHandler3 network.StreamHandler) {

	ha.SetStreamHandler("/nodesyncing/1.0.0", streamHandler)

	ha.SetStreamHandler("/peerlist/1.0.0", streamHandler2)

	ha.SetStreamHandler("/validatorpeerlist/1.0.0", streamHandler3)

	var port string
	for _, la := range ha.Network().ListenAddresses() {
		if p, err := la.ValueForProtocol(multiaddr.P_TCP); err == nil {
			port = p
			break
		}
	}

	if port == "" {
		////log.Println("was not able to find actual local port")
		return
	}

	// // Set a stream handler on host A. /echo/1.0.0 is
	// // a user-defined protocol name.
	// ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
	// 	////log.Println("listener received new stream")
	// 	if err := doEcho1(s); err != nil {
	// 		////log.Println(err)
	// 		s.Reset()
	// 	} else {
	// 		s.Close()
	// 	}
	// })

	////log.Println("listening for connections")

}

//NODE SYNCING NETWROKEING CODE

func HandleStream(s network.Stream) {

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	NodeSyncingResponseRead(rw)
	NodeSyncingResponseWrite(rw, "")

}

// NodeSyncing
// validator read
func NodeSyncingResponseRead(rw *bufio.ReadWriter) {

	for {
		plus := "`"
		bytes_arr := []byte(plus)
		byte_arr := bytes_arr[0]
		str, _ := rw.ReadString(byte_arr)

		if str == "" {
			return
		}
		if str != "`" {

			//fmt.Println("block number to be fetched")

			//db
			currentBlockNumber := block.GetCurrentBlockNumber()
			//db

			// currentBlockNumber = strings.TrimRight(currentBlockNumber, "\n")
			currentBlockNumber_, _ := strconv.ParseInt(currentBlockNumber, 10, 64)
			////fmt.Println(err)
			currentBlockNumber_int := int(currentBlockNumber_)

			str = strings.TrimRight(str, "`")
			requestBlockNumber, _ := strconv.ParseInt(str, 10, 64)

			////fmt.Println(err)
			requestBlockNumber_int := int(requestBlockNumber)

			if requestBlockNumber_int < currentBlockNumber_int {

				block_number := str

				//fmt.Println(block_number)

				NodeSyncingResponseWrite(rw, block_number)
			} else {

				//fmt.Println("DONT HAVE THIS BLOCK NUMBER")
				NodeSyncingResponseWrite(rw, "-1")

			}

		}

	}
}

// NodeSyncing
// validator write
func NodeSyncingResponseWrite(rw *bufio.ReadWriter, block_number string) {

	if block_number != "" {

		if block_number != "-1" {
			//fmt.Println("fetching block from  validator node and send it")

			//db
			blockHash := block.GetBlockHashByNumber(block_number)

			block := block.GetBlockByHash(blockHash)
			//db

			rw.WriteString(fmt.Sprintf("%s`", block))

			rw.Flush()

		} else {
			//fmt.Println("Reached here!")

			rw.WriteString(fmt.Sprintf("%s`", block_number))
			rw.Flush()

		}

	}

}

// NodeSyncing
// node write
func NodeSyncingRequestWrite(rw *bufio.ReadWriter, blockHeight string) {

	block_height, _ := strconv.ParseInt(blockHeight, 10, 64)

	////fmt.Println(err)

	request_block := strconv.Itoa(int(block_height) + 1)

	//fmt.Println("requsting validator node for block number: ")
	//fmt.Println(request_block)

	rw.WriteString(fmt.Sprintf("%s`", request_block))

	rw.Flush()

}

// NodeSyncing
// node read
func NodeSyncingRequestRead(rw *bufio.ReadWriter) {
	for {

		plus := "`"
		bytes_arr := []byte(plus)
		byte_arr := bytes_arr[0]
		str, _ := rw.ReadString(byte_arr)
		// str, _ := rw.ReadString("`")

		if str == "" {
			return
		}
		if str != "`" {

			//fmt.Println("reciving block data")
			//fmt.Println("saving block data of block number: ", str)

			str = strings.TrimRight(str, "`")

			if str != "-1" {
				block_bytes, block_body := block.JsonToBlock(str)

				//db
				block.AddlatestBlock(block_bytes, block_body)
				//db

				//fmt.Println("block inserted ")
				//fmt.Println(block_body.BlockNumber)

				rw.Flush()

				next_blockNumber := block_body.BlockNumber

				request_blockNumber := strconv.Itoa(next_blockNumber)

				NodeSyncingRequestWrite(rw, request_blockNumber)

			} else {
				//fmt.Println("----------------------")
				//fmt.Println("its TRUE")
				//fmt.Println("----------------------")
				NodeSyncingInProcess = false

			}

		}

	}
}

// NodeSyncing
func StartPeerAndConnect_NodeSyncing(ctx context.Context, h host.Host, destination string) (*bufio.ReadWriter, error) {

	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		return nil, err
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, err
	}

	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	s, err := h.NewStream(context.Background(), info.ID, "/nodesyncing/1.0.0")
	if err != nil {
		return nil, err
	}

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	return rw, nil
}

//NODE SYNCING NETWROKEING CODE

func RunSender(ha host.Host, targetPeer string) peer.AddrInfo {

	// Set a stream handler on host A. /echo/1.0.0 is
	// a user-defined protocol name.
	ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
		////log.Println("sender received new stream")
		if err := doEcho1(s); err != nil {
			////log.Println(err)
			s.Reset()
		} else {
			s.Close()
		}
	})

	// Turn the targetPeer into a multiaddr.
	maddr, err := ma.NewMultiaddr(targetPeer)
	if err != nil {
		//log.Println(err)
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		//log.Println(err)
	}

	// We have a peer ID and a targetAddr so we add it to the peerstore
	// so LibP2P knows how to contact it
	ha.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	////log.Println("sender opening stream")

	return *info
	// make a new stream from host B to host A
	// it should be handled on host A by the handler we set above because
	// we use the same /echo/1.0.0 protocol

	// s, err := ha.NewStream(context.Background(), info.ID, "/echo/1.0.0")
	// if err != nil {
	// 	////log.Println(err)
	// 	return
	// }

	// ////log.Println("sender saying hello")

	// message := "ye hai message"

	// ////fmt.Println(message + "\n")

	// _, err = s.Write([]byte(message + "\n"))
	// if err != nil {
	// 	////log.Println(err)
	// 	return
	// }

	// out, err := io.ReadAll(s)
	// if err != nil {
	// 	////log.Println(err)
	// 	return
	// }

	// ////log.Printf("read reply: %q\n", out)
}

func SendStream(ha host.Host, info peer.AddrInfo) {

	trxs_inMemPool := temp.ReadCsv("TrxMemPool.csv")

	for i := 0; i < len(trxs_inMemPool); i++ {

		trx_details := trxs_inMemPool[i][0]

		s, err := ha.NewStream(context.Background(), info.ID, "/echo/1.0.0")
		if err != nil {
			//log.Println(err)
		}

		// ////log.Println("sender saying hello")

		// message := "ye hai message"

		//log.Println(trx_details)

		_, err = s.Write([]byte(trx_details + "\n"))
		////fmt.Println(err)
		if err != nil {
			//log.Println(err)
		} else {
			//fmt.Println("done")
		}

		out, err := io.ReadAll(s)
		if err != nil {
			//log.Println(err)
		}

		log.Printf("read reply: %q\n", out)

	}

	// do stuff

}

func SendStream1(ha host.Host, info peer.AddrInfo) {

	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				//ee

				trxs_inMemPool := temp.ReadCsv("TrxMemPool.csv")

				for i := 0; i < len(trxs_inMemPool); i++ {

					trx_details := trxs_inMemPool[i][0]

					s, err := ha.NewStream(context.Background(), info.ID, "/echo/1.0.0")
					if err != nil {
						////log.Println(err)
						return
					}

					// ////log.Println("sender saying hello")

					// message := "ye hai message"

					////log.Println(trx_details)

					_, err = s.Write([]byte(trx_details + "\n"))
					////fmt.Println(err)
					if err != nil {
						////log.Println(err)
						return
					}

					io.ReadAll(s)
					if err != nil {
						////log.Println(err)
						return
					}

					////log.Printf("read reply: %q\n", out)

					if err := os.Truncate("TrxMemPool.csv", 0); err != nil {
						////log.Printf("Failed to truncate: %v", err)
					}

				}

				// do stuff
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

}

func getPublicIp() string {
	cmd := exec.Command("curl", "ifconfig.me")
	output, err := cmd.Output()
	if err != nil {
		//fmt.Println("Failed to get public IP:", err)
		return err.Error()
	}
	publicIP := strings.TrimSpace(string(output))
	return publicIP
}

func GetHostAddress(ha host.Host) string {
	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", ha.ID().Pretty()))
	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := ha.Addrs()[0]

	// //fmt.Println(addr)
	addr_stringArray := strings.Split(addr.String(), "/")
	publicIp := getPublicIp()

	halfAddr := "/" + addr_stringArray[1] + "/" + publicIp + "/" + addr_stringArray[3] + "/" + addr_stringArray[4]

	//fmt.Println(halfAddr)

	// //fmt.Println(addr.String())
	fullAddr := halfAddr + hostAddr.String()
	//fmt.Println(fullAddr)

	////log.Printf("I am %s\n", fullAddr)
	return fullAddr
}

func doEcho(s network.Stream) error {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	////log.Println("0000000")
	////log.Printf("read : %s", str)
	_, err = s.Write([]byte(str))
	return err
}

func doEcho1(s network.Stream) error {
	buf := bufio.NewReader(s)

	str, err := buf.ReadBytes('\n')
	if err != nil {
		return err
	}

	//log.Println("111111111")
	//log.Printf("read : %s", str)

	temp.UpdateCsv("TrxMemPoolValidator.csv", string(str))

	// byy := temp.DecodeToPerson(str)

	// ////fmt.Println(byy)

	// _, err = s.Write([]byte(str))
	return err
}

//-----------------------------------------------------------------------------------------------------------------------------
// block sync two way connection is below

//-----------------------------------------------------------------------------------------------------------------------------
// update peerlist code is below

func HandleStream2(s network.Stream) {
	////log.Println(" 1   Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go ReadData3(rw)
	go WriteData3(rw, "")

	// stream 's' will stay open until you close it (or the other side closes it).
}

func WriteData2(rw *bufio.ReadWriter, command string) {
	rw.WriteString(fmt.Sprintf("%s\n", command))

	rw.Flush()
}

func ReadData2(rw *bufio.ReadWriter) {

	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			str = strings.TrimRight(str, "\n")

			isAlreadyInFile := filemanagement.IsAlreadyInFile(string(enum.Peerlist), str)
			if isAlreadyInFile == false {
				filemanagement.AppendTofile(string(enum.Peerlist), str)
			}

		}
	}
}

func WriteData3(rw *bufio.ReadWriter, peer_multiAddr string) {
	if peer_multiAddr != "" {
		rw.WriteString(fmt.Sprintf("%s\n", peer_multiAddr))

		rw.Flush()
	}

}

func ReadData3(rw *bufio.ReadWriter) {

	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {

			if str == "updatePeerlist\n" {
				peerlist := temp.ReadCsv("peerlist.csv")
				for i := 0; i < len(peerlist); i++ {
					target := peerlist[i][0]
					WriteData3(rw, target)

				}

			} else {
				peerlist := temp.ReadCsv("peerlist.csv")
				for i := 0; i < len(peerlist); i++ {
					target := peerlist[i][0]
					WriteData3(rw, target)

				}
				str = strings.TrimRight(str, "\n")
				filemanagement.AppendTofile(string(enum.Peerlist), str)
			}
		}
	}
}

func ConnectionForPeerlist(ctx context.Context, h host.Host, destination string) (*bufio.ReadWriter, error) {
	////log.Println("This node's multiaddresses:")

	// for _, la := range h.Addrs() {
	// 	////log.Printf(" - %v\n", la)
	// }
	////log.Println()

	// uniqueId := h.ID().Pretty()
	// multiAddressStr := h.Addrs()[0].String() + "/" + uniqueId

	////fmt.Println("multi address : ")
	////fmt.Println(multiAddressStr)

	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	// Start a stream with the destination.
	// Multiaddress of the destination peer is fetched from the peerstore using 'peerId'.
	s, err := h.NewStream(context.Background(), info.ID, "/peerlist/1.0.0")
	if err != nil {
		////log.Println(err)
		return nil, err
	}
	////log.Println("Established connection to destination")

	// s2, err := h.NewStream(context.Background(), info.ID, "/peerlist/1.0.0")
	// if err != nil {
	// 	////log.Println(err)
	// 	return nil, err
	// }

	// Create a buffered stream so that read and writes are non blocking.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// rw2 := bufio.NewReadWriter(bufio.NewReader(s2), bufio.NewWriter(s2))

	return rw, nil
}

//-----------------------------------------------------------------

//validator peer list / update validator peer list code below

func HandleStream3(s network.Stream) {
	////log.Println(" 1   Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go ReadData5(rw)
	go WriteData5(rw, "")

	// stream 's' will stay open until you close it (or the other side closes it).
}

func WriteData4(rw *bufio.ReadWriter) {
	rw.WriteString(fmt.Sprintf("%s\n", "updatevalidatorPeerlist"))

	rw.Flush()
}

func ReadData4(rw *bufio.ReadWriter) {

	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			str = strings.TrimRight(str, "\n")

			validaotrInfo := strings.Split(string(str), ",")

			IsAlreadyInFile := filemanagement.IsAlreadyInValidatorPool(string(enum.ValidatorPeerlist), validaotrInfo)
			////fmt.Println(IsAlreadyInFile)
			if !IsAlreadyInFile {
				filemanagement.AddArrayToFile(string(enum.ValidatorPeerlist), validaotrInfo)
			}

		}
	}
}

func WriteData5(rw *bufio.ReadWriter, validaotrInfo string) {
	if validaotrInfo != "" {
		rw.WriteString(fmt.Sprintf("%s\n", validaotrInfo))

		rw.Flush()
	}

}

func ReadData5(rw *bufio.ReadWriter) {

	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {

			if str == "updatevalidatorPeerlist\n" {
				peerlist := filemanagement.GetAllRecordsPeerlist(string(enum.ValidatorPeerlist))
				for i := 0; i < len(peerlist); i++ {
					// ////fmt.Println(peerlist[i][0])
					// target := peerlist[i][0]
					validatorInfo := peerlist[i][0] + "," + peerlist[i][1] + "," + peerlist[i][2]
					WriteData5(rw, validatorInfo)

				}

			}
		}
	}
}
func ConnectionForValidatorPeerlist(ctx context.Context, h host.Host, destination string) (*bufio.ReadWriter, error) {
	////log.Println("This node's multiaddresses:")

	// for _, la := range h.Addrs() {
	// 	////log.Printf(" - %v\n", la)
	// }
	////log.Println()

	// uniqueId := h.ID().Pretty()
	// multiAddressStr := h.Addrs()[0].String() + "/" + uniqueId

	////fmt.Println("multi address : ")
	////fmt.Println(multiAddressStr)

	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	// Start a stream with the destination.
	// Multiaddress of the destination peer is fetched from the peerstore using 'peerId'.
	s, err := h.NewStream(context.Background(), info.ID, "/validatorpeerlist/1.0.0")
	if err != nil {
		////log.Println(err)
		return nil, err
	}
	////log.Println("Established connection to destination")

	// s2, err := h.NewStream(context.Background(), info.ID, "/peerlist/1.0.0")
	// if err != nil {
	// 	////log.Println(err)
	// 	return nil, err
	// }

	// Create a buffered stream so that read and writes are non blocking.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// rw2 := bufio.NewReadWriter(bufio.NewReader(s2), bufio.NewWriter(s2))

	return rw, nil
}

//--------------------------
//

//-----------------------------------------------------------------
//consessus

//validator pool

func ValidatorPool_runListner(ctx context.Context, ha host.Host, listenPort int, streamHandler network.StreamHandler, streamHandler2 network.StreamHandler) {
	ha.SetStreamHandler("/agent/1.0.0", streamHandler)

	ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
		////log.Println("listener received new stream")
		if err := doEcho1(s); err != nil {
			//log.Println(err)
			s.Reset()
		} else {
			s.Close()
		}
	})
	//fmt.Println("yo")

	ha.SetStreamHandler("/transactions/1.0.0", streamHandler2)

}

func ValidatorPool_HandleStream(s network.Stream) {

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go ReadData6(rw)

}

func ValidatorPool_StartPeerAndConnect(ctx context.Context, h host.Host, destination string) (*bufio.ReadWriter, error) {

	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"ERROR_1: ", time.Now().Format("2006-01-02 15:04:05"), err.Error()})
		return nil, err
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"ERROR_2: ", time.Now().Format("2006-01-02 15:04:05"), err.Error()})
		return nil, err
	}

	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	s, err := h.NewStream(context.Background(), info.ID, "/agent/1.0.0")
	if err != nil {
		filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"ERROR_3: ", time.Now().Format("2006-01-02 15:04:05"), err.Error()})
		return nil, err
	}

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	return rw, nil
}

func WriteData6(rw *bufio.ReadWriter, totalScore_string string) {
	rw.WriteString(fmt.Sprintf("%s`", totalScore_string))

	rw.Flush()
}

// validator pool listner
func ReadData6(rw *bufio.ReadWriter) {

	filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"READ_RECIVER_1", time.Now().Format("2006-01-02 15:04:05")})

	plus := "`"
	bytes_arr := []byte(plus)
	byte_arr := bytes_arr[0]
	system_info, err := rw.ReadString(byte_arr)
	if err != nil {
		filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"ERROR: ", time.Now().Format("2006-01-02 15:04:05"), err.Error()})
	}

	system_info = strings.TrimRight(system_info, "`")

	filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"READ_RECIVER_2", time.Now().Format("2006-01-02 15:04:05"), system_info})

	totalScore, isInValidatorpool := agent.UpdateValidatorPool([]byte(system_info))

	filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"READ_RECIVER_3", time.Now().Format("2006-01-02 15:04:05"), "isInValidatorpool and totala score updated is db as ->", strconv.FormatBool(isInValidatorpool), strconv.FormatFloat(totalScore, 'f', -1, 64)})

}

func WriteData7(rw *bufio.ReadWriter, system_info []byte) {

	filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"5", time.Now().Format("2006-01-02 15:04:05"), "Writing System info"})

	byte_ := []byte("`")
	system_info = append(system_info, byte_[0])

	for i := 0; i < len(system_info); i++ {
		err := rw.WriteByte(system_info[i])
		if err != nil {
			filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"ERROR", time.Now().Format("2006-01-02 15:04:05"), err.Error()})

		}
	}

	rw.Flush()
	filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"6", time.Now().Format("2006-01-02 15:04:05"), "sent!"})

}

func ReadData7(rw *bufio.ReadWriter) {

	for {
		totalScore_string, _ := rw.ReadString('`')

		if totalScore_string == "" {
			return
		}
		if totalScore_string != "`" {
			filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"READ_SENDER_1", time.Now().Format("2006-01-02 15:04:05")})

			totalScore_string = strings.TrimRight(totalScore_string, "`")
			totalScore, _ := strconv.ParseFloat(totalScore_string, 64)

			consensus.IsInValidatorPool = true
			consensus.TotalScore = totalScore

			filemanagement.AddArrayToFile(string(enum.Agent_logFile), []string{"READ_SENDER_2", time.Now().Format("2006-01-02 15:04:05"), "Saving IsInValidatorPool & TotalScore as", strconv.FormatBool(consensus.IsInValidatorPool), strconv.FormatFloat(totalScore, 'f', -1, 64)})

		}
	}
}

// validator

func Validator_runListner(ctx context.Context, ha host.Host, listenPort int, streamHandler network.StreamHandler, streamHandler2 network.StreamHandler) {
	ha.SetStreamHandler("/validation/1.0.0", streamHandler)
	ha.SetStreamHandler("/savetrx/1.0.0", streamHandler2)

}

func Validator_HandleStream(s network.Stream) {
	////log.Println(" 1   Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go ValidatorRead(rw)
	// go ValidatorWrite(rw, )
	// go WriteData6(rw, "")

	// stream 's' will stay open until you close it (or the other side closes it).
}

func Validator_StartPeerAndConnect(h host.Host, destination string) (*bufio.ReadWriter, error) {

	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	s, err := h.NewStream(context.Background(), info.ID, "/validation/1.0.0")
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	// Create a buffered stream so that read and writes are non blocking.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// rw2 := bufio.NewReadWriter(bufio.NewReader(s2), bufio.NewWriter(s2))

	return rw, nil
}

func ValidatorWrite(rw *bufio.ReadWriter, isVerified bool) {

	//fmt.Println("in Vsalidartor wrinte")
	concent := strconv.FormatBool(isVerified)
	//fmt.Println("concent is: ", concent)
	rw.WriteString(fmt.Sprintf("%s`", concent))

	rw.Flush()

}

func ValidatorRead(rw *bufio.ReadWriter) {
	//fmt.Println("in validator read")

	plus := "`"
	bytes_arr := []byte(plus)
	byte_arr := bytes_arr[0]
	bytes, _ := rw.ReadString(byte_arr)

	bytes = strings.TrimRight(bytes, "`")

	//fmt.Println("------------------")
	//fmt.Println(bytes)
	//fmt.Println("---------------------")

	//convert bytes to transactions in string

	separator := "+"
	substrings := strings.Split(string(bytes), separator)
	transactions := make([]string, len(substrings))
	for i, substring := range substrings {
		transactions[i] = substring
	}
	////fmt.Println(transactions)

	//this fuction will
	//read the function
	isVerified := consensus.VerifyTransaactionFromLeader(transactions)

	ValidatorWrite(rw, isVerified)

	transaction.StoreTransactionsInMemory(transactions)

}

func LeaderRead(rw *bufio.ReadWriter) {
	//fmt.Println("in Leader read")

	for {
		plus := "`"
		bytes_arr := []byte(plus)
		byte_arr := bytes_arr[0]
		concent_string, _ := rw.ReadString(byte_arr)

		//fmt.Println("in Leader read for loop")
		if concent_string == "" {
			//fmt.Println("empty string")
			return
		}
		if concent_string != "`" {

			filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"READ_Leader_1", time.Now().Format("2006-01-02 15:04:05")})

			//fmt.Println("got somethng")
			//fmt.Println(concent_string)

			//fmt.Println("------------------------")

			concent_string = strings.TrimRight(concent_string, "`")
			concent, _ := strconv.ParseBool(concent_string)

			filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"READ_Leader_2", time.Now().Format("2006-01-02 15:04:05"), "consent is: ", strconv.FormatBool(concent)})
			if concent {

				consensus.AddVerificationConcent()
				filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"READ_Leader_3", time.Now().Format("2006-01-02 15:04:05"), "consent added, now total number of consents are", strconv.Itoa(consensus.NumberOfConcentToFormBlock)})

				if consensus.CheckVerificationConcent() > consensus.NumberOfConcentToFormBlock {

					filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"READ_Leader_4", time.Now().Format("2006-01-02 15:04:05")})
					//create block
					//fmt.Println("I am Creating Block now!")

					// seperate validator's peer list and address
					involvedPeers, addresses := consensus.SeperatePeerAndAdddressOfValidator()

					//fmt.Println(involvedPeers, addresses)

					//create block
					involvedTransactions, blockNumber := block.CreateBlock(addresses)

					filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"READ_Leader_5", time.Now().Format("2006-01-02 15:04:05"), "BLOCK CREATED"})
					//create block extended
					block.CreateBlockExtended(involvedTransactions, involvedPeers, blockNumber)
					Validator_peerlist := consensus.Validator_peerlist
					consensus.Validator_peerlist = [][]string{}
					filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"READ_Leader_6", time.Now().Format("2006-01-02 15:04:05"), "now, will sent concent to other"})
					sendValidatorsConcentToSaveTransactions(Validator_peerlist)
				}
			}

		}
	}

}

func LeaderWrite(rw *bufio.ReadWriter, transactions []string) {

	filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"10", time.Now().Format("2006-01-02 15:04:05"), "Leader sending transactions to other validator"})
	separator := "+"
	bytes := []byte(strings.Join(transactions, separator))

	byte_ := []byte("`")
	bytes = append(bytes, byte_[0])

	for i := 0; i < len(bytes); i++ {
		rw.WriteByte(bytes[i])
	}

	rw.Flush()

	filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"11", time.Now().Format("2006-01-02 15:04:05"), "sent"})

}

//-----------------

func Validator_SaveTrx_HandleStream(s network.Stream) {
	////log.Println(" 1   Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go ValidatorSaveTrxRead(rw)
	// go ValidatorWrite(rw, )
	// go WriteData6(rw, "")

	// stream 's' will stay open until you close it (or the other side closes it).
}

func Validator_SaveTrx_StartPeerAndConnect(h host.Host, destination string) (*bufio.ReadWriter, error) {

	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	s, err := h.NewStream(context.Background(), info.ID, "/savetrx/1.0.0")
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	// Create a buffered stream so that read and writes are non blocking.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// rw2 := bufio.NewReadWriter(bufio.NewReader(s2), bufio.NewWriter(s2))

	return rw, nil
}

func sendValidatorsConcentToSaveTransactions(validator_peerlist [][]string) {

	filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"READ_Leader_6", time.Now().Format("2006-01-02 15:04:05"), "number of members that are validator: ", strconv.Itoa(len(validator_peerlist))})
	for i := len(validator_peerlist) - 1; i > -1; i-- {
		validator_info := validator_peerlist[i][0]
		var validatorInfo consensus.ValidatorInfo
		json.Unmarshal([]byte(validator_info), &validatorInfo)

		target := validatorInfo.MultiAddr
		if target != Node_multiAddr {
			filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"READ_Leader_6", time.Now().Format("2006-01-02 15:04:05"), "validator: ", validatorInfo.Address})
			rw, err := Validator_SaveTrx_StartPeerAndConnect(Host, target)

			if err == nil {
				filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"READ_Leader_7", time.Now().Format("2006-01-02 15:04:05")})
				LeaderSaveTrxWrite(rw)

			} else {
				filemanagement.AddArrayToFile(string(enum.Consensus_logFile), []string{"ERROR", time.Now().Format("2006-01-02 15:04:05"), err.Error()})
				continue
			}

		}

	}
}

func LeaderSaveTrxWrite(rw *bufio.ReadWriter) {

	bytes := []byte("Ok")
	byte_ := []byte("`")
	bytes = append(bytes, byte_[0])
	for i := 0; i < len(bytes); i++ {
		rw.WriteByte(bytes[i])
	}

	rw.Flush()

}

func ValidatorSaveTrxRead(rw *bufio.ReadWriter) {
	plus := "`"
	bytes_arr := []byte(plus)
	byte_arr := bytes_arr[0]
	bytes, _ := rw.ReadString(byte_arr)

	bytes = strings.TrimRight(bytes, "`")
	inString := string(bytes)

	if inString == "OK" {
		transaction.SaveTransactionsInLDB()
	}

}

//-------------

func ValidatorPool_Trxs_HandleStream(s network.Stream) {

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	//fmt.Println(rw)

	TransactionResponseRead(rw)
}

func Validator_Trxs_StartPeerAndConnect(h host.Host, destination string) (*bufio.ReadWriter, error) {

	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	s, err := h.NewStream(context.Background(), info.ID, "/transactions/1.0.0")
	if err != nil {
		////log.Println(err)
		return nil, err
	}

	// Create a buffered stream so that read and writes are non blocking.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// rw2 := bufio.NewReadWriter(bufio.NewReader(s2), bufio.NewWriter(s2))

	return rw, nil
}

func TransactionRequestWrite(rw *bufio.ReadWriter, blockExtendedNumber int) {

	blockExtendedNumber_str := strconv.Itoa(blockExtendedNumber)

	bytes := []byte(blockExtendedNumber_str)
	byte_ := []byte("`")

	bytes = append(bytes, byte_[0])
	for i := 0; i < len(bytes); i++ {
		rw.WriteByte(bytes[i])
	}

	rw.Flush()
}

func TransactionResponseRead(rw *bufio.ReadWriter) {
	plus := "`"
	bytes_arr := []byte(plus)
	byte_arr := bytes_arr[0]
	bytes, _ := rw.ReadString(byte_arr)

	bytes = strings.TrimRight(bytes, "`")
	blockExtendedNumber_str := string(bytes)

	TransactionResponseWrite(rw, blockExtendedNumber_str)

}

func TransactionRequestRead(rw *bufio.ReadWriter) {

	plus := "`"
	bytes_arr := []byte(plus)
	byte_arr := bytes_arr[0]
	bytes, _ := rw.ReadString(byte_arr)

	bytes = strings.TrimRight(bytes, "`")

	separator := "+/+/+"
	substrings := strings.Split(string(bytes), separator)
	transactionbodies := make([]string, len(substrings))
	for i, substring := range substrings {
		transactionbodies[i] = substring
	}

	block.SaveTransactionsInLDB(transactionbodies)

}

func TransactionResponseWrite(rw *bufio.ReadWriter, blockExtendedNumber_str string) {

	transactionbodies := block.GetTransactionsOfBlockNumber(blockExtendedNumber_str)

	separator := "+/+/+"
	bytes := []byte(strings.Join(transactionbodies, separator))

	byte_ := []byte("`")
	bytes = append(bytes, byte_[0])

	for i := 0; i < len(bytes); i++ {
		rw.WriteByte(bytes[i])
	}
	//fmt.Println("sent")

	rw.Flush()

}
