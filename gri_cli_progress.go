package goremoteinstall

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	term_w    = 0
	term_h    = 0
	tc_Reset  = "\033[0m"
	tc_Red    = "\033[31m"
	tc_Green  = "\033[32m"
	tc_Yellow = "\033[33m"
	tc_Blue   = "\033[34m"
	tc_Purple = "\033[35m"
	tc_Cyan   = "\033[36m"
	tc_Gray   = "\033[37m"
	tc_White  = "\033[97m"
)

const (
	cli_progress_clear_screen = "\033[2J"
)

func (gri *RemoteInstall) cliProgress() {
	var err error

	// get term W, H
	term_w, term_h, err = gri._cliProgress_getTermWH()
	if err != nil {
		// back to debug mod
		gri.debug = true
		gri.silent = true
		return
	}

	gri.progressTicker = time.NewTicker(500 * time.Millisecond)

	go func() {

		for {
			select {
			case <-gri.progressTicker.C:

				// clear screen
				gri._cliProgress_clear()

				// print global info
				gri._cliPrgress_print_globalInfo(false)

			case <-gri.closeChannel:
				gri.progressTicker.Stop()
				return
			}
		}
	}()

	// fmt.Printf("Width: %d, Height: %d", term_w, term_h)

}

// clear screen
func (gri *RemoteInstall) _cliProgress_clear() {
	fmt.Printf(cli_progress_clear_screen)
}

// clear screen
func (gri *RemoteInstall) _cliProgress_getTermWH() (int, int, error) {
	w, h, err := terminal.GetSize(int(os.Stdin.Fd()))

	if err != nil {
		fmt.Println("could not get terminal width and height", err)
		return 0, 0, err
	}

	return w, h, nil
}

func (gri *RemoteInstall) _cliPrgress_print_globalInfo(all bool) {
	// set cursor to 0,0
	fmt.Printf("\033[%d;%dH", 0, 0)

	c, d, e, w := gri._cliProgress_GetTargetInfo()

	var percentage float64
	percentage = float64(100*(d+e)) / float64(len(gri.targets))

	fmt.Printf("%v%%%.2f%v %vTask: %v%s %vTotal: %v%d %vDone: %v%d %vCurr: %v%d %vErr: %v%d %vWait: %v%d\n",
		tc_Cyan, percentage, tc_Reset,
		tc_Cyan, tc_Reset,
		tc_Blue,
		tc_Reset, gri.TaskID,
		len(gri.targets),
		tc_Green, tc_Reset, d,
		tc_Yellow, tc_Reset, c,
		tc_Red, tc_Reset, e,
		tc_Purple, tc_Reset, w,
	)

	for _, _t := range gri.targets {

		if _t.IsInProgress {
			strPayload := ""
			if _t.DescPayload != "" {
				strPayload = fmt.Sprintf("; %s", _t.DescPayload)
			}
			fmt.Printf("%v%s %v-->%v %s%s%v\n",
				tc_Yellow, _t.Host, tc_Reset,
				tc_Yellow, _t.Status.String(),
				strPayload,
				tc_Reset,
			)
		}

		// if all = true
		// it will display all info
		// usually when task finished
		if all {
			if _t.IsDone {
				fmt.Printf("%v%s %v-->%v %s%s%v\n",
					tc_Green, _t.Host, tc_Reset,
					tc_Cyan, _t.Status.String(),
					"",
					tc_Reset,
				)
			}
			if _t.HasError {
				strPayload := ""
				if _t.DescPayload != "" {
					strPayload = fmt.Sprintf("; %s", _t.DescPayload)
				}
				fmt.Printf("%v%s %v-->%v %s%s%v\n",
					tc_Red, _t.Host, tc_Reset,
					tc_Red, _t.Status.String(),
					strPayload,
					tc_Reset,
				)
			}
		}
	}

}

// get current info
func (gri *RemoteInstall) _cliProgress_GetTargetInfo() (c, d, e, w int) {
	e = 0
	w = 0
	d = 0
	c = 0

	for _, t := range gri.targets {
		if t.IsInProgress {
			c++
			continue
		}
		if t.IsDone {
			d++
			continue
		}
		if t.HasError {
			e++
			continue
		}

		w++
	}

	//
	return c, d, e, w
}
