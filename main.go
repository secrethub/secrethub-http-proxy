package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/keylockerbv/secrethub-go/pkg/api"
	"github.com/keylockerbv/secrethub-go/pkg/errio"
	"github.com/keylockerbv/secrethub-go/pkg/secrethub"
)

var (
	credential           string
	credentialPassphrase string
	address              string
	client               secrethub.Client
)

func init() {
	flag.StringVar(&credential, "c", "", "(Required) SecretHub credential")
	flag.StringVar(&credentialPassphrase, "p", "", "Passphrase to unlock SecretHub credential")
	flag.StringVar(&address, "a", ":8080", "HTTP server address")
	flag.Parse()

	if credential == "" {
		flag.Usage()
		os.Exit(1)
	}

	cred, err := secrethub.NewCredential(credential, credentialPassphrase)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client = secrethub.NewClient(cred, nil)
}

func main() {
	err := startHTTPServer()
	if err != nil {
		panic(err)
	}
}

func startHTTPServer() error {
	mux := mux.NewRouter()
	v1 := mux.PathPrefix("/v1/").Subrouter()

	v1.PathPrefix("/secrets/").Handler(
		http.StripPrefix("/v1/secrets/", http.HandlerFunc(handleSecret)),
	)

	fmt.Println("SecretHub Clientd started, press ^C to exit")
	return http.ListenAndServe(address, mux)
}

func handleSecret(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	err := api.ValidateSecretPath(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	switch r.Method {
	case "GET":
		secret, err := client.Secrets().Versions().GetWithData(path)
		if err != nil {
			var errCode int

			if err, ok := err.(errio.PublicStatusError); ok {
				errCode = err.StatusCode
			}

			switch err {
			case api.ErrSecretNotFound:
				errCode = http.StatusNoContent
			}

			if errCode == 0 {
				errCode = http.StatusInternalServerError
			}

			w.WriteHeader(errCode)
			io.WriteString(w, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(secret.Data)
	case "POST":
		secret, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}

		_, err = client.Secrets().Write(path, secret)
		if err != nil {
			var errCode int

			if err, ok := err.(errio.PublicStatusError); ok {
				errCode = err.StatusCode
			}

			switch err {
			case secrethub.ErrCannotWriteToVersion,
				secrethub.ErrEmptySecret,
				secrethub.ErrSecretTooBig:
				errCode = http.StatusBadRequest
			}

			if errCode == 0 {
				errCode = http.StatusInternalServerError
			}

			w.WriteHeader(errCode)
			io.WriteString(w, err.Error())
			return
		}

		w.WriteHeader(http.StatusCreated)
	default:
		w.Header().Add("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
