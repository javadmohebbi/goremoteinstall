package deployer

import (
	"log"
	"net"
)

func (d *Deployer) onConnectionTerminate(c net.Conn) {
	log.Println("client closed the connection by terminating the process,", c.RemoteAddr())
}
