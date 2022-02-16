package agent

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/javadmohebbi/goremoteinstall"
)

func (s *GriAgent) ExecStarting() error {

	bootstrapPath := fmt.Sprintf("%s\\%s\\%s",
		os.Getenv("systemroot"),
		s.Config.Dir,
		s.Config.Bootstrap,
	)
	s.bootstrapPath = bootstrapPath

	s.bootstrap = exec.Command(bootstrapPath, s.Config.Params...)

	log.Printf("starting bootstrap: %s %s\n", bootstrapPath, s.Config.Params)

	err := s._rq(goremoteinstall.CMD_BOOTSTRAP_START, s.bootstrapPath)
	if err != nil {
		log.Println("Error:", err)
		s.ExecError(err)
		// os.Exit(1)
		return err
	}

	// // start service
	// if err := s.bootstrap.Start(); err != nil {
	// 	log.Println("error: ", err)
	// 	s.ExecError(err)
	// 	// os.Exit(int(goremoteinstall.ERR_AGENT_START_CMD_ERROR))
	// 	return err
	// }

	go func() {

		//starting cmd
		if err := s.bootstrap.Start(); err != nil {
			log.Printf("starting bootstrap failed: %v\n", err)
			s.err <- err
			return
		}

		log.Printf("bootstrap started: %s %s\n", bootstrapPath, s.Config.Params)

		// started
		err := s._rq(goremoteinstall.CMD_BOOTSTRAP_STARTED, s.bootstrapPath)
		if err != nil {
			s.err <- err
			return
		}

		log.Printf("wait bootstrap (%v) to finish: %s %s\n", s.bootstrap.Process.Pid, bootstrapPath, s.Config.Params)

		// check status
		if err := s.bootstrap.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				// This code copied from: https://stackoverflow.com/questions/10385551/get-exit-code-go/55055100
				// answer by https://stackoverflow.com/users/82219/tux21b
				//
				// The program has exited with an exit code != 0

				// This works on both Unix and Windows. Although package
				// syscall is generally platform dependent, WaitStatus is
				// defined for both Unix and Windows and in both cases has
				// an ExitStatus() method with the same signature.

				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {

					log.Printf("[2]bootstrap (%v) failed: %s %s\n", s.bootstrap.Process.Pid, bootstrapPath, s.Config.Params)

					s.err <- errors.New(fmt.Sprintf("Exit Status: %d", status.ExitStatus()))
					return
				}
			} else {

				log.Printf("[1]bootstrap (%v) failed: %s %s\n", s.bootstrap.Process.Pid, bootstrapPath, s.Config.Params)

				s.err <- err
				return
			}
		}
		// cmd finished ok
		s.done <- true

	}()

	for {
		select {
		case err := <-s.err:
			s.ExecError(err)
		case d := <-s.done:
			if d {
				s.ExecSuccess()
			}
		case n := <-s.task:
			switch n {
			case task_cmd_started:
				// task started
			}

		}
	}

}

// func (s *GriAgent) ExecStarted() {

// }

func (s *GriAgent) ExecSuccess() {

	log.Printf("[4] bootstrap (%v) done\n", s.bootstrap.Process.Pid)

	s._rq(goremoteinstall.CMD_BOOTSTRAP_FINISH_DONE, s.bootstrapPath)
}

func (s *GriAgent) ExecError(err error) {
	msg := fmt.Sprintf("%v, err: %v", s.bootstrap, err)

	log.Printf("[3] bootstrap (%v) failed: %v\n", s.bootstrap.Process.Pid, msg)

	s._rq(goremoteinstall.CMD_BOOTSTRAP_FINISH_ERROR, msg)
}

// func (s *GriAgent) ExecShouldRemove() {
// }

// build request and return string
func (s *GriAgent) _rq(cmd goremoteinstall.Command, descPaylaod string) error {
	rq := goremoteinstall.ClientServerReqResp{
		Agent:       true,
		Command:     cmd,
		HostID:      s.Config.HostID,
		Host:        s.Config.Host,
		RequestID:   fmt.Sprintf("req-%v", time.Now().Unix()),
		DescPayload: descPaylaod,
		TaskID:      s.Config.TaskID,
	}

	bts, err := rq.JSONToStringClientServerReqResp()
	if err != nil {
		return errors.New(fmt.Sprintf("could not marshal to error: %v", err))
	}

	_, err = s.Conn.Write([]byte(fmt.Sprintf("%s\n", bts)))
	if err != nil {
		if err != nil {
			return errors.New(fmt.Sprintf("[%d] could not send tcp request: %v", goremoteinstall.ERR_TCP_CLIENT_AGENT_ERROR, err))
		}
	}

	return err

}
