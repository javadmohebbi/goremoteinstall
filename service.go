package goremoteinstall

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func (gri *RemoteInstall) CreateAndStartService(t *Target) {

	// gri._deleteServiceIfExist(t)

	_ = gri._createService(t)
	// if err != nil {
	gri._stopService(t)
	time.Sleep(10 * time.Second)
	gri._startService(t)
	// }

}

func (gri *RemoteInstall) _ifSvcExist(t *Target) (bool, error) {
	cmd := exec.Command("net",
		"rpc",
		"service",
		"status",
		"griAgent",
		"-I",
		t.Host,
		"-U",
		fmt.Sprintf("%s%%%s", gri.Username, gri.Password),
		"-W",
		gri.Domain,
	)

	var out, _err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &_err

	err := cmd.Run()
	if err != nil {
		// target service is not exists
		if strings.Contains(_err.String(), "Failed to open service") {
			return false, nil
		} else {
			return false, errors.New(
				fmt.Sprintf("%v, %s", err, _err.String()),
			)
		}
	} else {
		return true, nil
	}
}

func (gri *RemoteInstall) _createService(t *Target) error {

	t.Status = STATUS_CREATING_SERVICE
	t.CurrentDesc = "Creating Service"

	check, err := gri._ifSvcExist(t)
	if err != nil {
		// fmt.Println("[ERR]", err)
		t.Status = STATUS_SERVICE_CREATION_FAILED
		t.CurrentDesc = "Creating failed:" + err.Error()
		return err
	}
	if !check {

		cmd := exec.Command("net",
			"rpc",
			"service",
			"create",
			"griAgent",
			"YARMA Go Remote Install Agent",
			fmt.Sprintf("\"%%windir%%\\TEMP\\YARMA\\%s\"", "griAgent.exe"),
			"-I",
			t.Host,
			"-U",
			fmt.Sprintf("%s%%%s", gri.Username, gri.Password),
			"-W",
			gri.Domain,
		)

		var out, _err bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &_err

		err := cmd.Run()
		if err != nil {
			fmt.Println("[ERR]", err)
			t.Status = STATUS_SERVICE_CREATION_FAILED
			t.CurrentDesc = "Creating failed:" + err.Error()
			return errors.New(
				fmt.Sprintf("%v, %s", err, _err.String()),
			)
		} else {
			// fmt.Println("=====", out.String())
			t.Status = STATUS_SERVICE_CREATED
			t.CurrentDesc = "Installation service created"
			return nil
		}
	} else {
		return nil
	}

}

func (gri *RemoteInstall) _startService(t *Target) error {

	t.Status = STATUS_STARTING_SERVICE
	t.CurrentDesc = "Starting Service"

	cmd := exec.Command("net",
		"rpc",
		"service",
		"start",
		"griAgent",
		"-I",
		t.Host,
		"-U",
		fmt.Sprintf("%s%%%s", gri.Username, gri.Password),
		"-W",
		gri.Domain,
	)

	var out, _err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &_err

	err := cmd.Run()
	if err != nil {
		// fmt.Println("[ERR]", err)
		t.Status = STATUS_STARTING_SERVICE_FAILED
		t.CurrentDesc = "Starting service failed:" + err.Error()
		// write stautus
		_ = t._tm.InserStatus(t, STATUS_STARTING_SERVICE_FAILED, err.Error(), gri.db)

		return errors.New(
			fmt.Sprintf("%v, %s", err, _err.String()),
		)

	} else {
		// fmt.Println("=====", out.String())
		t.Status = STATUS_SERVICE_STARTED
		t.CurrentDesc = "Installation service started"
		return nil
	}

}

func (gri *RemoteInstall) _stopService(t *Target) error {

	t.Status = STATUS_STARTING_SERVICE
	t.CurrentDesc = "Starting Service"

	cmd := exec.Command("net",
		"rpc",
		"service",
		"stop",
		"griAgent",
		"-I",
		t.Host,
		"-U",
		fmt.Sprintf("%s%%%s", gri.Username, gri.Password),
		"-W",
		gri.Domain,
	)

	var out, _err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &_err

	err := cmd.Run()
	if err != nil {
		// fmt.Println("[ERR]", err)
		t.Status = STATUS_STARTING_SERVICE_FAILED
		t.CurrentDesc = "Starting service failed:" + err.Error()
		// write stautus
		_ = t._tm.InserStatus(t, STATUS_STARTING_SERVICE_FAILED, err.Error(), gri.db)

		return errors.New(
			fmt.Sprintf("%v, %s", err, _err.String()),
		)
	} else {
		// fmt.Println("=====", out.String())
		t.Status = STATUS_SERVICE_STARTED
		t.CurrentDesc = "Installation service started"
		return nil
	}

}

// func (gri *RemoteInstall) _deleteServiceIfExist(t *Target) {
// 	cmd := exec.Command("net",
// 		"rpc",
// 		"service",
// 		"restart",
// 		"griAgent",
// 		"-I",
// 		t.Host,
// 		"-U",
// 		fmt.Sprintf("%s%%%s", gri.Username, gri.Password),
// 		"-W",
// 		gri.Domain,
// 	)

// 	var out bytes.Buffer
// 	cmd.Stdout = &out

// 	err := cmd.Run()
// 	if err != nil {
// 		fmt.Println("[ERR]", err)
// 	} else {
// 		// fmt.Println("=====", out.String())
// 	}
// }
