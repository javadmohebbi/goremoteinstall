package deployer

import (
	"net"
	"os"
	"sync"
)

type Deployer struct {
	SocketAddr string
	Listen     string // address and port eg: ip:port, fqdn:port

	Listener       net.Listener
	ListenerSocket net.Listener

	ch                  chan os.Signal
	closeSignalRecvd    chan interface{} // if true, it will not accept new connections
	closeSignalRecvdTcp chan interface{} // if true, it will not accept new connections

	Count  uint
	ItSelf string

	socketClientList *socketClientList
	tcpClientList    *tcpClientList

	// wait group
	waitGroup *sync.WaitGroup
}

func New(socketAddr, listen string) *Deployer {
	return &Deployer{
		SocketAddr:          socketAddr,
		waitGroup:           &sync.WaitGroup{},
		Listen:              listen,
		closeSignalRecvd:    make(chan interface{}),
		closeSignalRecvdTcp: make(chan interface{}),
	}
}

type Client struct {
	ID string
}
