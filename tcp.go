package goremoteinstall

// func (gri *RemoteInstall) Listen() {

// 	// close connection when done
// 	defer gri.l.Close()

// 	var counter = 0

// 	for {
// 		conn, err := gri.l.Accept()
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}

// 		counter += 1
// 		go gri.handleClientRequest(counter, conn)
// 	}

// }

// func (gri *RemoteInstall) handleClientRequest(counter int, con net.Conn) {
// 	defer con.Close()

// 	clientReader := bufio.NewReader(con)

// 	for {
// 		// Waiting for the client request
// 		clientRequest, err := clientReader.ReadString('\n')

// 		switch err {
// 		case nil:
// 			clientRequest := strings.TrimSpace(clientRequest)
// 			if clientRequest == ":QUIT" {
// 				log.Println("client requested server to close the connection so closing")
// 				return
// 			} else {
// 				log.Println(clientRequest)
// 			}
// 		case io.EOF:
// 			log.Println("client closed the connection by terminating the process")
// 			return
// 		default:
// 			log.Printf("error: %v\n", err)
// 			return
// 		}

// 		// Responding to the client request
// 		if _, err = con.Write([]byte("GOT IT!\n")); err != nil {
// 			log.Printf("failed to respond to client: %v\n", err)
// 		}
// 	}
// }
