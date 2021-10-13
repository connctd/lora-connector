package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	stdlog "log"

	"path"

	"github.com/connctd/restapi-go"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

const (
	// APIBaseURL defines how to reach connctd api
	APIBaseURL = "https://connectors.connctd.io/api/v1/"

	connectorThingsEndpoint            = "connectorhub/callback/instances/things"
	connectorActionsEndpoint           = "connectorhub/callback/instances/actions/requests"
	connectorInstanceStateEndpoint     = "connectorhub/callback/instances/state"
	connectorInstallationStateEndpoint = "connectorhub/callback/installations/state"
)

// DefaultOptions returns default client options
func DefaultOptions() *ClientOptions {
	url, _ := url.Parse(APIBaseURL)

	return &ClientOptions{
		ConnctdBaseURL: url,
		HTTPClient:     http.DefaultClient,
	}
}

// DefaultLogger uses go standard logging capabilities
var DefaultLogger = stdr.New(stdlog.New(os.Stderr, "", stdlog.LstdFlags|stdlog.Lshortfile))

// Client defines api client functionalities
type Client interface {
	// CreateThing can be used to create a thing. A thingID is returned if
	// operation was successul. Otherwise an error is thrown.
	CreateThing(ctx context.Context, token InstantiationToken, thing restapi.Thing) (result restapi.Thing, err error)
	UpdateThingPropertyValue(ctx context.Context, token InstantiationToken, thingID string, componentID string, propertyID string, value string, lastUpdate time.Time) error
	UpdateThingStatus(ctx context.Context, token InstantiationToken, thingID string, status restapi.StatusType) error

	// UpdateActionStatus can be used to inform about the new state of an action. Optional err can be set for additional error details
	UpdateActionStatus(ctx context.Context, token InstantiationToken, actionRequestID string, status restapi.ActionRequestStatus, err string) error
	UpdateInstallationState(ctx context.Context, token InstallationToken, state InstallationState, details json.RawMessage) error
	UpdateInstanceState(ctx context.Context, token InstantiationToken, state InstantiationState, details json.RawMessage) error

	DeleteThing(ctx context.Context, token InstantiationToken, thingID string) error
}

// ClientOptions allow modification of api client behaviour
type ClientOptions struct {
	ConnctdBaseURL *url.URL
	HTTPClient     *http.Client
}

// APIClient implements Client interface
type APIClient struct {
	httpClient *http.Client
	baseURL    url.URL
	logger     logr.Logger
}

// NewClient creates a new api client
func NewClient(opts *ClientOptions, logger logr.Logger) (Client, error) {
	httpClient := http.DefaultClient
	url, _ := url.Parse(APIBaseURL)

	if logger == nil {
		return nil, ErrorMissingLogger
	}

	if opts != nil {
		if opts.HTTPClient != nil {
			httpClient = opts.HTTPClient
		}

		if opts.ConnctdBaseURL != nil {
			// url needs to end with slash
			if !strings.HasSuffix(opts.ConnctdBaseURL.String(), "/") {
				return nil, ErrorInvalidBaseURL
			}

			url = opts.ConnctdBaseURL
		}
	}

	return &APIClient{httpClient: httpClient, baseURL: *url, logger: logger.WithName("connector-go-client")}, nil
}

// CreateThing implements interface definition
func (a *APIClient) CreateThing(ctx context.Context, token InstantiationToken, thing restapi.Thing) (result restapi.Thing, err error) {
	message := AddThingRequest{
		Thing: thing,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return restapi.Thing{}, fmt.Errorf("failed to marshal thing: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.baseURL.String()+connectorThingsEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return restapi.Thing{}, fmt.Errorf("failed to create new request: %w", err)
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+string(token))

	resp, err := a.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		a.logger.Error(err, "Failed to create thing", "name", thing.Name)
		return restapi.Thing{}, fmt.Errorf("failed to create thing: %w", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return restapi.Thing{}, fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		a.logger.Error(ErrorUnexpectedStatusCode, "Could not create thing", "expectedStatusCode", http.StatusCreated, "givenStatusCode", resp.StatusCode, "body", string(body))
		return restapi.Thing{}, ErrorUnexpectedStatusCode
	}

	var res AddThingResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return restapi.Thing{}, ErrorUnexpectedResponse
	}

	thing.ID = res.ID
	return thing, nil
}

