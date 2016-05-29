// Package main is a demonstration application placed under test. This application
// allows users to be created, confirmed, and returned using an in-memory database.
//
// The objective of this application is to implement a multi-step workflow that we
// can place under test to demonstrate the Truth package.
package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func main() {
	fmt.Printf("Starting `github.com/aarongreenlee/truth/examples/advanced`\n")

	// We'll separate configuration from mounting/listening so we can test
	// without actually going over the wire.
	bootstrap()

	http.ListenAndServe(":65432", router)
}

var router Multiplexer

// Because we separated configuration from mounting/listening we can test
// the behavior of the server without going over the wire.
func bootstrap() {

	router = NewMux()

	// This sample web application was built to show a multi-step workflow
	// as a demonstration for the truth package.
	//
	// In this application we have a three step workflow to create and access
	// a new user:

	// Step 1: Create a User
	router.Handle(createUserDef, onCreateUser)

	// Step 2: Unlock/Confirm a User
	router.Handle(confirmUserDef, onConfirmUser)

	// Step 3: Get the user or users in the system.
	router.Handle(getUsersDef, onGetUser)
}

var (
	ErrNotFound   = errors.New("Unable to find the results you were looking for")
	ErrBadRequest = errors.New("Unable to build request due to an invalid use of the API")
)

// User demonstrates some typical real-world behavior
type User struct {
	ID        *int    `json:"ID,omitempty"`
	Name      *string `json:"name,omitempty"`
	Email     *string `json:"email,omitempty"`
	confirmed bool
}

// Validate is a simple example of how an application may validate a User.
func (u User) Validate() error {
	fails := []string{}

	if u.Name == nil {
		fails = append(fails, "Name is required")
	} else if *u.Name == "" {
		fails = append(fails, "Name must not be empty")
	}

	switch {
	case u.Email == nil:
		fails = append(fails, "E-mail is required")
	case *u.Email == "":
		fails = append(fails, "E-mail must not be empty")
	case !strings.Contains(*u.Email, "@"):
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






