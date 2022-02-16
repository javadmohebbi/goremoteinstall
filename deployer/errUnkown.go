package deployer

import (
	"log"
	"net"
)

func (d *Deployer) onUnkown(c net.Conn, err error) {
	log.Printf("%v error: %v\n", c.LocalAddr(), err)
}
