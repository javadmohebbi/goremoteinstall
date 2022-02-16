package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/javadmohebbi/goremoteinstall"
	"github.com/javadmohebbi/goremoteinstall/deployer"
)

func main() {
	globalConf := flag.String("gc", "/opt/yarma/goremoteinstall/etc/", "Global configuration directory")
	flag.Parse()

	gconf, err := goremoteinstall.NewGloablConfig(*globalConf)
	if err != nil {
		log.Fatal(err)
	}

	l := ""
	if gconf.DeployerListen == "0.0.0.0" {
		l = fmt.Sprintf(":%d", gconf.DeployerPort)
	} else {
		l = fmt.Sprintf("%s:%d", gconf.DeployerListen, gconf.DeployerPort)
	}

	d := deployer.New(gconf.DeployerSocket,
		l)

	go d.ListenSocket()
	time.Sleep(3 * time.Second)
	go d.ListenTCP()

	select {}
}
