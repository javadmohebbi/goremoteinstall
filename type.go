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

type RemoteInstall struct {
	startTime time.Time

	Target       string
	Username     string
	Domain       string
	Password     string
	FilesToCopy  string
	FileToRemote string

	backslashStyle string
	atSignStyle    string

	targets []*Target
	f2cs    []string

	progressTicker *time.Ticker

	debug  bool
	silent bool

	t time.Time
	// l         net.Listener
	agentPath string

	srvTaskPath   string
	srvTaskConfig *ServerTaskConfig

	srvGlobalConfig *ServerGlobalConfig

	wg *sync.WaitGroup

	db *gorm.DB

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

	unixConn net.Conn
}

type Agent_file_config_type struct {
	Time string `json:"Time"`

	SrvAddress string `json:"DeployerAddress"`
	SrvPort    int    `json:"DeployerPort"`

	Host   string `json:"Host"`
	HostID string `json:"HostID"`

	TaskID     string `json:"TaskID"`
	ComputerID string `json:"ComputerID"`

	IP           string `json:"IP"`
	ComputerName string `json:"ComputerName"`

	Bootstrap string   `json:"Bootstrap"`
	Params    []string `json:"Params"`

	Dir string `json:"Dir"`

	Timeout uint `json:"Timeout"`
}

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

	tskId := fmt.Sprintf("tsk-%s", confName)
	db := GetTaskDB(tskId)
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

		Username:     tc.Username,
		Domain:       tc.Domain,
		Password:     tc.Password,
		FilesToCopy:  tc.Files,
		FileToRemote: tc.Bootstrap,

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
