package goremoteinstall

import (
	"encoding/json"
	"errors"
)

type ClientServerReqResp struct {
	ItSelf bool `json:"itSelf,omitempty"`
	Agent  bool `json:"agent,omitempty"`
	Server bool `json:"server,omitempty"`

	Ack bool `json:"ack,omitempty"`

	Command Command `json:"cmd"`

	RequestID string `json:"reqId,omitempty"`

	// task id
	TaskID string `json:"taskId,omitempty"`

	// host ip or name
	Host string `json:"host,omitempty"`

	// host id or name
	HostID string `json:"hostId,omitempty"`

	// current status
	Status Status `json:"status,omitempty"`

	// current status desc
	Desc string `json:"desc,omitempty"`

	DescPayload string `json:"descPayload,omitempty"`
}

type Command int

const (
	CMD_INIT            = iota + 0x1000
	CMD_BOOTSTRAP_START //starting
	CMD_BOOTSTRAP_STARTED
	CMD_BOOTSTRAP_FINISH_DONE
	CMD_BOOTSTRAP_FINISH_ERROR

	CMD_UNKNOWN
)

func StringToJSONClientServerReqResp(reqJson string) (ClientServerReqResp, error) {
	var jsonReq ClientServerReqResp
	err := json.Unmarshal([]byte(reqJson), &jsonReq)
	if err != nil {
		return ClientServerReqResp{}, err
	}

	if jsonReq.RequestID == "" {
		return ClientServerReqResp{}, errors.New("reqId could not be empty")
	}

	return jsonReq, nil
}

func (req ClientServerReqResp) JSONToStringClientServerReqResp() (string, error) {
	// var jsonReq ClientServerReqResp
	// err := json.Unmarshal([]byte(reqJson), &jsonReq)
	// if err != nil {
	// 	return ClientServerReqResp{}, err
	// }

	// return jsonReq, nil
	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (req ClientServerReqResp) JSONToByteClientServerReqResp() ([]byte, error) {
	// var jsonReq ClientServerReqResp
	// err := json.Unmarshal([]byte(reqJson), &jsonReq)
	// if err != nil {
	// 	return ClientServerReqResp{}, err
	// }

	// return jsonReq, nil
	b, err := json.Marshal(req)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}
