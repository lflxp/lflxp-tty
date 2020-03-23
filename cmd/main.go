package main

import (
	"flag"

	"github.com/lflxp/lflxp-tty/pkg"
)

var (
	enableTLS      bool
	crtPath        string
	keyPath        string
	isProf         bool
	isXsrf         bool
	isAudit        bool
	isPermitWrite  bool
	MaxConnections int64
	isReconnect    bool
	isDebug        bool
	username       string
	password       string
	port           string
	host           string
)

func init() {
	flag.StringVar(&username, "u", "", "username")
	flag.StringVar(&password, "p", "", "password")
	flag.StringVar(&host, "H", "0.0.0.0", "http bind host")
	flag.StringVar(&port, "P", "8080", "http bind port")
	flag.BoolVar(&isDebug, "d", false, "debug log mode")
	flag.BoolVar(&isReconnect, "r", false, "auto reconnect")
	flag.BoolVar(&isPermitWrite, "w", false, "Whether to turn on write mode")
	flag.BoolVar(&isAudit, "a", false, "Open audit or not")
	flag.BoolVar(&isXsrf, "x", false, "Whether to turn on xsrf, which is on by default")
	flag.BoolVar(&isProf, "f", false, "Whether to enable pprof performance analysis")
	flag.BoolVar(&enableTLS, "t", false, "Open HTTPS or not")
	flag.StringVar(&crtPath, "c", "./server.crt", "*.crt file path")
	flag.StringVar(&keyPath, "k", "./server.key", "*.key file path")
	flag.Int64Var(&MaxConnections, "m", 0, "maximum connection")
	flag.Parse()
}

func main() {
	tty := pkg.Tty{
		EnableTLS:      enableTLS,
		CrtPath:        crtPath,
		KeyPath:        keyPath,
		IsProf:         isProf,
		IsXsrf:         isXsrf,
		IsAudit:        isAudit,
		IsPermitWrite:  isPermitWrite,
		MaxConnections: MaxConnections,
		IsReconnect:    isReconnect,
		IsDebug:        isDebug,
		Username:       username,
		Password:       password,
		Port:           port,
		Host:           host,
		Cmds:           flag.Args(),
	}

	err := tty.Execute()
	if err != nil {
		panic(err)
	}
}
