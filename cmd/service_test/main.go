package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("net",
		"rpc",
		"service",
		"status",
		"griAgent",
		"-I",
		"192.168.59.55",
		"-U",
		fmt.Sprintf("%s%%%s", "mj", "xxXX1234!"),
		"-W",
		".",
	)

	var out, _err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &_err

	err := cmd.Run()
	if err != nil {
		if strings.Contains(_err.String(), "Failed to open service") {
			log.Println(_err.String())
			return
		}

		log.Fatalln(err, _err.String())

	}

	log.Println(out.String())
}
