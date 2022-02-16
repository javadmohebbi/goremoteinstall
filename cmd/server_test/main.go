package main

// import (
// 	"flag"
// 	"fmt"
// 	"os"
// 	"syscall"

// 	goremoteinstallation "github.com/javadmohebbi/goremoteinstall"
// 	"golang.org/x/term"
// )

// const (
// 	_targets    = "targets"
// 	_username   = "username"
// 	_domainname = "domain"
// 	_f2c        = "f2c"
// 	_f2r        = "f2r"
// )

// func main() {
// 	gri := initApp()

// 	gri.Run()

// }

// func initApp() *goremoteinstallation.RemoteInstall {
// 	t := flag.String(_targets, "/opt/yarma/goremoteinstall/etc/targets.txt", "List of targets in a file delimited with \\n")
// 	u := flag.String(_username, "", "Username to copy files using SMB and install the package")
// 	d := flag.String(_domainname, "", "Domain name for -username field")
// 	f2c := flag.String(_f2c, "/opt/yarma/goremoteinstall/etc/f2c.txt", "Required files need to be copied on ADMIN$\\TMP. It is a file that each line must be a path to the required files")
// 	f2r := flag.String(_f2r, "/opt/yarma/goremoteinstall/etc/f2r.txt", "File that should be run to start the installation with it's required parameters")

// 	pathToAgent := flag.String("agent", "/opt/yarma/goremoteinstall/bin/win/griAgent.exe", "Path to windows service agent")

// 	listenOn := flag.String("l", "0.0.0.0:9999", "listen address and port for TCP connection between this app and our windows installer service")

// 	v := flag.Bool("v", false, "Verbose logs")

// 	flag.Parse()

// 	ifEmptyFlag(t, _targets, goremoteinstallation.ERR_TARGETS_NOT_PROVIDED)
// 	ifEmptyFlag(u, _username, goremoteinstallation.ERR_USERNAME_NOT_PROVIDED)
// 	ifEmptyFlag(d, _domainname, goremoteinstallation.ERR_DOMAIN_NOT_PROVIDED)
// 	ifEmptyFlag(f2c, _f2c, goremoteinstallation.ERR_F2C_NOT_PROVIDED)
// 	ifEmptyFlag(f2r, _f2r, goremoteinstallation.ERR_F2R_NOT_PROVIDED)

// 	// get password from STDIN
// 	fmt.Printf("Enter password for '%s\\%s': ", *d, *u)
// 	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(int(goremoteinstallation.ERR_READ_PASSWORD_STDIN))
// 	}
// 	p := string(bytePassword)
// 	fmt.Println()
// 	fmt.Println()

// 	// create new instance of GRI
// 	gri := goremoteinstallation.New(
// 		*t, *u, *d, p, *f2c, *f2r, *v, *listenOn, *pathToAgent,
// 		"/opt/yarma/goremoteinstall/etc/conf.d/epskit_x64_6.6.20.294.yaml",
// 	)

// 	return gri

// }

// func ifEmptyFlag(s *string, f string, i goremoteinstallation.Errors) {
// 	if *s == "" {
// 		fmt.Printf("\n[ERR]-%v must be provided\n\n", f)
// 		flag.PrintDefaults()
// 		os.Exit(int(i))
// 	}
// }
