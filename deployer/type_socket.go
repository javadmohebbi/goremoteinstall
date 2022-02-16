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
	"time"

	"github.com/javadmohebbi/goremoteinstall"
)

type socketClientList struct {
	mu          sync.Mutex
	sockClients []*netSockClient
}

type netSockClient struct {
	Name string
	Conn net.Conn

	Disconnected bool

	IsIdentified      bool
	IsTarget          bool
	IsItSelf          bool
	IsRemoteInstaller bool
}

// identify client
func (ncl *netSockClient) IdentifySockClient(rr goremoteinstall.ClientServerReqResp) error {

	// defer ncl.Conn.Close()

	ncl.IsIdentified = false
	ncl.IsTarget = false
	ncl.IsRemoteInstaller = false
	ncl.IsIdentified = false

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

	// log.Println(rr.Command, ncl.IsIdentified)

	return nil
}

// add network socket client
func (cList *socketClientList) AddSockClient(cl *netSockClient) {
	cList.mu.Lock()
	defer cList.mu.Unlock()

	// if len(cList.sockClients) == 0 {
	// 	cl.IsItSelf = true
	// } else {
	// 	cl.IsRemoteInstaller = true
	// }
	cl.IsIdentified = false

	cList.sockClients = append(cList.sockClients, cl)
}

// network socket handle connections
func (ncl *netSockClient) HandleSockConnection(scl *socketClientList, tcl *tcpClientList) {
	defer ncl.Conn.Close()
	notify := make(chan error)

	go func() {
		// buf := make([]byte, 1024)

		log.Println("new client connected: ", ncl.Conn.LocalAddr().String())
		clientReader := bufio.NewReader(ncl.Conn)

		for {

			clientRequest, err := clientReader.ReadString('\n')
			if err != nil {
				if io.EOF == err {
					ncl.Disconnected = true
				}
				notify <- err
				return
			}

			clientRequest = strings.TrimSpace(clientRequest)

			// log.Println("Req: ", clientRequest)

			var req goremoteinstall.ClientServerReqResp
			err = json.Unmarshal([]byte(clientRequest), &req)
			if err != nil {
				log.Println("u marshal err")
				notify <- err
				return
			}

			// identify if not identified yet
			if !ncl.IsIdentified {
				_ = ncl.IdentifySockClient(req)
			}

			if !ncl.Disconnected {
				// if req.Command == goremoteinstall.CMD_INIT {
				// 	resp = req
				// 	resp.Ack = true
				// 	resp.HostID = fmt.Sprintf("host_%v", time.Now().Unix())

				// 	b, err := json.Marshal(&resp)
				// 	_, err = ncl.Conn.Write([]byte(b))
				// 	if err != nil {

				// 		if io.EOF == err {
				// 			ncl.Disconnected = true
				// 		}

				// 		notify <- err
				// 		return
				// 	}
				// } else {
				// 	ncl.ForwardCommandsToRemoteInstallerServer(
				// 		req, scl,
				// 	)
				// }
				if req.Command != goremoteinstall.CMD_INIT {
					if ncl.IsTarget {
						// if client is an agent, request will be sent to
						// remote installer server
						ncl.ForwardCommandsToRemoteInstallerServer(
							req, scl, tcl,
						)
					}

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
			} else {
				log.Println("err::", err)
			}

		case <-time.After(time.Second * 5):
			countAgent := 0
			countServer := 0
			countSelf := 0
			countNotIdentified := 0
			for _, c := range scl.sockClients {
				if !c.Disconnected {
					if !c.IsIdentified {
						countNotIdentified++
					}
					if c.IsRemoteInstaller {
						countServer++
					}
					if c.IsItSelf {
						countSelf++
					}
					if c.IsTarget {
						countAgent++
					}
				}
			}
			for _, c := range tcl.tcpClients {
				if !c.Disconnected {
					if !c.IsIdentified {
						countNotIdentified++
					}
					if c.IsRemoteInstaller {
						countServer++
					}
					if c.IsItSelf {
						countSelf++
					}
					if c.IsTarget {
						countAgent++
					}
				}
			}
			fmt.Println("s:", countServer, "it:", countSelf, "ag:", countAgent, "na:", countNotIdentified)
			// 	for _, c := range scl.sockClients {
			// 		if c.IsIdentified && c.IsRemoteInstaller && !c.Disconnected {
			// 			rr := goremoteinstall.ClientServerReqResp{
			// 				Command:   goremoteinstall.CMD_TEST,
			// 				RequestID: "mmamado",
			// 			}
			// 			b, err := json.Marshal(&rr)
			// 			if err == nil {
			// 				_, err := c.Conn.Write(b)
			// 				if err != nil {
			// 					log.Println("echo error", err)
			// 				}
			// 			}
			// 		}
			// 	}
		}
	}

}

func (ncl *netSockClient) ForwardCommandsToRemoteInstallerServer(
	rr goremoteinstall.ClientServerReqResp,
	scl *socketClientList,
	tcl *tcpClientList,
) {

	if rr.RequestID == "" {
		log.Println("invalid request. no reqId is in the request")
		return
	}

	shouldEcho := true

	switch rr.Command {
	case goremoteinstall.CMD_BOOTSTRAP_START:
		log.Println("starting")

	case goremoteinstall.CMD_BOOTSTRAP_STARTED:
		log.Println("started")

	case goremoteinstall.CMD_BOOTSTRAP_FINISH_ERROR:
		log.Println("finish")

	case goremoteinstall.CMD_BOOTSTRAP_FINISH_DONE:
		log.Println("finish done")

	default:
		shouldEcho = false
		log.Println("unkown")
	}

	if shouldEcho {
		for _, c := range scl.sockClients {
			// forward original message to servers
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

		// in case remote installer listen on tcp
		for _, c := range tcl.tcpClients {
			// forward original message to servers
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
}
