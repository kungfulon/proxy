package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"proxy/internal/proxy"
)

var (
	host            = flag.String("host", "0.0.0.0:3333", "Host to listen on")
	protocol        = flag.String("prot", "https", "Protocol")
	certFile        = flag.String("cert", "", "Certificate file")
	keyFile         = flag.String("key", "", "Private key file")
	credentialsFile = flag.String("creds", "", "Credentials file")
)

func main() {
	flag.Parse()

	if err := verifyFlags(); err != nil {
		fmt.Println(err)
		flag.Usage()
		os.Exit(1)
	}

	proxy.Start(proxy.Config{
		Host:            *host,
		Protocol:        *protocol,
		CertFile:        *certFile,
		KeyFile:         *keyFile,
		CredentialsFile: *credentialsFile,
	})
}

func verifyFlags() error {
	if *protocol != "http" && *protocol != "https" {
		return errors.New("only http and https protocol are allowed")
	}

	if *protocol == "https" {
		if _, err := os.Stat(*certFile); err != nil {
			return err
		}

		if _, err := os.Stat(*keyFile); err != nil {
			return err
		}
	}

	//if _, err := os.Stat(*credentialsFile); err != nil {
	//	return err
	//}

	return nil
}
