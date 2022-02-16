package goremoteinstall

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hirochachacha/go-smb2"
	"github.com/zenthangplus/goccm"
)

func (gri *RemoteInstall) Run() {

	gri.startTime = time.Now()

	// // read targets from target.txt
	// gri.readTargets()

	// // read f2c from f2c.txt
	// gri.readFilesToCopy()

	// gri.copyFilesOverSMB()

	// // listen on tcp and wait for installer
	// // to update it's status
	// gri.Listen()

	gri.closeChannel = make(chan os.Signal, 1)
	signal.Notify(gri.closeChannel,
		syscall.SIGINT,  // "the normal way to politely ask a program to terminate"
		syscall.SIGTERM, // Ctrl+C
		syscall.SIGHUP,  // "terminal is disconnected"
	)

	// handle signals
	go func() {
		<-gri.closeChannel

		//
		log.Println("CTRL + C recvd")

		// stop jobs
		gri.closeSignalRecvd = true

		// close socket listener
		gri.close()
		os.Exit(0)
	}()

	gri.readTargets()

	// prepare remote agent conf
	gri.agent_file_config.Bootstrap = gri.srvTaskConfig.Bootstrap
	gri.agent_file_config.Params = gri.srvTaskConfig.Params
	gri.agent_file_config.SrvAddress = gri.srvGlobalConfig.DeployerAddress
	gri.agent_file_config.SrvPort = gri.srvGlobalConfig.DeployerPort
	gri.agent_file_config.TaskID = gri.TaskID

	// do the job with the limited concurrent goroutines
	go gri.doIt()

	// wait
	select {}
}

func (gri *RemoteInstall) doIt() {

	c := goccm.New(gri.srvTaskConfig.Concurrent)

	if !gri.silent {
		gri.debug = false

		// prepare pretty output
		gri.cliProgress()
	}

	for ind, t := range gri.targets {
		// This function have to call before any goroutine
		c.Wait()

		go func(i int, t *Target) {
			tm := &TaskModel{
				Host: t.Host,
				// HostID: t.HostID,
			}
			// only for network-based updates
			t._tm = tm

			errTsk := gri.doTheTaskForEachTarget(t, tm)
			if errTsk != nil {

				tm.InserStatus(t, STATUS_FINISHED_ERR, errTsk.Error(), gri.db)
			}

			// This function have to when a goroutine has finished
			// Or you can use `defer c.Done()` at the top of goroutine.
			c.Done()
		}(ind, t)

	}
	// This function have to call to ensure all goroutines have finished
	// after close the main program.
	c.WaitAllDone()
	gri.close()

}

func (gri *RemoteInstall) close() {

	if !gri.debug {
		// clear screen
		gri._cliProgress_clear()

		// print global info
		gri._cliPrgress_print_globalInfo(true)
	}

	// close unix client
	_cc := *gri.UnixClientSock
	_cc.Close()

	os.Exit(0)

}

func (gri *RemoteInstall) _init() {

	// read targets and add it to
	// grti.targets array
	gri.readTargets()

}

// this function will handle
// from agents (this requests are forwarded from Deployer socket)
func (gri *RemoteInstall) handleClientConnections() {

	clientReader := bufio.NewReader(*gri.UnixClientSock)

	for {

		clientRequest, err := clientReader.ReadString('\n')

		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			// if _, err = gri.UnixClientSock.Write([]byte(clientRequest + "\n")); err != nil {
			// 	log.Printf("failed to send the client request: %v\n", err)
			// }
			// log.Println("==", string(clientRequest))
			rr, err := StringToJSONClientServerReqResp(clientRequest)
			if err != nil {
				log.Println(err)
				return
			}
			// change status
			gri.parseAgentReqAndStore(rr)

		case io.EOF:
			log.Println("client closed the connection")
			gri.debug = true
			gri.close()
			return
		default:
			log.Printf("client error: %v\n", err)
			gri.debug = true
			gri.close()
			return
		}
	}
}

