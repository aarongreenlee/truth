package main

import (
	"testing"
	"github.com/aarongreenlee/truth"
	"strings"
)

// SetupTest sets up the application under test. Any dependencies should be loaded by this
// "bootstrap" call and the entire application under test should be ready with the
// exception of actually listening on a port. We won't need to listen on a port
// for integration testing since we're calling the ServeMux directly.
func SetupTest() {
	bootstrapApp()
	truth.SetMux(mux)
}


// TestHelloWorld will run two test cases. They should both pass.
func TestHelloWorld(t *testing.T) {
	SetupTest()

	// Define our handler under test
	def := truth.Definition{
		Package: "main",
		Description: "A simple handler we can test",
		Name: "HelloWorld",
		MIMETypeRequest: "text/plain",
		MIMETypeResponse: "text/plain",
		Method: "GET",
		Path: "/helloworld",
	}

	happy := truth.TestCase{
		Name: "Hello World Positive Test",
		Status: 200,
		ExpectBody: []byte("Hello world!"),
		Verbose: true,
	}

	sad := truth.TestCase{
		Name: "Hello World Negative Test",
		Status: 200,
		Integration: func(t truth.Integration) {
			goodbye := "Good Bye World"
			if strings.Contains(string(t.Body), goodbye) {
				t.T.Errorf("Endpoint said %#v!", goodbye)
			}
		},
	}

	// We want to run both the happy and sad test
	tests := []truth.TestCase{happy, sad}

	truth.RunIntegrationTests(t, def, tests, nil)
}









