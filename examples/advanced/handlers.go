package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"log"

	"github.com/aarongreenlee/truth"

)

// *********************************************************************************
// Step One: `Create User`
// *********************************************************************************

// createUserDef defines the "Create User" endpoint and is used to register the endpoint
// with the Multiplexer. It also can be used to simplify tests, generate documentation and generate
// client code.
var createUserDef = truth.Definition{
	Method: http.MethodPost,
	Path: "/users",
	MIMETypeRequest: "application/json",
	MIMETypeResponse: "application/json",
	Package: "main",
	Name: "Create User",
	Description: "Create a new user using the provided values.",
}

// createUser is a demonstration of what "Create User" service might be like in your
// application. There are quite a few things we can test here.
func onCreateUser(res http.ResponseWriter, r *http.Request, params url.Values) {
	user := User{}

	log.Println("Serving a request for", createUserDef.Name)

	input, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("[Applicaton Error] Unable to read request body!")
		res.WriteHeader(http.StatusInternalServerError)
		io.WriteString(res, "Unable to read response")
		return
	}

	if len(input) == 0 {
		log.Println("[Applicaton Error] Empty request body!")
		res.WriteHeader(http.StatusBadRequest)
		io.WriteString(res, "Request body empty!")
		return
	}

	if err := json.Unmarshal(input, &user); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		io.WriteString(res, "Unable to decode JSON")
		return
	}

	if err := user.Validate(); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		io.WriteString(res, err.Error())
		return
	}

	_, err = db.QueryOneUser(nil, user.Email, true)
	switch err {
	case nil:
		// If the database found a user then we have a conflict.
		res.WriteHeader(http.StatusConflict)
		res.Write([]byte(fmt.Sprintf("Unable to register user. The user with the e-mail %#v already exists.", *user.Email)))
	case ErrBadRequest:
		// If we could not perform a query we have a bad request.
		res.Write([]byte(err.Error()))
		res.WriteHeader(http.StatusBadRequest)
		return
	default:
		// If any other error condition happened we have an
		// internal server error
		res.WriteHeader(http.StatusInternalServerError)
		log.Println("[Applicaton Error] Unable to query user!", err.Error())
		return
	case ErrNotFound:
		// If we can't find a user with this e-mail we actually
		// will treat this as a success since we don't want two
		// users with the same e-mail.
		err = db.SaveUser(&user)
		if err != nil {
			log.Println("[Applicaton Error] Unable to save user!", err.Error())
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Generate and save a token so the user can be confirmed
		h := sha256.New()
		h.Write([]byte(*user.Email))
		token := base64.URLEncoding.EncodeToString(h.Sum(nil))
		db.CreateToken(token, *user.ID)

		// Respond
		response, err := json.Marshal(user)
		if err != nil {
			log.Println("[Applicaton Error] Unable to encode user as JSON!", err.Error())
			res.WriteHeader(http.StatusInternalServerError)
			return
		}


		res.WriteHeader(http.StatusCreated)
		res.Write(response)

		log.Printf("Created a new user! Welcome aboard %#v!", *user.Email)
		return
	}
}

// *********************************************************************************
// Step Two: `Confirm User` to unlock their account
// *********************************************************************************

var confirmUserDef = truth.Definition{
	Method:           http.MethodPost,
	Path:             "/user/confirm",
	MIMETypeRequest:  "text/plain",
	MIMETypeResponse: "text/plain",
	Package:          "main",
	Name:             "Confirm User",
	Description: `In many systems when a new User account is created an e-mail or text
	message is sent to the user with a link or code they must use to confirm and unlock
	their account. This sample Web application does not send any e-mails but it does create
	a token and insert it into the database.`,
}

// Unlock user is a example of how a application may want to confirm a user's account and
// is provided as an example of how a complex workflow can be tested.
func onConfirmUser(res http.ResponseWriter, r *http.Request, params url.Values) {
	token, ok := r.URL.Query()["token"]

	if !ok || len(token) != 1 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := db.ConsumeToken(token[0])
	switch err {
	case ErrNotFound:
		res.WriteHeader(http.StatusNotFound)
	case ErrBadRequest:
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(err.Error()))
		return
	}

	response, err := json.Marshal(user)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write(response)
}

// *********************************************************************************
// Step Three: `Get User(s)` that were created and confirmed
// *********************************************************************************

var getUsersDef = truth.Definition{
	Method:           http.MethodGet,
	Path:             "/users",
	MIMETypeRequest:  "application/json",
	MIMETypeResponse: "application/json",
	Package:          "main",
	Name:             "Create User",
	Description: `Create new users by posting to this endpoint. Once created, a
	confirmation token is generated and saved to the database. In a real application
	it would also be sent to the user's e-mail address and the application would require
	the user confirm their e-mail before unlocking the account.`,
}

func onGetUser(res http.ResponseWriter, r *http.Request, params url.Values) {
	if id, ok := r.URL.Query()["id"]; ok && len(id) == 1 {
		db.Lock()
		db.Unlock()
		if user, ok := db.users[id[0]]; ok {
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
