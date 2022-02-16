package deployer

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
)

func (d *Deployer) forwardRequestHandler(c net.Conn) {

	defer c.Close()

	log.Printf("Client connected [%s]", c.RemoteAddr().Network())
	// io.Copy(c, c)
	clientReader := bufio.NewReader(c)

	for {

		// Waiting for the client request
		clientRequest, err := clientReader.ReadString('\n')
		switch err {
		case nil:

			clientRequest := strings.TrimSpace(clientRequest)

			// echo response to unix socket client
			_, err := c.Write([]byte(clientRequest))
			if err != nil {
				log.Fatalln("write error:", err)
			}
		case io.EOF:
			d.onConnectionTerminate(c)
			return
		default:
			d.onUnkown(c, err)
			return
		}

	}

}
