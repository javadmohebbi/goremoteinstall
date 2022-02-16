package goremoteinstall

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func (gri *RemoteInstall) readTargets() {
	f, err := os.Open(gri.Target)
	if err != nil {
		fmt.Println(err)
		os.Exit(int(ERR_CAN_T_READ_TARGETS))
	}

	defer f.Close()

	var targets []*Target

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() != "" {
			targets = append(targets, &Target{
				HostID:      fmt.Sprintf("%d", time.Now().UnixNano()),
				Host:        scanner.Text(),
				Status:      STATUS_NOT_STARTED_YET,
				CurrentDesc: "",
				Progress:    0,
			})
		}
	}

	if len(targets) == 0 {
		fmt.Printf("%v file seems empty", gri.Target)
		os.Exit(int(ERR_TARGETS_IS_EMPTY))
	}

	gri.targets = targets

}
