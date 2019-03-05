package restproxy

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/keylockerbv/secrethub-go/pkg/secrethub"
	"github.com/keylockerbv/secrethub/api"
	"github.com/keylockerbv/secrethub/core/errio"
)

// SecretHubRESTProxy exposes SecretHub Client functionality with a RESTful interface
type SecretHubRESTProxy struct {
	Port   int
	Client *secrethub.Client
}

// Start starts the SecretHub REST proxy
func (c *SecretHubRESTProxy) Start() error {
	mux := mux.NewRouter()
	v1 := mux.PathPrefix("/v1/").Subrouter()

	v1.PathPrefix("/secrets/").Handler(
		http.StripPrefix("/v1/secrets/", http.HandlerFunc(c.handleSecret)),
	)

	return http.ListenAndServe(fmt.Sprintf(":%v", c.Port), mux)
}

func (c *SecretHubRESTProxy) handleSecret(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	err := api.ValidateSecretPath(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	switch r.Method {
	case "GET":
		secret, err := (*c.Client).Secrets().Versions().GetWithData(path)
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

		_, err = (*c.Client).Secrets().Write(path, secret)
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
