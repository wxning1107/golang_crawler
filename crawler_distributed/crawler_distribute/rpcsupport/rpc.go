package rpcsupport

import (
	"net/rpc"
	"net"
	"log"
	"net/rpc/jsonrpc"
)

func ServeRpc(host string, service interface{}) error{
	rpc.Register(service)

	listener, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}

	log.Printf("Listening on %s", host)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("aceept error: %v", err)
			continue
		}
		go jsonrpc.ServeConn(conn)
	}
	return nil
}

func NewClient(host string) (*rpc.Client, error){
	//jsonrpc.Dial()
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	return jsonrpc.NewClient(conn), nil

}
