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

		Result interface{}

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

func (tc *TestCase) init(def Definition, n, count int, caller string) {
	if tc.Path == "" {
		tc.Path = def.Path
	}

	if tc.Name == "" {
		tc.Name = fmt.Sprintf("'%s:%s' (%d of %d) called from %s", def.Method, def.Path, n+1, count+1, caller)
	}

	tc.alias = "Testcase: " + tc.Name
}

func (cases TestCases) init(def Definition, caller string) {
	for i, tc := range cases {
		tc.init(def, i, len(cases), caller)
	}
}
