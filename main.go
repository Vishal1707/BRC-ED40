package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kardianos/service"
	"github.com/urfave/cli"
	"BRC-ED40/scanner"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	appname = "BRC"
	appver  = "0.0.1"
)

var (
	host     string
	port     string
	svcOp    string
	certFile string
	keyFile  string
	noTLS    bool
	hub      *Hub
	logger   service.Logger
)

type BarCode struct {
	Code string
}

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	hub = newHub()
	go hub.run()
	go scanner.ScanForever(processScanFn)
	go p.run()
	return nil
}

func (p *program) run() {
	// Do work here
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	var addr = flag.String("addr", host+":"+port, "http service address")
	var err error

	if noTLS {
		err = http.ListenAndServe(*addr, nil)
	} else {
		err = http.ListenAndServeTLS(*addr, certFile, keyFile, nil)
	}

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func processScanFn(scanData string) {

	if scanData[0:2] == "-j" {
		scanData = scanData[2:]
	}
	fmt.Println("Scanned Data: %s", scanData)
	data := BarCode{
		Code: scanData,
	}

	obj, err := json.Marshal(data)
	hub.broadcast <- obj

	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = appname
	app.Usage = "BarCode Reader with Websocket event support"
	app.Version = appver
	app.Authors = []cli.Author{
		{
			Name:  "Vishal Kumar Singh",
			Email: "vishalkumarsingh1707@gmail.com",
		},
	}
	app.Copyright = "(c) 2017 Vishal Kumar Singh"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "host",
			Value:       "localhost",
			Usage:       "Host for websocket",
			EnvVar:      "BRC_HOST",
			Destination: &host,
		},
		cli.StringFlag{
			Name:        "port,p",
			Value:       "8080",
			Usage:       "the websocket port",
			EnvVar:      "BRC_PORT",
			Destination: &port,
		},
		cli.StringFlag{
			Name:        "cert",
			Value:       "cert.pem",
			Usage:       "The certificate file to use for TLS",
			EnvVar:      "BRC_CERT",
			Destination: &certFile,
		},
		cli.StringFlag{
			Name:        "key",
			Value:       "key.pem",
			Usage:       "The key file to use for TLS",
			EnvVar:      "BRC_KEY",
			Destination: &keyFile,
		},
		cli.BoolFlag{
			Name:        "notls",
			Usage:       "Disable TLS for the websoclket server",
			EnvVar:      "BRC_NOTLS",
			Destination: &noTLS,
		},
		cli.StringFlag{
			Name:        "service",
			Usage:       "Service operation",
			EnvVar:      "BRC_OP",
			Destination: &svcOp,
		},
	}

	app.Action = func(c *cli.Context) error {
		argHost := fmt.Sprintf("--host=%s", host)
		argPort := fmt.Sprintf("--port=%s", port)
		argCert := fmt.Sprintf("--cert=%s", certFile)
		argKey := fmt.Sprintf("--key=%s", keyFile)

		args := []string{argHost, argPort}

		if noTLS {
			args = append(args, "--notls")
		} else {
			args = append(args, argCert)
			args = append(args, argKey)
		}

		svcConfig := &service.Config{
			Name:        "BRC-ED40",
			DisplayName: "BRCService",
			Description: "Websocket for barcode reader service",
			Arguments:   args,
		}

		prg := &program{}

		svc, err := service.New(prg, svcConfig)
		if err != nil {
			log.Fatal(err)
		}

		if svcOp != "" {
			svcOp = strings.ToLower(svcOp)
			err := service.Control(svc, svcOp)
			if err != nil {
				log.Fatal(err)
				return err
			}
			return nil
		}

		logger, err = svc.Logger(nil)
		if err != nil {
			log.Fatal(err)
		}

		err = svc.Run()
		if err != nil {
			logger.Error(err)
		}
		return err
	}
	app.Run(os.Args)
}