func (gri *RemoteInstall) doTheTaskForEachTarget(t *Target, tm *TaskModel) error {

	var err error

	// set host task is in progress
	t.IsInProgress = true

	// Starting
	gri._tsk_1(t, tm)

	// SMB Check
	err = gri._tsk_2_dialSMB(t, tm)
	if err != nil {
		// set host task is in progress to false
		t.IsInProgress = false
		t.HasError = true
		return err
	}

	// SMB prepare
	err = gri._tsk_3_prepareSMBSession(t, tm)
	if err != nil {
		// set host task is in progress to false
		t.IsInProgress = false
		t.HasError = true
		return err
	}

	// SMB share list
	err = gri._tsk_4_listSMBShares(t, tm)
	if err != nil {
		// set host task is in progress to false
		t.IsInProgress = false
		t.HasError = true
		return err
	}

	// close smsession for the target machine
	defer t.smb2Session.Logoff()

	// Copy files over SMB
	err = gri._tsk_5_copyFilesOverSMB(t, tm)
	if err != nil {
		t.IsInProgress = false
		t.HasError = true
		return err
	}

	// Service started
	err = gri._tsk_6_createAndStartService(t, tm)
	if err != nil {
		t.IsInProgress = false
		t.HasError = true
		return err
	}

	/// last status IF DONE
	// gri._tsk_last_done(t, tm)

	select {}

	// return nil

}

// last status and task
func (gri *RemoteInstall) _tsk_last_done(t *Target, tm *TaskModel) {

	t.IsDone = true
	t.IsInProgress = false
	t.HasError = false

	if gri.debug {
		log.Printf("Task is completed on '%s'\n", t.Host)
	}

	// write status
	_ = tm.InserStatus(t, STATUS_FINISHED_OK, "", gri.db)

	gri.close()

}

func (gri *RemoteInstall) _tsk_6_createAndStartService(t *Target, tm *TaskModel) error {
	if gri.debug {
		log.Printf("Creating service on '%s'\n", t.Host)
	}

	// write status
	_ = tm.InserStatus(t, STATUS_CREATING_SERVICE, "", gri.db)

	// create and stop service
	err := gri._createService(t)
	if err != nil {
		// write status
		_ = tm.InserStatus(t, STATUS_SERVICE_CREATION_FAILED, err.Error(), gri.db)
		return err
	}
	gri._stopService(t)

	// write status
	_ = tm.InserStatus(t, STATUS_SERVICE_CREATED, "", gri.db)

	if gri.debug {
		log.Printf("Starting service on '%s'\n", t.Host)
	}
	// write status
	_ = tm.InserStatus(t, STATUS_STARTING_SERVICE, "", gri.db)
	gri._startService(t)

	return nil
}

// this task will read the files in the directory (files: in task.yaml) for the
// and will copy them to the destination machine
// It also copy the service installer.exe (agentPath: in griServer.yml )
func (gri *RemoteInstall) _tsk_5_copyFilesOverSMB(t *Target, tm *TaskModel) error {
	if gri.debug {
		log.Printf("Copy files over SMB on '%s'\n", t.Host)
	}

	// write status
	tm.InserStatus(t, STATUS_SMB_COPY, "", gri.db)

	shareName := "ADMIN$"
	_agent_path := "TEMP\\YARMA\\"
	_path := fmt.Sprintf("TEMP\\YARMA\\%v_%v", gri.TaskID, gri.t.Format("2006-01-02_15-04-05"))
	gri.agent_file_config.Dir = _path
	t.dirPath = _path

	// mount smb ADMIN$
	fs, err := t.smb2Session.Mount(shareName)
	if err != nil {
		// write status
		tm.InserStatus(t, STATUS_SMB_MOUNT_ADMIN_FAILED, err.Error(), gri.db)
		return err
	}
	defer fs.Umount()

	// create dir
	err = fs.MkdirAll(_path, 0777)
	if err != nil {
		// write status
		tm.InserStatus(t, STATUS_SMB_MKDIR_IN_ADMIN_FAILED, err.Error(), gri.db)
		return err
	}

	// BEGIN. copy agent to the target machine

	// write status
	tm.InserStatus(t, STATUS_SMB_COPY_AGENT, "", gri.db)
	err = gri._tsk_copy_agent(_agent_path, fs, t)
	if err != nil {
		// write status
		tm.InserStatus(t, STATUS_SMB_COPY_AGENT_FAILED, err.Error(), gri.db)
		return err
	}

	// write status
	tm.InserStatus(t, STATUS_SMB_COPY_AGENT_DONE, "", gri.db)

	// E N D. copy agent to the target machine

	// BEGIN . COPY PACKAGE FILE TO TARGET MACHINE
	err = gri._tsk_read_package_file(t, tm)
	if err != nil {
		return err
	}

	for _, pkgf := range gri.package_files_bytes {
		dst := fmt.Sprintf("%v\\%v",
			_path,
			pkgf.filename,
		)

		// write status
		tm.InserStatus(t, STATUS_SMB_COPY_PKG_FILE_TO_TARGET, pkgf.filename, gri.db)

		err := fs.WriteFile(
			dst,
			pkgf.file_bytes,
			0777,
		)

		if err != nil {
			// write status
			tm.InserStatus(t, STATUS_SMB_COPY_PKG_FILE_TO_TARGET_FAILED,
				fmt.Sprintf("%v: %v", pkgf.filename, err.Error()),
				gri.db,
			)
			return err
		}

		// write status
		tm.InserStatus(t, STATUS_SMB_COPY_PKG_FILE_TO_TARGET_DONE, pkgf.filename, gri.db)
	}

	if err != nil {
		return err
	}
	// E N D . COPY PACKAGE FILE TO TARGET MACHINE

	return nil
}

