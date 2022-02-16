package main

import (
	"bufio"
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

var confs []string
var selected int

func main() {
	confDir := flag.String("conf", "/opt/yarma/goremoteinstall/etc/conf.d/", "Configuration directory which includes all the defined tasks configurations")

	silent := flag.Bool("s", false, "if provided, no output will be displayed about progress")

	flag.Parse()

	if *confDir == "" {
		log.Fatal("-conf must be specified")
	}

	_, err := walkAndChoose(*confDir)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf(">>> Conf. [%s] is selected for running the task. ", confs[selected])

	getYesOrNo("\n\nDo you want to continue? [y=yes,others for no]")

	gc, err := goremoteinstall.NewGloablConfig("/opt/yarma/goremoteinstall/etc/")
	if err != nil {
		log.Fatal(err)
	}

	tc, err := goremoteinstall.NewServerTaskConfig(filepath.Dir(confs[selected]), filepath.Base(confs[selected]))
	if err != nil {
		log.Fatal(err)
	}

	var filename = filepath.Base(confs[selected])
	var extension = filepath.Ext(filename)
	confName := filename[0 : len(filename)-len(extension)]

	gri := goremoteinstall.New(
		false,
		*silent,
		gc,
		tc,
		confName,
	)

	gri.Run()

}

func walkAndChoose(confDir string) (int, error) {
	files, err := ioutil.ReadDir(confDir)
	if err != nil {
		log.Fatalln("Can not read conf.d: ", err)
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" {
			confs = append(confs, fmt.Sprintf("%s%s", confDir, file.Name()))
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
		// _, err = cmd.StdinPipe()
		// if err != nil {
		// 	log.Fatal(err)
		// }

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

func waitForEnterKey(msg string) {
	consoleReader := bufio.NewReaderSize(os.Stdin, 1)
	fmt.Printf("%s > ", msg)
	for {
		input, _ := consoleReader.ReadByte()
		ascii := input

		log.Println(ascii)

		if ascii == 10 {
			return
		}
	}
}

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
