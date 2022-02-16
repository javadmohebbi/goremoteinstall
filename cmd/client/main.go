package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/javadmohebbi/goremoteinstall/agent"
	"github.com/judwhite/go-svc"
)

// implements svc.Service
type program struct {
	LogFile *os.File

	svr *agent.GriAgent
	ctx context.Context
}

var prg *program

func (p *program) Context() context.Context {
	return p.ctx
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()

	prg = &program{
		svr: &agent.GriAgent{},
		ctx: ctx,
	}

	// time.Sleep(20 * time.Second)

	// call svc.Run to start your program/service
	// svc.Run will call Init, Start, and Stop
	if err := svc.Run(prg); err != nil {
		log.Fatal(err)
	}

}

func (p *program) Init(env svc.Environment) error {
	// log.Printf("is win service? %v\n", env.IsWindowsService())

	// write to "griAgent.log" when running as a Windows Service
	if env.IsWindowsService() {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return err
		}

		conf, err := agent.NewConfig(dir)
		if err != nil {
			return err
		}

		prg.svr.Config = conf
		prg.svr.SvcDir = dir

		logPath := filepath.Join(dir, fmt.Sprintf("griAgent_%v_%v.log",
			conf.TaskID, conf.Time,
		))

		f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		p.LogFile = f

		log.SetOutput(f)

	}

	return nil
}

func (p *program) Start() error {
	log.Printf("Starting griAgent service...\n")
	go p.svr.Start()
	return nil
}

func (p *program) Stop() error {

	log.Printf("Stopping griAgent service...\n")
	if err := p.svr.Stop(); err != nil {
		return err
	}
	log.Printf("The griAgent service Stopped.\n")
	return nil
}

// func main() {
// 	con, err := net.Dial("tcp", "192.168.59.1:9999")
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	defer con.Close()

// 	clientReader := bufio.NewReader(os.Stdin)
// 	serverReader := bufio.NewReader(con)

// 	for {
// 		// Waiting for the client request
// 		clientRequest, err := clientReader.ReadString('\n')

// 		switch err {
// 		case nil:
// 			clientRequest := strings.TrimSpace(clientRequest)
// 			if _, err = con.Write([]byte(clientRequest + "\n")); err != nil {
// 				log.Printf("failed to send the client request: %v\n", err)
// 			}
// 		case io.EOF:
// 			log.Println("client closed the connection")
// 			return
// 		default:
// 			log.Printf("client error: %v\n", err)
// 			return
// 		}

// 		// Waiting for the server response
// 		serverResponse, err := serverReader.ReadString('\n')

// 		switch err {
// 		case nil:
// 			log.Println(strings.TrimSpace(serverResponse))
// 		case io.EOF:
// 			log.Println("server closed the connection")
// 			return
// 		default:
// 			log.Printf("server error: %v\n", err)
// 			return
// 		}
// 	}
// }