func (gri *RemoteInstall) _tsk_4_listSMBShares(t *Target, tm *TaskModel) error {
	if gri.debug {
		log.Printf("Listing SMB shares on '%s'\n", t.Host)
	}

	// write status
	_ = tm.InserStatus(t, STATUS_PREPARE_SMB_SESSION_LIST_SHARES, "", gri.db)

	names, err := t.smb2Session.ListSharenames()
	if err != nil {
		log.Printf("Could not list shares using username '%v' on host '%v' due to error: %v", gri.Username, t.Host, err)
		// write status
		_ = tm.InserStatus(t, STATUS_PREPARE_SMB_SESSION_LIST_SHARES_FAILED, err.Error(), gri.db)
		return err
	}

	if len(names) > 0 {
		nm := ""
		hasAdminShare := false
		for _, name := range names {
			nm += name + ", "
			if strings.ToLower(name) == "admin$" {
				hasAdminShare = true
			}
		}

		// write status
		_ = tm.InserStatus(t, STATUS_PREPARE_SMB_SESSION_LIST_SHARES_OK, nm, gri.db)

		if !hasAdminShare {
			// NO ADMIN$
			// write status
			_ = tm.InserStatus(t, STATUS_PREPARE_SMB_SESSION_LIST_SHARES_NOADMIN, "", gri.db)
			return errors.New(fmt.Sprintf("%s", STATUS_PREPARE_SMB_SESSION_LIST_SHARES_NOADMIN.String()))
		} else {
			//  ADMIN$ IS OK
			// write status
			_ = tm.InserStatus(t, STATUS_PREPARE_SMB_SESSION_LIST_SHARES_ADMINOK, "", gri.db)
		}

	} else {
		// write status
		_ = tm.InserStatus(t, STATUS_PREPARE_SMB_SESSION_LIST_SHARES_NOADMIN, "", gri.db)
	}

	return nil
}

func (gri *RemoteInstall) _tsk_3_prepareSMBSession(t *Target, tm *TaskModel) error {
	if gri.debug {
		log.Printf("Preparing SMB Sessions on '%s'\n", t.Host)
	}

	// write status
	tm.InserStatus(t, STATUS_PREPARE_SMB_SESSION, "", gri.db)

	// prepare smb config
	t.smb2Dialer = &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     gri.Username,
			Password: gri.Password,
			Domain:   gri.Domain,
		},
	}

	// connect to smb
	s, err := t.smb2Dialer.Dial(t.Conn)
	if err != nil {
		log.Printf("Can not prepare session using username '%v' on host '%v' due to error: %v", gri.Username, t.Host, err)
		_ = tm.InserStatus(t, STATUS_PREPARE_SMB_SESSION_FAILED, err.Error(), gri.db)
		return err
	}

	t.smb2Session = s

	// write status
	_ = tm.InserStatus(t, STATUS_PREPARE_SMB_SESSION_DONE, "", gri.db)

	return nil
}

