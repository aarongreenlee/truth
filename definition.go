package truth

type (
	// Definition defines an API endpoint.
	Definition struct {
		// Properties that drive behavior.
		Method           string
		Path             string
		MIMETypeRequest  string
		MIMETypeResponse string

		// Documentation properties.
		Package     string
		Description string
		Name        string
	}
)
