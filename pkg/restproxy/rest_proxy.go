package restproxy

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/secrethub/secrethub-go/internals/api"
	"github.com/secrethub/secrethub-go/internals/errio"
	"github.com/secrethub/secrethub-go/pkg/secrethub"
)

// ClientProxy gives the SecretHub Client a certain communication layer
type ClientProxy interface {
	Start() error
	Stop() error
}

type restProxy struct {
	ClientProxy
	client secrethub.Client
	server *http.Server
}

// NewRESTProxy creates a proxy for the SecretHub Client, giving it a RESTful interface
func NewRESTProxy(client secrethub.Client, host string, port int) ClientProxy {
	if port == 0 {
		port = 8080
	}

	router := mux.NewRouter()
	proxy := &restProxy{
		client: client,
		server: &http.Server{
			Addr:    fmt.Sprintf("%v:%d", host, port),
			Handler: router,
		},
	}
	proxy.addRoutes(router)

	return proxy
}

func (p *restProxy) addRoutes(r *mux.Router) {
	v := r.PathPrefix("/v1beta/").Subrouter()

	v.PathPrefix("/secrets/").Handler(
		http.StripPrefix("/v1beta/secrets/raw/", http.HandlerFunc(p.handleSecret)),
	)
}

// Start starts the SecretHub REST proxy, starting an HTTP server
func (p *restProxy) Start() error {
	return p.server.ListenAndServe()
}

// Stop stops the SecretHub REST proxy, stopping the HTTP server
func (p *restProxy) Stop() error {
	return p.server.Shutdown(context.Background())
}

func (p *restProxy) handleSecret(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	err := api.ValidateSecretPath(path)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		secret, err := p.client.Secrets().Versions().GetWithData(path)
		if err != nil {
			writeError(w, err, 0)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(secret.Data)
	case "POST":
		secret, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeError(w, err, http.StatusInternalServerError)
			return
		}

		_, err = p.client.Secrets().Write(path, secret)
		if err != nil {
			statusCode := 0
			switch err {
			case secrethub.ErrCannotWriteToVersion,
				secrethub.ErrEmptySecret,
				secrethub.ErrSecretTooBig:
				statusCode = http.StatusBadRequest
			}

			writeError(w, err, statusCode)
			return
		}

		w.WriteHeader(http.StatusCreated)
	case "DELETE":
		err := p.client.Secrets().Delete(path)
		if err != nil {
			writeError(w, err, 0)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		w.Header().Add("Allow", "GET, POST, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// writeError writes an error message and HTTP status code to the ResponseWriter.
// The HTTP status code is derrived from the error, unless overriden by the statusCode argument.
func writeError(w http.ResponseWriter, err error, statusCode int) {
	if statusCode == 0 {
		if err, ok := err.(errio.PublicStatusError); ok {
			statusCode = err.StatusCode
		}

		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}
	}

	w.WriteHeader(statusCode)
	io.WriteString(w, err.Error())
}
