package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/secrethub/secrethub-http-proxy/pkg/restproxy"

	"github.com/secrethub/secrethub-go/pkg/secrethub"

	"golang.org/x/crypto/ssh/terminal"
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
	flag.StringVar(&host, "h", "127.0.0.1", "Host to listen on")
	flag.Parse()

	if credential == "" {
		flag.Usage()
		exit(fmt.Errorf("credential is required"))
	}

	cred, err := findCredential(credential, credentialPassphrase)
	if err != nil {
		exit(err)
	}

	client = secrethub.NewClient(cred, nil)
}

func findCredential(credential string, passphrase string) (secrethub.Credential, error) {
	parser := secrethub.NewCredentialParser(secrethub.DefaultCredentialDecoders)

	encoded, err := parser.Parse(credential)
	if err != nil {
		return nil, err
	}

	if encoded.IsEncrypted() {
		if passphrase == "" {
			passphrase, err = promptPassword()
			if err != nil {
				return nil, err
			}
		}

		key, err := secrethub.NewPassBasedKey([]byte(passphrase))
		if err != nil {
			return nil, err
		}

		credential, err := encoded.DecodeEncrypted(key)
		if err != nil {
			return nil, err
		}

		return credential, err
	}

	return encoded.Decode()
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

func promptPassword() (string, error) {
	fmt.Printf("Please put in the passphrase to unlock your credential:")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}

	return string(password), nil
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
	fmt.Printf("secrethub-http-proxy: error: %v\n", err)
	os.Exit(1)
}

func log(message string) {
	fmt.Printf("secrethub-http-proxy: %v\n", message)
}
