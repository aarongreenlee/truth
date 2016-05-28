package truth

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

type (
	TestCases []*TestCase

	// TestCase structures a specific test which can be applied unit, integration, full-stack, or load testing.
	TestCase struct {
		Name       string
		Path       string
		Headers    map[string]string
		Payload    interface{}
		Status     int
		ExpectBody []byte
		Contains   []string

		Verbose     bool
		Integration func(Integration)
		Unit        func(Unit)

		alias string // Used for test failure messages
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

func (tc *TestCase) init(def Definition, n int, caller string) {
	if tc.Path == "" {
		tc.Path = def.Path
	}

	if tc.Name == "" {
		tc.Name = fmt.Sprintf("#%d for '%s:%s' called from %s", n, def.Method, def.Path, caller)
	}

	tc.alias = "Testcase: " + tc.Name
}

func (cases TestCases) init(def Definition, caller string) {
	for i, tc := range cases {
		tc.init(def, i, caller)
	}
}
