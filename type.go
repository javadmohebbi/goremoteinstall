package goremoteinstall

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/hirochachacha/go-smb2"
	"gorm.io/gorm"
)

// this struct will help us to remotely install apps on the
// target machines.
type RemoteInstall struct {

	// task start time, it is shared between clients and servers using .json file
	startTime time.Time

	// Target file name taht include all target machines separeted by \n or \r\n
	Target string

	// Username that has access to ADMIN$ and enough privilege to install apps
	Username string

	// Domain name that user needs for it's authentication process. for local users, just a . (dot) is required
	Domain string

	// Password for the user
	Password string

	// list of files that should be copied to the target machines.
	FilesToCopy string

	// the file name that should run at first step
	FileToRun string

	// provide domain\user style
	backslashStyle string

	// provide user@domain style
	atSignStyle string

	// list of target objects to install
	targets []*Target

	// list of files to copy
	f2cs []string

	// just a ticker to show the progress if not silent
	progressTicker *time.Ticker

	// print debug info if it's true, it will disable the progress
	debug bool

	// if its true, nothing will be printed
	silent bool

	// just a time variable for start task time
	t time.Time

	// l         net.Listener

	// path to griAgent.exe file
	agentPath string

	// path to server task configuration (deprecated and kept here for backward compatibility)
	srvTaskPath string

	// Server task configuration
	srvTaskConfig *ServerTaskConfig

	// Global server configuration
	srvGlobalConfig *ServerGlobalConfig

	// for manage concurrent task using wait group
	wg *sync.WaitGroup

	// GORM database object for stroing logs on sqlite DB
	db *gorm.DB

	// Task ID is an identified for the current task
	TaskID string

	// socket connection to deployer server
	// deployer will listen both on tcp and unix socket
	// agents will connect using tcp and deployer will forward them
	// to all remote installer servers
	UnixClientSock *net.Conn

	// doneCh    chan bool
	// allDoneCh chan bool
	// managerCh chan interface{}

	// // The close flag allows we know when we can close the manager
	// closed bool

	// // The running count allows we know the number of goroutines are running
	// runningCount int32

	// allJobChan chan int

	// channel for the time that some signal that push from OS or User
	closeChannel     chan os.Signal
	closeSignalRecvd bool

	// this will help us read files one at first
	// place and then use this var
	// to copy files to the target machines
	package_files_bytes      []package_files
	package_files_bytes_done bool

	// like package_files_bytes, it will be the
	// installer agent at first copy and it will be user
	// for the next targets
	agent_files_bytes      []byte
	agent_files_bytes_done bool

	// config json for remote install agent
	agent_file_config            Agent_file_config_type
	agent_files_config_byte      []byte
	agent_files_config_byte_done bool

	//
	unixConn net.Conn
}

// this type is uses to provide dynamic .json configuration for
// griAgent.
type Agent_file_config_type struct {

	// time is
	Time string `json:"Time"`

	// gri server (deployer) address
	SrvAddress string `json:"DeployerAddress"`
	// gri server (deployer) port
	SrvPort int `json:"DeployerPort"`

	// Host - provided from targets
	Host string `json:"Host"`
	// Generated host id
	HostID string `json:"HostID"`

	// Task ID
	TaskID string `json:"TaskID"`

	// Computer ID - reserved for future usage
	ComputerID string `json:"ComputerID"`

	// IP - reserved for future usage
	IP string `json:"IP"`

	// Computer name - reserved for future usage
	ComputerName string `json:"ComputerName"`

	// Bootstrap is the first app that should run on the system
	Bootstrap string `json:"Bootstrap"`

	// Params is a array of command line arguments
	Params []string `json:"Params"`

	// Working dir usually "%systemroot\\TEMP\\YARMA\\%"
	Dir string `json:"Dir"`

	// timeout amount in minutes
	Timeout uint `json:"Timeout"`
}

// struct to keep the package file inside and used for multiple hosts
// it will be filled when files copied on the first target
// and will be used next time
type package_files struct {
	filename        string
	file_bytes      []byte
	file_bytes_done bool
}

// New - create new instance of RemoteInstall
func New(dbg, silent bool, gc *ServerGlobalConfig, tc *ServerTaskConfig, confName string) *RemoteInstall {

	// l, err := net.Listen("tcp", listen)
	// if err != nil {
	// 	fmt.Println("could not listen due to error: ", err)
	// 	os.Exit(int(ERR_TCP_LISTEN))
	// }

	// generate task ID
	tskId := fmt.Sprintf("tsk-%s", confName)

	// create or open sqlite db
	db := GetTaskDB(tskId)

	// just add the start time
	t := time.Now()

	// prepare and initialize unix socket clinet
	unixClient, err := net.Dial("unix", gc.DeployerSocket)
	if err != nil {
		log.Println("could not connect to unix socket: ", gc.DeployerSocket)
		os.Exit(int(ERR_UNIX_CLIENT_SOCKET))
	}

	// initialize socket client
	rq := ClientServerReqResp{
		Server:  true,
		Command: CMD_INIT,
		Host:    fmt.Sprintf("remote-installer_%v", t.Unix()),
	}
	bts, err := json.Marshal(&rq)
	if err != nil {
		log.Println("could not conver req to JSON unix socket client request: ", err)
		os.Exit(int(ERR_UNIX_CLIENT_SOCKET_MARSHAL))
	}

	// initialize socket client
	_, err = unixClient.Write([]byte(fmt.Sprintf("%s\n", bts)))
	if err != nil {
		if err != nil {
			fmt.Println("could not initialize unix socket client: ", err)
			os.Exit(int(ERR_UNIX_CLIENT_SOCKET_INIT))
		}
	}
	// println("write to server = ", string(bts))

	// reply := make([]byte, 1024)

	// _, err = unixClient.Read(reply)
	// if err != nil {
	// 	println("Write to server failed:", err.Error())
	// 	os.Exit(1)
	// }
	// println("reply from server=", string(reply))

	gri := &RemoteInstall{
		TaskID: tskId,
		db:     db,

		UnixClientSock: &unixClient,

		Username:    tc.Username,
		Domain:      tc.Domain,
		Password:    tc.Password,
		FilesToCopy: tc.Files,
		FileToRun:   tc.Bootstrap,

		backslashStyle: fmt.Sprintf("%s\\%s", tc.Domain, tc.Username),
		atSignStyle:    fmt.Sprintf("%s@%s", tc.Username, tc.Domain),

		debug:  dbg,
		silent: silent,

		t: t,
		// l:         l,
		agentPath: gc.AgentPath,

		srvTaskConfig: tc,
		Target:        tc.Targets,

		srvGlobalConfig: gc,

		closeChannel: make(chan os.Signal, 1),
		wg:           &sync.WaitGroup{},
	}

	// handle requests
	go gri.handleClientConnections()

	return gri
}

type Target struct {
	HostID string

	IsInProgress bool
	HasError     bool
	IsDone       bool

	Host        string
	Status      Status
	CurrentDesc string
	DescPayload string

	Progress uint

	Conn   net.Conn
	ConnOK bool

	smb2Dialer  *smb2.Dialer
	smb2Session *smb2.Session

	dirPath string

	// this variable only used for
	// updating network statuses which are comming
	// from our agent
	_tm *TaskModel

	done chan bool
}
