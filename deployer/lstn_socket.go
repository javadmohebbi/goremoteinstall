package deployer

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// listen on unix socket for
// the local app communications
func (d *Deployer) ListenSocket() {

	// remove current socket
	if err := os.RemoveAll(d.SocketAddr); err != nil {
		log.Fatalln(err)
	}

	log.Println("Deployer is listening on socket: ", d.SocketAddr)

	ls, err := net.Listen("unix", d.SocketAddr)
	if err != nil {
		log.Fatalln(err)
	}
	d.ListenerSocket = ls

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

		close(d.closeSignalRecvd)

		// close socket listener
		d.Close()
		os.Exit(0)
	}()

	if d.socketClientList == nil {
		d.socketClientList = &socketClientList{}
	}

	// handle requests
	for {

		conn, err := d.ListenerSocket.Accept()
		if err != nil {
			select {
			case <-d.closeSignalRecvd:
				return
			default:
				log.Println("err in connection: ", d.ListenerSocket.Addr().String())
				continue
			}
		} else {
			// handle request
			go func() {
				_c := &netSockClient{
					Conn: conn,
				}
				// _c.IdentifySockClient()
				d.socketClientList.AddSockClient(_c)

				// for _, sc := range d.socketClientList.sockClients {
				// 	fmt.Printf("ident:%v, agent:%v, server:%v, itself:%v\n",
				// 		sc.IsIdentified, sc.IsTarget, sc.IsRemoteInstaller, sc.IsItSelf,
				// 	)
				// }

				_c.HandleSockConnection(d.socketClientList, d.tcpClientList)
			}()

		}

	}

}
