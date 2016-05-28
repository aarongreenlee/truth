package main

import (
	"github.com/aarongreenlee/truth"
	"strings"
	"testing"
)

// SetupTest sets up the application under test. Any dependencies should be loaded by this
// "bootstrap" call and the entire application under test should be ready with the
// exception of actually listening on a port. We won't need to listen on a port
// for integration testing since we're calling the ServeMux directly.
func SetupTest() {
	configure()
	truth.SetMux(mux)
}

// TestHelloWorld will run two test cases. They should both pass.
func TestHelloWorld(t *testing.T) {
	SetupTest()

	// Define our handler under test.
	def := truth.Definition{
		Method: "GET",
		Path:   "/helloworld",
	}

	tests := truth.TestCases{
		// This is a simple example. We name the test, verify the
		// response equals our expectation exactly, and we also verify
		// a `200 OK` response code (by default).
		{
			Name:       "Response body should equal",
			ExpectBody: []byte("Hello world!"),
		},
		// We can also make a test that we don't have to configure
		// anything for! This example test uses the definition to
		// figure out which HTTP method and path to call, and even
		// lets the Truth test harness name the test case something
		// like:
		//
		// `Testcase: #2 for 'GET:/helloworld' called from {some/file:66}`.
		//
		// This test basically confirms we have a `200 OK ` status code
		// when we call the endpoint defined in the Definition.
		{},
		// The previous example uses the Definition to tell our test which
		// HTTP route we want to test. Here, we override the path and also
		// specify the status code we expect.
		{
			// Example using a custom URL.
			Name:   "`/helloworld-bad-spelling` should not be found",
			Path:   "/helloworld-bad-spelling",
			Status: 404,
		},
		// Like the previous test but we expect a 200 OK this time.
		{
			// Example using a custom URL.
			Name:   "Custom URL example",
			Path:   "/helloworld?abc",
			Status: 200,
		},
		// We can also search the response body for specific strings.
		// This technique can be a quick and simple way to verify you've
		// got the response you are looking for without worrying about
		// providing custom types and decoding the response from the
		// server.
		{
			Name:     "Response should contain `Hello`, `world`, and `!`",
			Contains: []string{"Hello", "world", "!"},
		},
		// In this example we use the Integration function to perform some
		// custom validation. This allows you to do anything you want!
		// This example is quite simple. Checkout some of the other tests for
		// more complex demonstrations.
		{
			Name: "Response does not contain `Good bye`",
			Integration: func(t truth.Integration) {
				goodbye := "Good bye"
				if strings.Contains(string(t.Body), goodbye) {
					t.T.Errorf("Endpoint said %#v!", goodbye)
				}
			},
		},
	}

	// Print some basic output as the tests run.
	truth.TogglePrintAsTestsRun()

	// Run the tests!
	truth.RunIntegrationTests(t, def, tests, nil)
}
