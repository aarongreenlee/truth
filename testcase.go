package truth

import (
	"net/http/httptest"
	"testing"
)

type (
	// TestCase structures a specific test which can be applied unit, integration, full-stack, or load testing.
	TestCase struct {
		Name        string
		Path        string
		Headers     map[string]string
		Payload     interface{}
		Status      int
		ExpectBody  []byte
		Contains    []string

		Verbose bool
		Integration func(Integration)
		Unit        func(Unit)
	}

	Integration struct {
		*testing.T
		TC     TestCase
		Body   []byte
		RR     *httptest.ResponseRecorder
		Client *Client
		Runner Runner
	}

	Unit struct {
		*testing.T
		TC     TestCase
		Result interface{}
		Err    error
	}
)

