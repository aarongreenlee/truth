package truth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
)

type (
	Runner func(t *testing.T, md Definition, tc TestCase) error
)

// RunIntegrationTests runs integration or full-stack tests using the provided metadata and test cases.
// Provide a client to perform full-stack. If nil is provided the server's Mux will be called directly.
func RunIntegrationTests(t *testing.T, def Definition, cases TestCases, c *Client) error {

	cases.init(def, getCaller(2))

	for _, tc := range cases {
		if err := NewRunner(c)(t, def, *tc); err != nil {
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
var muxUnderTest http.Handler

// SetMux allows the mux under test to be access by the truth test harness.
func SetMux(mux http.Handler) {
	muxUnderTest = mux
}

var printTestRuns, verbose bool

func TogglePrintAsTestsRun() {
	printTestRuns = !printTestRuns
}

func ToggleVerbose() {
	printTestRuns = true
	verbose = !verbose
}

// NewRunner builds a function to test an API endpoint. Provide a client to
// perform a full-stack call to a webserver. Without a client the server MUX
// will be called directly to perform the test in-process.
func NewRunner(c *Client) Runner {
	return func(t *testing.T, def Definition, tc TestCase) error {

		print := (verbose || tc.Verbose)

		if printTestRuns {
			fmt.Printf("Running %#v\n", tc.alias)
		}

		// Basic sanity check that the metadata is valid.
		if err := preflight(def, tc.Path); err != nil {
			return fmt.Errorf("%s: Preflight failed: %s", tc.alias, err.Error())
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
				return fmt.Errorf("%s: Unable to make HTTP request: %s", tc.alias, err.Error())
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

			if print {
				fmt.Printf("Calling the server mix call to `%s:%s`\n", req.Method, req.RequestURI)
			}

			muxUnderTest.ServeHTTP(RR, req)

			body, err = ioutil.ReadAll(RR.Body)
			if err != nil {
				return fmt.Errorf("%s: Unable to read response from Response Recorder: %s", tc.alias, err.Error())
			}
		}

		// Default to a 200 OK Expectation
		if tc.Status == 0 {
			tc.Status = 200
		}

		if RR.Code != tc.Status {
			t.Errorf("%s: Expected statuscode %d but received %d at `%s:%s`", tc.alias, tc.Status, RR.Code, def.Method, tc.Path)
			return nil
		}

		// Do we have an exact response we expect?
		// If so, we won't bother with any deeper testing of the body than this exact match check.
		if tc.ExpectBody != nil {
			if len(body) == 0 {
				t.Errorf("%s: Empty response body when a error response was expected", tc.alias)
				return nil
			}

			if 0 != bytes.Compare(body, tc.ExpectBody) {
				t.Fatalf("%s: Response was not an exact match", tc.alias)
				return nil
			}

			return nil
		}

		// TODO Should we parse test cases in advance and then search the byte array to avoid
		// converting  the body to a string for each test run?
		if len(tc.Contains) > 0 {
			content := string(body)
			for i, q := range tc.Contains {
				if !strings.Contains(content, q) {
					t.Errorf("%s: Response body did not contain search term #%d %#v", tc.alias, i, q)
				}
			}
		}

		if tc.Result != nil {
			// TODO Use the decoders
			if err := json.Unmarshal(body, &tc.Result); err != nil {
				t.Fatalf("%s: Unable to decode response into Result %#T", tc.Result)
				tc.Result = nil
				return nil
			}
		}

		if tc.Integration != nil {
			tc.Integration(Integration{
				T:      t,
				TC:     tc,
				Body:   body,
				RR:     RR,
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

	//if def.MIMETypeRequest == "" {
	//	return errors.New("MIMETypeRequest is not defined in metadata")
	//}
	//
	//if def.MIMETypeResponse == "" {
	//	return errors.New("MIMETypeResponse is not defined in metadata")
	//}

	return nil
}

func getCaller(depth int) string {
	gopath, gok := os.LookupEnv("GOPATH")

	_, file, line, ok := runtime.Caller(depth)

	if ok {
		if gok {
			file = strings.Replace(file, gopath+"/src/", "", 1)
		}
		return fmt.Sprintf("'%s:%d'", file, line)
	}

	return ""
}
