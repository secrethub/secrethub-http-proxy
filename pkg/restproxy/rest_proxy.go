package restproxy

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/keylockerbv/secrethub-go/pkg/secrethub"
	"github.com/keylockerbv/secrethub/api"
	"github.com/keylockerbv/secrethub/core/errio"
)

// SecretHubProxy gives the SecretHub Client a certain communication layer
type SecretHubProxy interface {
	Start() error
	Stop() error
}

type secretHubRESTProxy struct {
	SecretHubProxy
	client secrethub.Client
	server *http.Server
}

// NewSecretHubRESTProxy creates a proxy of the SecretHub Client, giving it a RESTful interface
func NewSecretHubRESTProxy(client secrethub.Client, port int) SecretHubProxy {
	if port == 0 {
		port = 8080
	}

	router := mux.NewRouter()
	proxy := &secretHubRESTProxy{
		client: client,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%v", port),
			Handler: router,
		},
	}
	proxy.addRoutes(router)

	return proxy
}

func (proxy *secretHubRESTProxy) addRoutes(r *mux.Router) {
	v1 := r.PathPrefix("/v1/").Subrouter()

	v1.PathPrefix("/secrets/").Handler(
		http.StripPrefix("/v1/secrets/", http.HandlerFunc(proxy.handleSecret)),
	)
}

// Start starts the SecretHub REST proxy, starting an HTTP server
func (proxy *secretHubRESTProxy) Start() error {
	return proxy.server.ListenAndServe()
}

// Stop stops the SecretHub REST proxy, stopping the HTTP server
func (proxy *secretHubRESTProxy) Stop() error {
	return proxy.server.Shutdown(context.Background())
}

func (proxy *secretHubRESTProxy) handleSecret(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	err := api.ValidateSecretPath(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	switch r.Method {
	case "GET":
		secret, err := proxy.client.Secrets().Versions().GetWithData(path)
		if err != nil {
			var errCode int

			if err, ok := err.(errio.PublicStatusError); ok {
				errCode = err.StatusCode
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

		_, err = proxy.client.Secrets().Write(path, secret)
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
