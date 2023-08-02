package jumborpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"jumbochain.org/block"
)

type RPC_methods struct{}

type Args struct {
	A, B int
}

type Reply1 struct {
	Result1 string
}

func (a *RPC_methods) CurrentBlockNumber(_ *Args, reply *Reply1) {

	currentBlockNumber := block.GetCurrentBlockNumber()
	fmt.Println("current block number is :")
	fmt.Println(currentBlockNumber)
	reply.Result1 = currentBlockNumber
}

func RunRpcServer() {
	rpcMethods := new(RPC_methods)
	rpc.Register(rpcMethods)

	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Error listening:", err)
	}

	log.Println("RPC server is listening on port 1234")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting connection:", err)
		}
		go rpc.ServeConn(conn)
	}

}

func RunRpcClient() {

}
