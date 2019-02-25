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
	r := mux.NewRouter()
	r.PathPrefix("/secrets/").Handler(
		http.StripPrefix("/secrets/", http.HandlerFunc(handleSecret)),
	)

	fmt.Println("SecretHub Clientd started, press ^C to exit")
	panic(http.ListenAndServe(address, r))
}

func handleSecret(res http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	err := api.ValidateSecretPath(path)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		io.WriteString(res, err.Error())
		return
	}

	switch req.Method {
	case "GET":
		sec, err := client.Secrets().Versions().GetWithData(path)
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

			res.WriteHeader(errCode)
			io.WriteString(res, err.Error())
			return
		}

		res.WriteHeader(http.StatusOK)
		res.Write(sec.Data)
		return
	case "POST":
		secret, err := ioutil.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			io.WriteString(res, err.Error())
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

			res.WriteHeader(errCode)
			io.WriteString(res, err.Error())
			return
		}

		res.WriteHeader(http.StatusCreated)
	default:
		res.Header().Add("Allow", "GET, POST")
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}
