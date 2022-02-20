package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/javadmohebbi/goremoteinstall"
)

// variable for task confs
var confs []string

// keep the selected conf index
var selected int

// default path
var defaultConfDir = "/opt/yarma/goremoteinstall/etc/conf.d/"

func main() {

	/**
	parse the flag and commandline completion
	**/

	// define conf dir
	// confDir := flag.String("d", defaultConfDir, "Configuration directory which includes all the defined tasks configurations")
	confDir := &defaultConfDir

	// if provided, no output will be displayed
	silent := flag.Bool("s", false, "if provided, no output will be displayed about progress. if it provided, -n {CONFIGURATION FILE NAME} must be specified")

	// confName
	fConfName := flag.String("n", "", "Configuration file name. eg: installation-task-1.yaml")

	// parse the command line arguments
	flag.Parse()

	/**

	start the program logic

	**/

	// check if conf dir is empty
	if *confDir == "" {
		fmt.Println("-d dir must be specified")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// check if silent
	if *silent {
		if *fConfName == "" {
			fmt.Println("when you use -s, -n {CONFIGURATION FILE NAME} must be specified")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	// validate fConfName
	if filepath.Ext(*fConfName) != "yaml" || filepath.Ext(*fConfName) != "yml" {
		fmt.Printf("Configuration file should have 'yml' or 'yaml' extention. Your provided name is '%s'\n", *fConfName)
		flag.PrintDefaults()
		os.Exit(1)
	}

	// walk through dir and choose the conf
	_, err := walkAndChoose(*confDir, *fConfName)
	if err != nil {
		log.Fatalln(err)
	}

	// if not silent, ask user to
	if !*silent {
		// prinf the selected conf file name and path
		fmt.Printf(">>> Conf. [%s] is selected for running the task. ", confs[selected])

		// ask user to confirm
		getYesOrNo("\n\nDo you want to continue? [y=yes,others for no]")
	}

	// read gloabl configuration file
	gc, err := goremoteinstall.NewGloablConfig("/opt/yarma/goremoteinstall/etc/")
	if err != nil {
		log.Fatal(err)
	}

	// read task configuration file
	tc, err := goremoteinstall.NewServerTaskConfig(filepath.Dir(confs[selected]), filepath.Base(confs[selected]))
	if err != nil {
		log.Fatal(err)
	}

	// extract file and extentions
	var filename = filepath.Base(confs[selected])
	var extension = filepath.Ext(filename)
	confName := filename[0 : len(filename)-len(extension)]

	// create new instance of gri server
	gri := goremoteinstall.New(
		false,
		*silent,
		gc,
		tc,
		confName,
	)

	// run the gri server
	gri.Run()

}

// walk through conf dir and populate all of the possilbe file names
// if there is just one, select it, otherwise, ask user to select
// if fConfName is provided, it will only look for that file name and will select it
// fo the task
func walkAndChoose(confDir, fConfName string) (int, error) {
	files, err := ioutil.ReadDir(confDir)
	if err != nil {
		log.Fatalln("Can not read conf.d: ", err)
	}

	// check if fConfName is available
	if fConfName != "" {

		// check if file exists
		fconfNameCheck := false

		for _, file := range files {
			if file.Name() == fConfName {
				fconfNameCheck = true
				confs = append(confs, fmt.Sprintf("%s%s", confDir, file.Name()))
				break
			}
		}

		if !fconfNameCheck {
			return -1, errors.New(fmt.Sprintf("'%s' not found in the path '%s'", fConfName, confDir))
		}

	} else {
		// if fConfname is empty, loop through all confs
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".yaml" {
				confs = append(confs, fmt.Sprintf("%s%s", confDir, file.Name()))
			}
		}
	}

	if len(confs) == 0 {
		log.Fatal("no .yaml file available in ", confDir)
	}

	if len(confs) == 1 {
		return 0, nil
	}

	if len(confs) > 15 {

		stdinStr := ""
		for i, c := range confs {
			stdinStr += fmt.Sprintf("%d) %s\n", i+1, c)
		}

		waitForEnterKey("There are more than 15 .yaml conf, By pressing Enter, you will see all of them using 'less pager utility'. After choosing your .yml file, please press Q to quite 'less' pager and then you will be asked for the number assigned to your config file")

		cmd := exec.Command("/usr/bin/less")
		cmd.Stdin = strings.NewReader(stdinStr)

		cmd.Stdout = os.Stdout

		// Fork off a process and wait for it to terminate.
		err := cmd.Run()
		if err != nil {
			// log.Fatalln(err)
		}
		fmt.Println("")

	} else {

		for i, c := range confs {
			fmt.Printf("%d) %s\n", i+1, c)
		}

	}

	// get selected item from stdin
	getSelectedIterm()

	return selected, nil
}

// this will wait for enter key and continue,
// this is really handy for pager utility when we hint user that we have more than
// 15 files and /bin/less will be used to page the result
func waitForEnterKey(msg string) {
	consoleReader := bufio.NewReaderSize(os.Stdin, 1)
	fmt.Printf("%s > ", msg)
	for {
		input, _ := consoleReader.ReadByte()
		ascii := input

		// log.Println(ascii)

		if ascii == 13 {
			return
		}
	}
}

// get the selected item index
func getSelectedIterm() {
	var i int

	for {

		fmt.Printf("Please enter your selected item (it should be a number assigned to your .yaml file): ")
		_, err := fmt.Scanf("%d", &i)

		if err == nil {
			if i < 1 || i > len(confs) {
				fmt.Printf("Invalid number, number should be between %d and %d\n", 1, len(confs))
			} else {
				selected = i - 1
				break
			}
		} else {
			fmt.Printf("Got input: %v. ", i)
			fmt.Println("error: ", err)
		}
	}
	fmt.Println()

}

// ask user to continue the process
// only 'y' or 'Y' will continue the process, othewise it will be exited
// with 0 (zero) code
func getYesOrNo(msg string) {
	consoleReader := bufio.NewReaderSize(os.Stdin, 1)
	fmt.Printf("%s > ", msg)
	for {
		input, _ := consoleReader.ReadByte()
		ascii := input

		if ascii == 'Y' || ascii == 'y' {
			return
		} else {
			fmt.Println("Exiting...!")
			os.Exit(0)
		}
	}
}