// UpdateThingPropertyValue implements interface definition
func (a *APIClient) UpdateThingPropertyValue(ctx context.Context, token InstantiationToken, thingID string, componentID string, propertyID string, value string, lastUpdate time.Time) error {
	message := UpdateThingPropertyValueRequest{
		Value:      value,
		LastUpdate: lastUpdate,
	}

	return a.doRequest(ctx, http.MethodPut, path.Join(connectorThingsEndpoint, thingID, "components", componentID, "properties", propertyID), string(token), message, http.StatusNoContent)
}

// UpdateThingPropertyValue implements interface definition
func (a *APIClient) UpdateThingStatus(ctx context.Context, token InstantiationToken, thingID string, status restapi.StatusType) error {
	message := UpdateThingStatusRequest{
		Status: status,
	}

	return a.doRequest(ctx, http.MethodPut, path.Join(connectorThingsEndpoint, thingID, "status"), string(token), message, http.StatusNoContent)
}

// UpdateActionStatus implements interface definition
func (a *APIClient) UpdateActionStatus(ctx context.Context, token InstantiationToken, actionRequestID string, status restapi.ActionRequestStatus, e string) error {
	message := ActionRequestStatusUpdate{
		Status: status,
		Error:  e,
	}

	return a.doRequest(ctx, http.MethodPut, path.Join(connectorActionsEndpoint, actionRequestID), string(token), message, http.StatusNoContent)
}

// UpdateInstallationState implements interface definition
func (a *APIClient) UpdateInstallationState(ctx context.Context, token InstallationToken, state InstallationState, details json.RawMessage) error {
	message := InstallationStateUpdateRequest{
		State:   state,
		Details: details,
	}

	return a.doRequest(ctx, http.MethodPost, connectorInstallationStateEndpoint, string(token), message, http.StatusNoContent)
}

// UpdateThingPropertyValue implements interface definition
func (a *APIClient) UpdateInstanceState(ctx context.Context, token InstantiationToken, state InstantiationState, details json.RawMessage) error {
	message := InstanceStateUpdateRequest{
		State:   state,
		Details: details,
	}

	return a.doRequest(ctx, http.MethodPost, connectorInstanceStateEndpoint, string(token), message, http.StatusNoContent)
}

// UpdateThingPropertyValue implements interface definition
func (a *APIClient) DeleteThing(ctx context.Context, token InstantiationToken, thingID string) error {
	return a.doRequest(ctx, http.MethodDelete, path.Join(connectorThingsEndpoint, thingID), string(token), nil, http.StatusNoContent)
}

func (a *APIClient) doRequest(ctx context.Context, method string, endpoint string, token string, payload interface{}, expectedStatusCode int) error {
	logger := a.logger.WithValues("endpoint", endpoint, "expectedStatusCode", expectedStatusCode)

	var err error
	var req *http.Request

	// append payload if given
	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			logger.Error(err, "Failed to marshal request")
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err = http.NewRequest(method, a.baseURL.String()+endpoint, bytes.NewBuffer(payloadBytes))
		if err != nil {
			logger.Error(err, "Failed to create new request")
			return fmt.Errorf("failed to create new request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, a.baseURL.String()+endpoint, nil)
		if err != nil {
			logger.Error(err, "Failed to create new request")
			return fmt.Errorf("failed to create new request: %w", err)
		}
	}

	// set additional headers
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := a.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		logger.Error(err, "Failed to send request")
		return fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err, "Failed to read response body")
		return fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode != expectedStatusCode {
		logger.Error(ErrorUnexpectedStatusCode, "Could not update thing property", "givenStatusCode", resp.StatusCode, "body", string(body))
		return ErrorUnexpectedStatusCode
	}

	return nil
}

// some error defintions
var (
	ErrorInvalidBaseURL       = errors.New("the base url needs to end with a slash")
	ErrorMissingLogger        = errors.New("a logger needs to be passed")
	ErrorUnexpectedStatusCode = errors.New("the resulting status code does not match with expectation")
	ErrorUnexpectedResponse   = errors.New("remote site replied with unexpected contents")
)
