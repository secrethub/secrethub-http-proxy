package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/keylockerbv/secrethub-go/pkg/secrethub"
	"github.com/keylockerbv/secrethub-proxy/pkg/restproxy"
)

var (
	credential           string
	credentialPassphrase string
	port                 int
	host                 string
	client               secrethub.Client
)

func init() {
	flag.StringVar(&credential, "C", "", "(Required) SecretHub credential")
	flag.StringVar(&credentialPassphrase, "P", "", "Passphrase to unlock SecretHub credential")
	flag.IntVar(&port, "p", 8080, "Port to listen on")
	flag.StringVar(&host, "h", "localhost", "Host to listen on")
	flag.Parse()

	if credential == "" {
		flag.Usage()
		exit(fmt.Errorf("credential is required"))
	}

	cred, err := secrethub.NewCredential(credential, credentialPassphrase)
	if err != nil {
		exit(err)
	}

	client = secrethub.NewClient(cred, nil)
}

func main() {
	proxy := restproxy.NewRESTProxy(client, host, port)

	go gracefulShutdown(proxy)

	log("SecretHub REST proxy started, press ^C to stop it")
	err := proxy.Start()
	if err != nil && err != http.ErrServerClosed {
		exit(err)
	}
}

func gracefulShutdown(proxy restproxy.ClientProxy) {
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, os.Interrupt)
	signal.Notify(sigint, syscall.SIGTERM)
	<-sigint

	log("Shutting down gracefully...")
	err := proxy.Stop()
	if err != nil {
		exit(err)
	}
}

func exit(err error) {
	fmt.Printf("secrethub-proxy: error: %v\n", err)
	os.Exit(1)
}

func log(message string) {
	fmt.Printf("secrethub-proxy: %v\n", message)
}
