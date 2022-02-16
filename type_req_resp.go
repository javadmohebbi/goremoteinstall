package goremoteinstall

import (
	"errors"
	"fmt"
	"log"
)

func (gri *RemoteInstall) parseAgentReqAndStore(
	rr ClientServerReqResp,
) error {

	if rr.TaskID != gri.TaskID {
		if gri.debug {
			log.Println("invalid task ID")
			return errors.New("invalid task ID")
		}
	}

	_t, err := gri.parseAgentReqAndStore_validateHostID(rr)
	if err != nil {
		return err
	}
	err = gri.parseAgentReqAndStore_validateCmdAndStore(rr, _t)
	return err
}

func (gri *RemoteInstall) parseAgentReqAndStore_validateHostID(
	rr ClientServerReqResp,
) (*Target, error) {

	var _t *Target

	for _, t := range gri.targets {
		if rr.HostID == t.HostID {
			_t = t
			break
		}
	}
	if _t == nil {
		return nil, errors.New(fmt.Sprintf("unknow host id %v", rr.HostID))
	}

	return _t, nil
}

func (gri *RemoteInstall) parseAgentReqAndStore_validateCmdAndStore(
	rr ClientServerReqResp,
	t *Target,
) error {

	switch rr.Command {
	case CMD_INIT:
		// nothing to do
	case CMD_BOOTSTRAP_START:

		// write status
		_ = t._tm.InserStatus(t, STATUS_EXECUTING, rr.DescPayload, gri.db)

		if gri.debug {
			log.Printf("Bottstrap is starting on '%s'\n", t.Host)
		}

	case CMD_BOOTSTRAP_STARTED:
		// write status
		_ = t._tm.InserStatus(t, STATUS_EXECUTED, rr.DescPayload, gri.db)

		if gri.debug {
			log.Printf("Bottstrap is started on '%s'\n", t.Host)
		}

	case CMD_BOOTSTRAP_FINISH_ERROR:

		// write status
		_ = t._tm.InserStatus(t, STATUS_EXECUTING_FAILED, rr.DescPayload, gri.db)

		if gri.debug {
			log.Printf("Bottstrap is finish with error on '%s'\n", t.Host)
		}

	case CMD_BOOTSTRAP_FINISH_DONE:

		// write status
		_ = t._tm.InserStatus(t, STATUS_EXECITING_DONE, rr.DescPayload, gri.db)

		if gri.debug {
			log.Printf("Bottstrap is finish successfully on '%s'\n", t.Host)
		}
	}

	return nil
}
