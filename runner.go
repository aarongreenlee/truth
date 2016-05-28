package truth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"fmt"
	"errors"
	"bytes"
	"io/ioutil"
	"strings"
	"github.com/stretchr/testify/assert"
)

type (
	Runner func(t *testing.T, md Definition, tc TestCase) error
)

// RunIntegrationTests runs integration or full-stack tests using the provided metadata and test cases.
// Provide a client to perform full-stack. If nil is provided the server's Mux will be called directly.
func RunIntegrationTests(t *testing.T, md Definition, cases []TestCase, c *Client) error {
	for _, tc := range cases {
		if err := NewRunner(c)(t, md, tc); err != nil {
			t.Error(err)
			t.FailNow()

			fmt.Printf("%s: Error running integration tests: %s\n", tc.Name, err)
			return err
		}
	}
	return nil
}

var integrationClient *Client
func init() {
	integrationClient = NewClient("")
}

// Mux under test
var muxUnderTest *http.ServeMux

// SetMux allows the mux under test to be access by the truth test harness.
func SetMux(mux *http.ServeMux) {
	muxUnderTest = mux
}

// NewRunner builds a function to test an API endpoint. Provide a client to perform a full-stack call
// to a webserver. Without a client the server MUX will be called directly to perform the test in-process.
func NewRunner(c *Client) Runner {
	return func(t *testing.T, def Definition, tc TestCase) error {
		// Simplify reference to failures
		alias := "Testcase: " + tc.Name

		// Basic sanity check that the metadata is valid.
		if err := preflight(def, tc.Path); err != nil {
			return fmt.Errorf("%s: Preflight failed: %s", alias, err.Error())
		}

		// Prepare to execute the request.
		RR := httptest.NewRecorder()
		var body []byte

		// If we have a client we're going to perform a full HTTP test.
		if c != nil {
			var err error
			var rsp *http.Response
			rsp, body, err = c.MakeRequest(def, tc, nil)
			if err != nil {
				return fmt.Errorf("%s: Unable to make HTTP request: %s", alias, err.Error())
			}
			// Copy the response into recorder
			RR.Code = rsp.StatusCode
			RR.Body = bytes.NewBuffer(body)
			for k, v := range rsp.Header {
				RR.HeaderMap[k] = v
			}
		} else {
			var err error
			var req *http.Request

			req, err = integrationClient.BuildRequest(def, tc)
			if err != nil {
				return err
			}

			if tc.Verbose {
				fmt.Printf("Simulating call to `%s:%s`\n", req.Method, req.RequestURI)
			}

			muxUnderTest.ServeHTTP(RR, req)

			body, err = ioutil.ReadAll(RR.Body)
			if err != nil {
				return fmt.Errorf("%s: Unable to read response from Response Recorder: %s", alias, err.Error())
			}
		}

		if RR.Code != tc.Status {
			t.Errorf("%s: Expected statuscode %d but received %d", alias, tc.Status, RR.Code)
			return nil
		}

		// Do we have an exact response we expect?
		// If so, we won't bother with any deeper testing of the body than this exact match check.
		if tc.ExpectBody != nil {
			if len(body) == 0 {
				t.Errorf("%s: Empty response body when a error response was expected", alias)
				return nil
			}

			if bytes.Compare(body, tc.ExpectBody) != 0 {
				assert.EqualValues(t, string(tc.ExpectBody), string(body), fmt.Sprintf("%s: Response was not an exact match", alias))
			}

			return nil
		}

		// TODO Should we parse test cases in advance and then search the byte array to avoid
		// converting  the body to a string for each test run?
		if len(tc.Contains) > 0 {
			content := string(body)
			for i, q := range tc.Contains {
				if !strings.Contains(content, q) {
					t.Errorf("%s: Response body did not contain search term #%d %#v", alias, i, q)
				}
			}
		}

		if tc.Integration != nil {
			tc.Integration(Integration{
				T:    t,
				TC:   tc,
				Body: body,
				RR:   RR,
				Client: c,
			})
		}

		return nil
	}
}

func preflight(def Definition, path string) error {
	switch def.Method {
	case http.MethodPost, http.MethodConnect, http.MethodDelete, http.MethodGet, http.MethodHead,
		http.MethodOptions, http.MethodPatch, http.MethodTrace, http.MethodPut:
		// Do nothing
	default:
		return fmt.Errorf("HTTP method %#v is not supported", def.Method)
	}

	if def.MIMETypeRequest == "" {
		return errors.New("MIMETypeRequest is not defined in metadata")
	}

	if def.MIMETypeResponse == "" {
		return errors.New("MIMETypeResponse is not defined in metadata")
	}

	return nil
}