package truth

import (
	"errors"
	"fmt"
)

const (
	// POST HTTP POST Request Method
	POST = "POST"
	// GET HTTP GET Request Method
	GET = "GET"
	// HEAD HTTP HEAD Request Method
	HEAD = "HEAD"
	// PUT HTTP PUT Request Method
	PUT = "PUT"
	// PATCH HTTP PATCH Request Method
	PATCH = "PATCH"
	// DELETE HTTP DELETE Request Method
	DELETE = "DELETE"
	// CONNECT HTTP CONNECT Request Method
	CONNECT = "CONNECT"
	// OPTIONS HTTP OPTIONS Request Method
	OPTIONS = "OPTIONS"
	// TRACE HTTP TRACE Request Method
	TRACE = "TRACE"

	// AuthorizationCredentials marks a resource as requiring credentials
	// be supplied.
	AuthorizationCredentials = "credentials"
	// AuthenticationChecksum marks a resource as requiring a checksum
	// be supplied.
	AuthenticationChecksum = "checksum"
	// AuthorizationOpenID marks a resource as requiring OpenID credentials.
	AuthorizationOpenID = "openID"
	// AuthorizationNone marks a resource as not requiring credentials.
	AuthorizationNone = "insecure"
)

type (
	// Definition defines an API endpoint.
	Definition struct {
		// Properties that drive behavior.
		Method           string
		Path             string
		MIMETypeRequest  string
		MIMETypeResponse string

		RequestHeaders  map[string]string
		ResponseHeaders map[string]string
		// InputParams URL Path variables /users/{ID}
		InputParams interface{}
		// QueryParams Query string parameters
		QueryParams  interface{}
		RequestBody  interface{}
		ResponseBody interface{}

		Authenticated  bool
		Authentication string

		// Attributes for documentation and logging.

		Package     string // Unique name or path to the package. Useful for generated documentation.
		Description string // Description of the endpoint to be used in generated documentation.
		Name        string // Name of the endpoint to be used in generated documentation.
		StatsKey    string // Key for instrumentation metrics.

		initialized bool
	}
)

// Init bootstraps a new Definition unless the given Definition is
// already marked as initialized.
func (def *Definition) Init() error {
	if def.initialized {
		return nil
	}

	if def.Path == "" {
		return errors.New("Definition Path is empty")
	}

	switch def.Method {
	case POST, PUT, GET, PATCH, DELETE, OPTIONS, TRACE, CONNECT, HEAD:
	// Everything is ok!
	default:
		return fmt.Errorf("Definition's HTTP method %#v is an unknown HTTP method", def.Method)
	}

	def.initialized = true

	return nil
}

// Configure returns a new Metadata struct initialized to default values unless
// customized by passing optional functions.
func Configure(d Definition, options ...func(*Definition)) Definition {
	for _, f := range options {
		f(&d)
	}

	d.Init()

	return d
}

// UsingNoAuth specifies that the provided definition does not require
// authentication to be accessed.
func UsingNoAuth(d *Definition) {
	d.Authentication = AuthorizationNone
}

// UsingCredentials specifies that the provided definition requires
// authentication to be accessed.
func UsingCredentials(d *Definition) {
	d.Authentication = AuthorizationCredentials
}

// ResourceMIMEType builds a custom mimetype such as
//	application/vnd.{your-namespace}.user
// using the provided class. The formula is:
// 	application/vnd.{your-namespace}.{class}.
//
// If an endpoint is working with messages and not domain specific resources
// use the `MessageMimeType` which focuses only on the encoding.
func ResourceMIMEType(class string) string {
	return fmt.Sprintf("application/vnd.TBD.%s", class)
}

// MessageMimeType returns a traditional mimetype to communicate the encoding
// of a message. Example:
//  application/json
//
// Preference is to pass one of the known mimetype constants. Example:
//  json
//
// If the provided key is unknown it will simply be returned prefixed as
// follows:
//  application/{key}
func MessageMimeType(key string) string {
	switch key {
	case "json":
		return MIMETypeJSON
	case "xml":
		return MIMETypeXML
	case "gob":
		return MIMETypeGOB
	}

	return "application/" + key
}
