package connector

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/connctd/api-go"
	"github.com/connctd/api-go/crypto"
)

type signatureValidationHandler struct {
	preProcessor ValidationPreProcessor
	next         http.HandlerFunc
	publicKey    ed25519.PublicKey
}

// NewSignatureValidationHandler creates a new handler capable of verifying the signature header
// Validation can be influenced by passing a ValidationPreProcessor. Quite common
// functionalities are offered by DefaultValidationPreProcessor and ProxiedRequestValidationPreProcessor
func NewSignatureValidationHandler(validationPreProcessor ValidationPreProcessor, publicKey ed25519.PublicKey, next http.HandlerFunc) http.Handler {
	return &signatureValidationHandler{preProcessor: validationPreProcessor, publicKey: publicKey, next: next}
}

// ServeHTTP handles request
func (h *signatureValidationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get(crypto.SignatureHeaderKey)

	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		ErrorBadSignature.Write(w)
		return
	}

	var body []byte

	// in case body is given
	if r.ContentLength != 0 {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			ErrorInvalidBody.Write(w)
			return
		}

		defer r.Body.Close()
	}

	// apply preprocess and pass values to signing function
	extractedValues := h.preProcessor(r)
	expectedSignature, err := crypto.SignablePayload(r.Method, extractedValues.Scheme, extractedValues.Host, extractedValues.RequestURI, r.Header, body)
	if err != nil {
		if errors.Is(err, crypto.ErrorMissingHeader) {
			crypto.ErrorMissingHeader.Write(w)
			return
		}

		ErrorSigningFailed.Write(w)
		return
	}

	// lets check the signature manually
	if ed25519.Verify(h.publicKey, expectedSignature, decodedSignature) {
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		h.next.ServeHTTP(w, r)
	} else {
		ErrorBadSignature.Write(w)
		return
	}
}

// ValidationPreProcessor can be used to influence the signature validation algorithm
// by returning a modified url struct. This becomes handy if your service is sitting behind
// a proxy that modifies the original request headers which normally would lead to a
// validation error
type ValidationPreProcessor func(r *http.Request) ValidationParameters

// DefaultValidationPreProcessor extracts all relevant values from request fields
// Use this processor if there are no proxies between connctd platform and your connector
func DefaultValidationPreProcessor() ValidationPreProcessor {
	return func(r *http.Request) ValidationParameters {
		// on server side scheme is not populated: https://github.com/golang/go/issues/28940
		// manually picking https since this is currently the only supported protocol by connctd for callbacks
		return ValidationParameters{
			Scheme:     "https",
			Host:       r.Host,
			RequestURI: r.URL.RequestURI(),
		}
	}
}

// ProxiedRequestValidationPreProcessor allows passing modified headers to validate signature
// function. This is necessary in case received request headers do not match up with
// sent request headers because of e.g. proxies in between
func ProxiedRequestValidationPreProcessor(scheme string, host string, endpoint string) ValidationPreProcessor {
	return func(r *http.Request) ValidationParameters {
		return ValidationParameters{
			Scheme:     scheme,
			Host:       host,
			RequestURI: endpoint,
		}
	}
}

// ValidationParameters reflects list of parameters that are relevant for request signature validation
type ValidationParameters struct {
	Scheme     string
	Host       string
	RequestURI string
}

// some error definitions
var (
	ErrorBadSignature  = api.NewError("BAD_SIGNATURE", "Signature seems to be invalid", http.StatusBadRequest)
	ErrorSigningFailed = api.NewError("SIGNING_FAILED", "Failed to sign the request", http.StatusBadRequest)
	ErrorInvalidBody   = api.NewError("INVALID_BODY", "Unable to read message body", http.StatusBadRequest)
)