func (gri *RemoteInstall) _tsk_2_dialSMB(t *Target, tm *TaskModel) error {
	if gri.debug {
		log.Printf("Dial TCP 445 on '%s'\n", t.Host)
	}

	// write status
	tm.InserStatus(t, STATUS_DIAL_445, "", gri.db)

	// dial tcp 445
	err := gri._smbDial445(t)
	if err != nil {
		// write status
		_ = tm.InserStatus(t, STATUS_DIAL_445_FAILED, err.Error(), gri.db)
		return err
	}

	// write status
	_ = tm.InserStatus(t, STATUS_DIAL_445_OK, "", gri.db)

	return nil
}

func (gri *RemoteInstall) _tsk_1(t *Target, tm *TaskModel) {
	if gri.debug {
		log.Printf("Task '%s' is starting on '%s'\n", gri.TaskID, t.Host)
	}

	// write status
	_ = tm.InserStatus(t, STATUS_STARTING, "", gri.db)

}

// copy agent service to the target machine
func (gri *RemoteInstall) _tsk_copy_agent(_path string, fs *smb2.Share, t *Target) error {
	dst := fmt.Sprintf("%v\\%v",
		_path,
		filepath.Base(gri.agentPath),
	)

	dstJson := fmt.Sprintf("%v\\%v.json",
		_path,
		"griAgent",
	)

	var err error
	var b []byte

	//
	//
	// BEGIN AGENT EXE

	if !gri.agent_files_bytes_done {
		b, err = ioutil.ReadFile(gri.agentPath)
		if err != nil {
			return err
		}

		gri.agent_files_bytes = b

	} else {
		b = gri.agent_files_bytes
	}

	err = fs.WriteFile(
		dst,
		b,
		0777,
	)

	if err != nil {
		return err
	}

	// E N D AGENT EXE
	//
	//

	//
	//
	// BEGIN AGENT JSON CONFIG

	tmpConf := Agent_file_config_type{
		Bootstrap:  gri.agent_file_config.Bootstrap,
		Params:     gri.agent_file_config.Params,
		SrvAddress: gri.agent_file_config.SrvAddress,
		SrvPort:    gri.agent_file_config.SrvPort,
		TaskID:     gri.agent_file_config.TaskID,
		Dir:        gri.agent_file_config.Dir,
		HostID:     t.HostID,
		Host:       t.Host,
		Time:       gri.t.Format("2006-01-02_15-04-05"),
	}
	b, err = json.Marshal(&tmpConf)
	if err != nil {
		return errors.New("can not create JSON conf for remote install agent")
	}

	err = fs.WriteFile(
		dstJson,
		b,
		0777,
	)

	if err != nil {
		return err
	}

	// E N D AGENT JSON CONFIG
	//
	//

	return nil

}

// read package files in byte
func (gri *RemoteInstall) _tsk_read_package_file(t *Target, tm *TaskModel) error {
	files, err := ioutil.ReadDir(gri.srvTaskConfig.Files)

	// if no error
	if err == nil {
		for _, fn := range files {

			isExist := false
			for _, fb := range gri.package_files_bytes {

				// check if exist in the array
				// if exist, it will recheck
				// files_bytes are OK or not,
				// if not already read
				if fb.filename == fn.Name() {
					var err error
					var b []byte

					// check if already read the content
					if !fb.file_bytes_done {
						b, err = ioutil.ReadFile(gri.srvTaskConfig.Files + fn.Name())
						if err != nil {
							// write status
							tm.InserStatus(t, STATUS_SMB_COPY_READ_PACKAGE_FILES, err.Error(), gri.db)
							return err
						}

						fb.file_bytes = b
						fb.file_bytes_done = true

					}

					isExist = true
					break
				}
			}

			// if not exist
			if !isExist {
				var newFB package_files
				var err error
				var b []byte
				b, err = ioutil.ReadFile(gri.srvTaskConfig.Files + fn.Name())
				if err != nil {
					// write status
					tm.InserStatus(t, STATUS_SMB_COPY_READ_PACKAGE_FILES, err.Error(), gri.db)
					return err
				}

				newFB.filename = fn.Name()
				newFB.file_bytes = b
				newFB.file_bytes_done = true

				gri.package_files_bytes = append(gri.package_files_bytes, newFB)
			}
		}

		// no error
		return nil
	}

	// write status
	tm.InserStatus(t, STATUS_SMB_COPY_READ_PKG_DIR_FAILED, err.Error(), gri.db)
	return err

}
