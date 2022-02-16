package deployer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/javadmohebbi/goremoteinstall"
)

type tcpClientList struct {
	mu         sync.Mutex
	tcpClients []*netTCPClient
}

type netTCPClient struct {
	Name string
	Conn net.Conn

	Disconnected bool

	IsIdentified      bool
	IsTarget          bool
	IsItSelf          bool
	IsRemoteInstaller bool

	UnixClientSock *net.Conn
}

// identify client
func (ncl *netTCPClient) IdentifyTCPClient(rr goremoteinstall.ClientServerReqResp) error {
	// defer ncl.Conn.Close()
	ncl.IsIdentified = false
	ncl.IsTarget = false
	ncl.IsRemoteInstaller = false
	ncl.IsIdentified = false

	log.Println("rrr", rr.Agent)
	log.Println("rr cmd", rr.Command)

	switch rr.Command {
	case goremoteinstall.CMD_INIT:
		if rr.ItSelf {
			ncl.IsItSelf = true
			ncl.IsIdentified = true
		} else if rr.Agent {
			ncl.IsTarget = true
			ncl.IsIdentified = true
		} else if rr.Server {
			ncl.IsRemoteInstaller = true
			ncl.IsIdentified = true
		} else {
			ncl.IsIdentified = false
		}
	default:
		ncl.IsIdentified = false
	}

	return nil
}

// add network tcp client
func (cList *tcpClientList) AddTCPClient(cl *netTCPClient) {
	cList.mu.Lock()
	defer cList.mu.Unlock()

	cl.IsIdentified = false

	cList.tcpClients = append(cList.tcpClients, cl)
}

// network tcp handle connections
func (ncl *netTCPClient) HandleTCPConnection(tcl *tcpClientList, scl *socketClientList, unixClient *net.Conn) {
	defer ncl.Conn.Close()
	notify := make(chan error)

	go func() {
		// buf := make([]byte, 1024)

		log.Println("new client connected: ", ncl.Conn.LocalAddr().String())
		clientReader := bufio.NewReader(ncl.Conn)

		for {

			clientRequest, err := clientReader.ReadString('\n')
			if err != nil {
				notify <- err
				return
			}

			clientRequest = strings.TrimSpace(clientRequest)

			// log.Println("Req: ", clientRequest, "?", ncl.IsIdentified, ncl.IsTarget)

			var req goremoteinstall.ClientServerReqResp
			err = json.Unmarshal([]byte(clientRequest), &req)
			if err != nil {
				log.Println("u marshal err")
				notify <- err
				return
			}

			// identify if not identified yet
			if !ncl.IsIdentified {
				_ = ncl.IdentifyTCPClient(req)
			}

			if !ncl.Disconnected {
				if req.Command != goremoteinstall.CMD_INIT {
					// forward request to unix socket
					ncl.ForwardToUnixSocketClient(
						req, tcl, scl, unixClient,
					)
				}
			}
		}
	}()
	for {
		select {
		case err := <-notify:
			if io.EOF == err {
				ncl.Disconnected = true
				log.Println("connection dropped message", err)
				return
			}
			// case <-time.After(time.Second * 1):
			// 	log.Println("timeout 1, still alive")
		}
	}

}

func (ncl *netTCPClient) ForwardToUnixSocketClient(
	rr goremoteinstall.ClientServerReqResp,
	tcl *tcpClientList,
	scl *socketClientList,
	conn *net.Conn,
) {
	if rr.RequestID == "" {
		log.Println("invalid request. no reqId is in the request")
		return
	}

	for _, c := range scl.sockClients {
		// only send to remote installers
		// which are identified and connected
		if c.IsRemoteInstaller && !c.Disconnected {
			b, err := json.Marshal(&rr)
			fw := fmt.Sprintf("%s\n", string(b))
			if err == nil {
				_, err := c.Conn.Write([]byte(fw))
				if err != nil {
					log.Println("echo error", err)
				}
			}
		}
	}

	// if remote installer connected via tcp not unix socket
	for _, c := range tcl.tcpClients {
		// only send to remote installers
		// which are identified and connected
		if c.IsRemoteInstaller && !c.Disconnected {
			b, err := json.Marshal(&rr)
			fw := fmt.Sprintf("%s\n", string(b))
			if err == nil {
				_, err := c.Conn.Write([]byte(fw))
				if err != nil {
					log.Println("echo error", err)
				}
			}
		}
	}

}
