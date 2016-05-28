package truth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type (
	Client struct {
		Hostname string
	}
)

func NewClient(hostname string) *Client {
	return &Client{
		Hostname: hostname,
	}
}

// MakeRequest dispatches an HTTP request. If a payload is provided the response body will be read. If the returned
// status code is 2XX the response will be unmarshalled into the result. Provide a `nil` result to manage unmarshalling
// manually.
func (c Client) MakeRequest(md Definition, tc TestCase, result interface{}) (*http.Response, []byte, error) {

	req, err := c.BuildRequest(md, tc)
	if err != nil {
		return nil, nil, err
	}

	rsp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, nil, err
	}

	if err != nil || result == nil {
		body, parseErr := ioutil.ReadAll(rsp.Body)
		if parseErr != nil {
			return rsp, nil, parseErr
		}
		return rsp, body, err
	}

	body, err := c.ParseResponse(rsp, result)
	return rsp, body, err
}

// NewRequest builds an HTTP Request for tests using the provided metadata. If non-empty string is
// provided for the path it will be used instead of the metadata's path which is useful for route params
// embedded within the path such as:
// 	/users/:name
// Optionally, a payload and headers may be provided.
func (c Client) BuildRequest(def Definition, tc TestCase) (*http.Request, error) {

	var body io.Reader
	var err error

	if tc.Payload != nil {
		body, err = encode(tc.Payload)
		if err != nil {
			fmt.Printf("Error encoding payload. Unable to build HTTP request.\n")
			return nil, err
		}
	}

	addr := c.Hostname + def.Path

	if tc.Path != "" {
		// Allow the test case to override the URL
		addr = c.Hostname + tc.Path
	}

	if tc.Verbose {
		fmt.Printf("%s: Building request for `%s:%s`\n", tc.Name, def.Method, addr)
	}

	req, err := http.NewRequest(def.Method, addr, body)
	if err != nil {
		fmt.Printf("Error building request: %s\n", err)
		return req, err
	}

	headers := map[string]string{}

	if tc.Headers != nil {
		headers = copyHeaders(tc.Headers)
	}

	// TODO Expand for the other encoders
	if def.MIMETypeRequest != MIMETypeJSON {
		headers["Content-Type"] = def.MIMETypeRequest + "+json"
	}
	if def.MIMETypeResponse != MIMETypeJSON {
		headers["Accept"] = def.MIMETypeResponse + "json"
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func (c Client) ParseResponse(r *http.Response, result interface{}) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return body, json.Unmarshal(body, result)
}

func encode(v interface{}) (io.Reader, error) {
	if v == nil {
		return nil, nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(b), nil
}

// buildURLPath returns a string which concatenates a + b. The returned string will always
// begin with a "/" and a "/" will be inserted between "a" and "b" if one was not provided on either "a" or "b".
func buildURLPath(a, b string) string {
	// Ensure the namespace starts with a "/" and does not end with a "/"
	if len(a) > 0 && !strings.HasPrefix(a, "/") {
		a = "/" + a
	}
	if len(a) > 0 && strings.HasSuffix(a, "/") {
		a = a[:len(a)-1]
	}

	// Ensure the b starts with a "/"
	if len(b) > 0 && !strings.HasPrefix(b, "/") {
		b = "/" + b
	}

	return a + a
}

func copyHeaders(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))

	for k, v := range m {
		out[k] = v
	}

	return out
}
