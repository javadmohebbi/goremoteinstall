package deployer

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/javadmohebbi/goremoteinstall"
)

func (d *Deployer) ListenTCP() {
	log.Println("Deployer is listening on tcp:", d.Listen)
	l, err := net.Listen("tcp", d.Listen)
	if err != nil {
		log.Fatal(err)
	}
	d.Listener = l

	d.ch = make(chan os.Signal, 1)
	signal.Notify(d.ch,
		syscall.SIGINT,  // "the normal way to politely ask a program to terminate"
		syscall.SIGTERM, // Ctrl+C
		syscall.SIGHUP,  // "terminal is disconnected"
	)

	// handle signals
	go func() {
		<-d.ch

		//
		log.Println("CTRL + C recvd")

		// do not accept new connections
		close(d.closeSignalRecvdTcp)

		// close socket listener
		d.Close()
		os.Exit(0)
	}()

	if d.tcpClientList == nil {
		d.tcpClientList = &tcpClientList{}
	}

	unixClient, err := net.Dial("unix", d.SocketAddr)
	if err != nil {
		log.Fatalln(err)
	}
	defer unixClient.Close()

	rq := goremoteinstall.ClientServerReqResp{
		ItSelf:  true,
		Command: goremoteinstall.CMD_INIT,
		Host:    "itself",
	}
	bts, err := json.Marshal(&rq)
	if err != nil {
		log.Fatalln(err)
	}

	// initialize socket client
	_, err = unixClient.Write([]byte(fmt.Sprintf("%s\n", bts)))
	if err != nil {
		log.Fatalln(err)
	}

	// handle requests
	for {
		conn, err := d.Listener.Accept()
		if err != nil {
			select {
			case <-d.closeSignalRecvdTcp:
				return
			default:
				log.Println("err in connection: ", d.ListenerSocket.Addr().String())
				continue
			}
		} else {

			go func() {
				_c := &netTCPClient{
					Conn:           conn,
					UnixClientSock: &unixClient,
				}

				d.tcpClientList.AddTCPClient(_c)

				_c.HandleTCPConnection(d.tcpClientList, d.socketClientList, &unixClient)

			}()

		}
	}

}
