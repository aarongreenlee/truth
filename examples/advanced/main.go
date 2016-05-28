// Package main is a basic example of how to perform integration testing in-process
// using the Truth package. This method allows test coverage to be calculated by Go.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"errors"
	"strconv"
	"crypto/sha256"
	"encoding/base64"
	"sync"
)

const (
	HeaderIntegrationKey = "vnd.example.com.truth-test"
	HeaderIntegrationValue = "one-point-two-one-gigawatts"
	HeaderConfirmationToken = "vnd.example.com.confirm-token"
)

// To perform integration tests without actually going out to the network the server
// mux needs to be accessible to the Truth package. The tests are in this package and will
// have access to this un-exported mux.
var mux *http.ServeMux

func main() {
	fmt.Printf("Starting `github.com/aarongreenlee/truth/examples/basic`\n")

	// We'll separate configuration from mounting/listening so we can test
	// without actually going over the wire.
	bootstrap()

	http.ListenAndServe(":65432", mux)
}

// Mount the database
var db *Database
func init() {
	db = &Database{
		users: make(map[string]User),
		tokens: make(map[string]string),
	}
}

// Because we separated configuration from mounting/listening we can test
// the behavior of the server without going over the wire.
func bootstrap() {

	mux = http.NewServeMux()

	mux.HandleFunc("/user", func(res http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getUser(res, r)
		case http.MethodPut:
			createUser(res, r)
		}

		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	})

	mux.HandleFunc("/user/confirm", unlockUser)
}

// User demonstrates some typical real-world behavior
type User struct {
	ID    *int    `json:"ID"`
	Name  *string `json:"name"`
	Email *string `json:"email"`
	confirmed bool
}


// Validate is a simple example of how an application may validate a User.
func (u User) Validate() error {
	fails := []string{}

	if u.Name == nil {
		fails = append(fails, "Name is required")
	} else if u.Name == "" {
		fails = append(fails, "Name must not be empty")
	}

	switch {
	case u.Email == nil:
		fails = append(fails, "E-mail is required")
	case u.Email == "":
		fails = append(fails, "E-mail must not be empty")
	case !strings.Contains(u.Email, "@"):
		fails = append(fails, "E-mail appears to be invalid")
	}

	if len(fails) == 0 {
		return nil
	}

	return fmt.Errorf("Unable to create user: There were %d errors: %s", len(fails), strings.Join(fails, "; "))
}

// Token simulates a confirmation token applications typically use
// to confirm new user accounts.
type Token struct {
	ID   int    `json:"ID"`
	Hash string `json:"hash"`
}

// Database simulates a database gateway. In real tests, we would just
// write some code that talked to our real database--whatever that is.
type Database struct {
	sync.RWMutex
	users  map[string]User
	tokens map[string]int
}


var (
	ErrNotFound = errors.New("Unable to find the results you were looking for")
	ErrBadRequest = errors.New("Unable to build request due to an invalid use of the API")
)

// createUser is a demonstration of what "Create User" service might be like in your
// application. There are quite a few things we can test here.
func createUser(res http.ResponseWriter, r *http.Request) {
	user := User{}

	if err := json.Unmarshal(ioutil.ReadAll(r.Body), &user); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		io.WriteString(res, "Unable to decode JSON")
		return
	}

	if err := user.Validate(); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		io.WriteString(res, err)
		return
	}

	_, err := db.QueryUser(nil, user.Email)
	switch err {
	case nil:
		// If the database found a user then we have a conflict.
		res.WriteHeader(http.StatusConflict)
		res.Write([]byte(fmt.Sprintf("Unable to register user. The user with the e-mail %#v already exists.", *user.Email))
	case ErrBadRequest:
		// If we could not perform a query we have a bad request.
		res.Write([]byte(err.Error()))
		res.WriteHeader(http.StatusBadRequest)
		return
	default:
		// If any other error condition happened we have an
		// internal server error
		res.WriteHeader(http.StatusInternalServerError)
		return
	case ErrNotFound:
		// If we can't find a user with this e-mail we actually
		// will treat this as a success since we don't want two
		// users with the same e-mail.
		// Save the user to our fancy database

		db.Lock()
		defer db.Unlock()

		user.ID = len(db.users)
		db.users[strconv.Itoa(user.ID)] = user
		db.users[user.Email] = user

		// Generate a token
		h := sha256.New()
		h.Write([]byte(*user.Email))
		token := base64.URLEncoding.EncodeToString(h.Sum(nil))
		// Save the token to our fancy database
		db.tokens[token] = strconv.Itoa(*user.ID)

		// If we're running under an integration test let's
		// listen for the special test header we configured
		// so we can chain work-flows.
		//
		// There is an argument that you should not change the way you write
		// software for tests. If you want to be "pure" then you can simply
		// not do this! If you look at the test you'll see the Truth package
		// has nothing to do with this technique--it's all the way you
		// write your tests.
		//
		// This technique allows us to test the workflow without having to
		// check a e-mail box for a confirmation e-mail. Of course, this is Go
		// so we could have our test actually do that. But, in this example
		// we're not sending an e-mail so.... we'll just send the token as a
		// header if we passed this test header.
		if v, ok := r.Header[HeaderIntegrationKey]; ok && v == HeaderIntegrationValue {
			res.Header()[HeaderConfirmationToken] = []string{token}
		}

		response, err := json.Marshal(user)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusCreated)
		res.Write(response)
		return
	}
}

func getUser(res http.ResponseWriter, r *http.Request) {
	if id, ok := r.URL.Query()["id"]; ok {
		db.Lock()
		db.Unlock()
		if user, ok := db.users[id]; ok {
			response, err := json.Marshal(user)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			res.WriteHeader(http.StatusOK)
			res.Write(response)
			return
		}
	}

	if len(db.users) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(db.users)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write(response)
}

// Unlock user is a example of how a application may want to confirm a user's account and
// is provided as an example of how a complex workflow can be tested.
func unlockUser(res http.ResponseWriter, r *http.Request) {
		token, ok := r.URL.Query()["token"]

		if !ok {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		db.Lock()
		defer db.Unlock()

		userID, ok := db.tokens[token]
		if !ok {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		user, ok := db.users[userID]
		if !ok {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		response, err := json.Marshal(user)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Confirm the user!
		user.confirmed = true
		db.users[userID] = user
		delete db.tokens[token]

		res.WriteHeader(http.StatusOK)
		res.Write(response)
	}
}

func (db Database) QueryUser(id *int, email *string) (*User, error) {
	var k string
	switch {
		case id == nil && email == nil:
			return nil, ErrBadRequest
	case id != nil:
		k = strconv.Itoa(*id)
	case email != nil:
		k = *email
	}

	if user, ok := db.users[k]; ok {
		return &user, nil
	}

	return nil, ErrNotFound
}
