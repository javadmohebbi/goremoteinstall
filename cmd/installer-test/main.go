package main

func main() {

	// shouldWait := make(chan bool)
	// _err := make(chan error)
	// var ticker *time.Ticker

	// var pid int
	// var cids []int

	// cmd := exec.Command(
	// 	"C:\\Windows\\Temp\\YARMA\\tsk-epskit_x64_6.6.20.294_2022-02-14_13-10-48\\epskit_x64_6.6.20.294.exe",
	// 	"/bdparams", "/silent", "/autocleanup",
	// )

	// go func() {
	// 	if err := cmd.Start(); err != nil {
	// 		log.Printf("starting bootstrap failed: %v\n", err)
	// 		_err <- err
	// 		return
	// 	}

	// 	pid = cmd.Process.Pid

	// 	ticker = time.NewTicker(100 * time.Millisecond)

	// 	// if err := cmd.Wait(); err != nil {
	// 	// 	if exitErr, ok := err.(*exec.ExitError); ok {
	// 	// 		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
	// 	// 			log.Println("exit staus: ", status.ExitStatus())
	// 	// 			_err <- err
	// 	// 			return
	// 	// 		}
	// 	// 	} else {
	// 	// 		_err <- err
	// 	// 		return
	// 	// 	}
	// 	// }

	// }()

	// for {
	// 	select {
	// 	case e := <-_err:
	// 		log.Fatalln(e)
	// 	case sw := <-shouldWait:
	// 		if !sw {
	// 			log.Println("done")
	// 			os.Exit(0)
	// 		}
	// 	case <-ticker.C:
	// 		// log.Println("some processes are need to be finished!")

	// 	}
	// }

}
