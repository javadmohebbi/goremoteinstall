package goremoteinstall

import (
	"bufio"
	"fmt"
	"os"
)

func (gri *RemoteInstall) readFilesToCopy() {
	f, err := os.Open(gri.FilesToCopy)
	if err != nil {
		fmt.Println(err)
		os.Exit(int(ERR_CAN_T_READ_F2C))
	}

	defer f.Close()

	var f2cs []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() != "" {
			s := scanner.Text()
			f2cs = append(f2cs, s)
		}
	}

	if len(f2cs) == 0 {
		fmt.Printf("%v file seems empty", gri.Target)
		os.Exit(int(ERR_TARGETS_IS_EMPTY))
	}

	gri.f2cs = f2cs

}
