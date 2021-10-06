package connhttp

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/connctd/connector-go"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type ConnectorService interface {
	AddInstallation(ctx context.Context, req connector.InstallationRequest) error
	AddInstance(ctx context.Context, req connector.InstantiationRequest) error
	PerformAction(ctx context.Context, req connector.ActionRequest) (*connector.ActionResponse, error)
}

type ConnectorHandler struct {
	r       *mux.Router
	service ConnectorService
	logger  logrus.FieldLogger
}

func NewConnectorHandler(subrouter *mux.Router, service ConnectorService, host string, publicKey ed25519.PublicKey) *ConnectorHandler {
	c := &ConnectorHandler{
		r:       subrouter,
		service: service,
		logger:  logrus.WithField("component", "connectorhandler"),
	}
	if c.r == nil {
		c.r = mux.NewRouter()
	}
	c.r.Path("/installations").Methods(http.MethodPost).Handler(connector.NewSignatureValidationHandler(
		connector.ProxiedRequestValidationPreProcessor("https", host, "/connector/installations"), publicKey, c.addInstallation))
	c.r.Path("/instances").Methods(http.MethodPost).Handler(connector.NewSignatureValidationHandler(
		connector.ProxiedRequestValidationPreProcessor("https", host, "/connector/instances"), publicKey, c.addInstance))
	c.r.Path("/actions").Methods(http.MethodPost).Handler(connector.NewSignatureValidationHandler(
		connector.ProxiedRequestValidationPreProcessor("https", host, "/connector/actions"), publicKey, c.performAction))
	return c
}

func (c *ConnectorHandler) addInstallation(w http.ResponseWriter, r *http.Request) {
	logger := c.logger.WithFields(logrus.Fields{
		"client": r.RemoteAddr,
		"url":    r.URL.String(),
		"method": "addInstallation",
	})
	var req connector.InstallationRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		logger.WithError(err).Error("Failed to decode client request")
		writeError(w, err)
		return
	}

	if err := c.service.AddInstallation(r.Context(), req); err != nil {
		logger.WithError(err).Error("Failed to add installation")
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (c *ConnectorHandler) addInstance(w http.ResponseWriter, r *http.Request) {
	logger := c.logger.WithFields(logrus.Fields{
		"client": r.RemoteAddr,
		"url":    r.URL.String(),
		"method": "addInstance",
	})
	var req connector.InstantiationRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		logger.WithError(err).Error("Failed to decode client request")
		writeError(w, err)
		return
	}

	if err := c.service.AddInstance(r.Context(), req); err != nil {
		logger.WithError(err).Error("Failed to add instance")
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (c *ConnectorHandler) performAction(w http.ResponseWriter, r *http.Request) {
	logger := c.logger.WithFields(logrus.Fields{
		"client": r.RemoteAddr,
		"url":    r.URL.String(),
		"method": "performAction",
	})
	var req connector.ActionRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		logger.WithError(err).Error("Failed to decode client request")
		writeError(w, err)
		return
	}

	resp, err := c.service.PerformAction(r.Context(), req)
	if err != nil {
		logger.WithError(err).Error("Failed to perform action")
		writeError(w, err)
		return
	}
	resp.ID = req.ID
	b, err := json.Marshal(resp)
	if err != nil {
		logger.WithError(err).Error("failed to marshal action error")
		writeError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// helps to decode the request body
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dest interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return ErrorBadContentType
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ErrorBadRequestBody
	}

	if err = json.Unmarshal(body, dest); err != nil {
		return ErrorInvalidJsonBody
	}

	return nil
}

func writeError(w http.ResponseWriter, err error) {
	var e Error
	if errors.As(err, &e) {
		e.WriteBody(w)
	} else {
		e = Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}

		e.WriteBody(w)
	}
}
