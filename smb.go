package goremoteinstall

import (
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"

	"github.com/hirochachacha/go-smb2"
)

func (gri *RemoteInstall) copyFilesOverSMB() {
	for _, _h := range gri.targets {

		fmt.Printf("[TARGET: %v]\n\n", _h.Host)

		// dial 445
		err := gri._smbDial445(_h)
		if err != nil {
			continue
		}
		defer _h.Conn.Close()

		// prepare session
		err = gri._smbPrepareSession(_h)
		if err != nil {
			continue
		}
		defer _h.smb2Session.Logoff()

		// copy files
		_ = gri._smbCopyFiles(_h)

		// start service
		gri.CreateAndStartService(_h)
	}
}

// dial 445
func (gri *RemoteInstall) _smbDial445(t *Target) error {
	t.Status = STATUS_DIAL_445
	t.CurrentDesc = "Dialing TCP 445..."

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:445", t.Host))
	if err != nil {
		// fmt.Printf("Can not dial %s:445 due to error: %v", t.Host, err)
		return err
	}

	t.Conn = conn
	t.ConnOK = true

	return nil

}

// list dir
func (gri *RemoteInstall) _smbPrepareSession(t *Target) error {
	t.Status = STATUS_PREPARE_SMB_SESSION
	t.CurrentDesc = "Prepairing SMB Session..."

	t.smb2Dialer = &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     gri.Username,
			Password: gri.Password,
			Domain:   gri.Domain,
		},
	}

	s, err := t.smb2Dialer.Dial(t.Conn)
	if err != nil {
		// fmt.Printf("Can not prepare session using username '%v' on host '%v' due to error: %v", gri.Username, t.Host, err)
		return err
	}

	t.smb2Session = s

	names, err := t.smb2Session.ListSharenames()
	if err != nil {
		// fmt.Printf("Could not list shares using username '%v' on host '%v' due to error: %v", gri.Username, t.Host, err)
		return err
	}

	if gri.debug && err == nil {
		if len(names) > 0 {
			// fmt.Printf("[DEBUG] This shares are available on '%v':\n", t.Host)
		}
		// for _, name := range names {
		// 	fmt.Printf("\t%v", name)
		// }
		// fmt.Println()
	}

	return nil

}

// copy
func (gri *RemoteInstall) _smbCopyFiles(t *Target) error {

	// _path := "\\goremoteinstall\\tmp\\"
	_path := fmt.Sprintf("TEMP\\goremoteinstall\\%v", gri.t.UnixNano())

	t.dirPath = _path

	fs, err := t.smb2Session.Mount("ADMIN$")
	if err != nil {
		panic(err)
	}
	defer fs.Umount()

	// fmt.Println(gri.f2cs)
	err = fs.MkdirAll(_path, 0777)
	if err != nil {
		panic(err)
	}

	//copy agent
	err = gri._copyAgent(t, fs)
	if err != nil {
		panic(err)
	}

	for _, f2c := range gri.f2cs {

		dst := fmt.Sprintf("%v\\%v",
			_path,
			filepath.Base(f2c),
		)

		if gri.debug {
			fmt.Printf("[DEBUG] Start copying file '%v' to '\\\\%v\\ADMIN$\\%v'\n", f2c, t.Host, dst)
		}

		b, err := ioutil.ReadFile(f2c)
		if err != nil {
			panic(err)
		}

		err = fs.WriteFile(
			dst,
			b,
			0777,
		)
		// fmt.Println(os.IsPermission(err)) // true
		if err != nil {
			panic(err)
		}

		// ctx, cancel := context.WithTimeout(context.Background(), 0)
		// defer cancel()
	}

	return nil
}

func (gri *RemoteInstall) _copyAgent(t *Target, fs *smb2.Share) error {

	_path := "TEMP\\goremoteinstall"
	dst := fmt.Sprintf("%v\\%v",
		_path,
		filepath.Base(gri.agentPath),
	)

	if gri.debug {
		fmt.Printf("[DEBUG] Start copying file '%v' to '\\\\%v\\ADMIN$\\%v'\n", gri.agentPath, t.Host, dst)
	}

	b, err := ioutil.ReadFile(gri.agentPath)
	if err != nil {
		return err
	}

	err = fs.WriteFile(
		dst,
		b,
		0777,
	)
	// fmt.Println(os.IsPermission(err)) // true
	if err != nil {
		return err
	}

	return nil
}
