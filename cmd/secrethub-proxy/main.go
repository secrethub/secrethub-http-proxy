package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/keylockerbv/secrethub-clientd/pkg/restproxy"
	"github.com/keylockerbv/secrethub-go/pkg/secrethub"
)

var (
	credential           string
	credentialPassphrase string
	port                 int
	client               secrethub.Client
)

func init() {
	flag.StringVar(&credential, "C", "", "(Required) SecretHub credential")
	flag.StringVar(&credentialPassphrase, "P", "", "Passphrase to unlock SecretHub credential")
	flag.IntVar(&port, "p", 8080, "HTTP port to listen on")
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
	clientd := restproxy.SecretHubRESTProxy{
		Client: &client,
		Port:   port,
	}
	fmt.Println("SecretHub REST proxy started, press ^C to stop it")
	err := clientd.Start()
	if err != nil {
		exit(err)
	}
}

func exit(err error) {
	fmt.Printf("secrethub-clientd: error: %v\n", err)
	os.Exit(1)
}
