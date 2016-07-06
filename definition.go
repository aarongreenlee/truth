package truth

import (
	"fmt"
	"errors"
)

const (
	POST    = "POST"
	GET     = "GET"
	HEAD    = "HEAD"
	PUT     = "PUT"
	PATCH   = "PATCH"
	DELETE  = "DELETE"
	CONNECT = "CONNECT"
	OPTIONS = "OPTIONS"
	TRACE   = "TRACE"

	AuthorizationCredentials = "credentials"
	AuthenticationChecksum   = "checksum"
	AuthorizationOpenID      = "openID"
	AuthorizationNone        = "insecure"
)

type (
	// Definition defines an API endpoint.
	Definition struct {
		// Properties that drive behavior.
		Method           string
		Path             string
		MIMETypeRequest  string
		MIMETypeResponse string

		Authenticated  bool
		Authentication string

		// Documentation properties.
		Package     string
		Description string
		Name        string

		initialized bool
	}
)

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

	// def.statsKey = strings.ToLower(fmt.Sprintf("%s.%s", def.Package, def.Action))

	def.initialized = true

	return nil
}

// NewMetadata returns a new Metadata struct initialized to default values unless
// customized by passing optional functions.
func Configure(d Definition, options ...func(*Definition)) Definition {
	// Now, customize using any options provided
	//
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

// UsingNoAuth specifies that the provided definition does not require
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