package main

import (
	"net/http"
	"testing"

	"github.com/aarongreenlee/truth"
	"github.com/stretchr/testify/assert"
)

// TestCreateUser will run our create user tests.
func TestCreateUser(t *testing.T) {
	SetupTest()

	var (
		name = "Testy Mc. TestFace"
		email = "testy.mc.t@example.com"
	)

	tests := truth.TestCases{
		SuccessfulRegistration(User{
			Name: &name,
			Email: &email,
		}),
	}

	// Print some basic output as the tests run.
	truth.TogglePrintAsTestsRun()

	// Let's run the tests. Here
	truth.RunIntegrationTests(t, createUserDef, tests, nil)
}

func SuccessfulRegistration(user User) *truth.TestCase {

	result := &User{}

	tc := truth.TestCase{
		Name: "Successful User Registration",
		Payload: user,
		Status: http.StatusCreated,
		Result: result,
	}

	tc.Integration = func(t truth.Integration) {
		assert.Equal(t.T, *user.Name, *result.Name, "Name")
		assert.Equal(t.T, *user.Email, *result.Email, "Email")
		if assert.NotNil(t.T, result.ID, "ID") {
			assert.NotZero(t.T, result.ID, "ID")
		}

		// Great! We've confirmed our user. Let's talk to the database
		// and verify we actually serialized the data the way we would expect.
		dbUser, err := db.QueryOneUser(result.ID, nil, true)
		if assert.NoError(t.T, err, "Test could not load user directly from database") {
			assert.Equal(t.T, *user.Name, *dbUser.Name, "Serialized Name")
			assert.Equal(t.T, *user.Email, *dbUser.Email, "Serialized E-mail")
		}

		// Sweet! The database serialization was what we would expect.
	}

	return &tc
}

