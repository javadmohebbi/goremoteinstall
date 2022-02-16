package agent

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/javadmohebbi/goremoteinstall"
)

type task_type int

const (
	task_cmd_starting task_type = iota + 1
	task_cmd_started
	task_cmd_completed_success
	task_cmd_completed_error

	task_service_should_remove
	task_done
)

type GriAgent struct {
	task chan task_type

	err  chan error
	done chan bool

	Conn net.Conn

	Config *GriConfig

	bootstrap     *exec.Cmd
	bootstrapPath string

	exit chan struct{}
	wg   sync.WaitGroup

	SvcDir string
}

func (s *GriAgent) Start() {
	s.task = make(chan task_type)
	s.exit = make(chan struct{})
	s.err = make(chan error)
	s.done = make(chan bool)

	deployer := fmt.Sprintf("%s:%d",
		s.Config.DeployerAddress,
		s.Config.DeployerPort,
	)

	log.Printf("dialing tcp on %s\n", deployer)

	//
	con, err := net.Dial("tcp", deployer)
	if err != nil {
		log.Println("could not connect to tcp: ", deployer)
		os.Exit(int(goremoteinstall.ERR_UNIX_CLIENT_SOCKET))
	}
	rq := goremoteinstall.ClientServerReqResp{
		Agent:     true,
		Command:   goremoteinstall.CMD_INIT,
		HostID:    s.Config.HostID,
		Host:      s.Config.Host,
		RequestID: fmt.Sprintf("req-%v", time.Now().Unix()),
		TaskID:    s.Config.TaskID,
	}

	log.Printf("initializing tcp connection with init command on %s\n", deployer)

	bts, err := rq.JSONToStringClientServerReqResp()
	if err != nil {
		log.Println("could not marshal to json: ", err)
		os.Exit(int(goremoteinstall.ERR_UNIX_CLIENT_SOCKET_MARSHAL))
	}

	// initialize socket client
	_, err = con.Write([]byte(fmt.Sprintf("%s\n", bts)))
	if err != nil {
		if err != nil {
			fmt.Println("could not initialize tcp client: ", err)
			os.Exit(int(goremoteinstall.ERR_UNIX_CLIENT_SOCKET_INIT))
		}
	}
	s.Conn = con
	go s.handleTcpReqests()

	s.wg.Add(1)
	// s.task <- task_type(1)
	go s.DoTheJob()

	// always status is starting

	// s.wg.Add(2)
	// go s.StartSender()
	// go s.StartReceiver()
}

func (s *GriAgent) Stop() error {
	close(s.exit)

	s.Conn.Close()

	s.wg.Wait()
	return nil
}

func (s *GriAgent) DoTheJob() {
	// whenever service starts
	// task will change to starting
	// s.task <- task_cmd_starting

	// defer s.wg.Done()
	// for {
	// 	select {
	// 	case n := <-s.task:
	// 		// if task changed
	// 		switch n {
	// 		case task_cmd_starting:
	// 			log.Println("starting")

	// 			// case task_cmd_started:
	// 			// 	log.Println("started")

	// 			// s.task <- task_cmd_completed_success

	// 			// case task_cmd_completed_success:
	// 			// 	log.Println("success")
	// 			// 	s.task <- task_service_should_remove
	// 			// case task_cmd_completed_error:
	// 			// 	log.Println("error")
	// 			// 	s.task <- task_service_should_remove
	// 			// case task_service_should_remove:
	// 			// 	log.Println("should remove")
	// 			// 	s.task <- task_done
	// 			// case task_done:
	// 			// 	log.Println("done")
	// 			// 	// stop task
	// 			// 	s.Stop()
	// 		}
	// 	case <-s.exit:
	// 		// when service stopped
	// 		return
	// 	default:

	// 	}

	// }

	// starting exec

	_ = s.ExecStarting()

}

func (s *GriAgent) handleTcpReqests() {
	clientReader := bufio.NewReader(s.Conn)
	for {

		clientRequest, err := clientReader.ReadString('\n')

		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			log.Println("SrvResp: ", string(clientRequest))
		case io.EOF:
			s.Stop()
			log.Println("client closed the connection")
			return
		default:
			s.Stop()
			log.Printf("client error: %v\n", err)
			return
		}
	}
}

// func (s *GriAgent) StartSender() {
// 	ticker := time.NewTicker(20 * time.Millisecond)
// 	defer s.wg.Done()
// 	count := 1
// 	for {
// 		select {
// 		case <-ticker.C:
// 			select {
// 			case s.data <- count:
// 				count++
// 			case <-s.exit:
// 				// if the other goroutine exits there'll be no one to receive on the data chan,
// 				// and this goroutine could block. you can simulate this by putting a time.Sleep
// 				// inside startReceiver's receive on s.data and a log.Println here
// 				return
// 			}
// 		case <-s.exit:
// 			return
// 		}
// 	}
// }

// func (s *GriAgent) StartReceiver() {
// 	defer s.wg.Done()
// 	for {
// 		select {
// 		case n := <-s.data:
// 			log.Printf("%d\n", n)
// 		case <-s.exit:
// 			return
// 		}
// 	}
// }
