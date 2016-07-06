package main

import "github.com/aarongreenlee/truth"

// SetupTest sets up the application under test. Any dependencies should be loaded by this
// "bootstrap" call and the entire application under test should be ready with the
// exception of actually listening on a port. We won't need to listen on a port
// for integration testing since we're calling the ServeMux directly.
func SetupTest() {
	bootstrap()
	truth.SetMux(router)
}
