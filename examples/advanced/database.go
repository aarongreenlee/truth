package main

import (
	"strconv"
	"sync"
	"time"
)

// Database simulates a database gateway. In real tests, we would just
// write some code that talked to our real database--whatever that is.
type Database struct {
	sync.RWMutex
	users  map[string]map[string]interface{}
	tokens map[string]int
}

// Mount the database
var db *Database

func init() {
	db = &Database{
		users:  make(map[string]map[string]interface{}, 100),
		tokens: make(map[string]int, 100),
	}
}

func (db *Database) CreateToken(token string, id int) error {
	db.Lock()
	defer db.Unlock()

	db.tokens[token] = id

	return nil
}

func (db *Database) ConsumeToken(token string) (*User, error) {
	db.Lock()
	defer db.Unlock()

	id, ok := db.tokens[token]
	if !ok {
		return nil, ErrNotFound
	}

	record, ok := db.users[strconv.Itoa(id)]
	if !ok {
		return nil, ErrNotFound
	}

	// Mark the user as confirmed
	db.users[strconv.Itoa(id)]["confirmed"] = true

	// Burn the token
	delete(db.tokens, token)

	return deserializeUser(record)
}

func (db *Database) SaveUser(user *User) error {
	db.Lock()
	defer db.Unlock()

	// We don't update users in this sample app. So, we'll create an ID every time we save.
	n := len(db.users)
	user.ID = &n

	db.users[strconv.Itoa(*user.ID)] = serializeUser(*user)

	return nil
}

func (db *Database) QueryOneUser(id *int, email *string, allowUnconfirmed bool) (*User, error) {
	db.Lock()
	defer db.Unlock()

	var k string
	switch {
	case id == nil && email == nil:
		return nil, ErrBadRequest
	case id != nil:
		k = strconv.Itoa(*id)
	case email != nil:
		k = *email
	}

	if record, ok := db.users[k]; ok {
		// Do we want to allow unconfirmed users?
		if !allowUnconfirmed {
			confirmed := record["confirmed"]
			if !confirmed.(bool) {
				return nil, ErrNotFound
			}
		}

		return deserializeUser(record)
	}

	return nil, ErrNotFound
}

func serializeUser(user User) map[string]interface{} {
	return map[string]interface{}{
		"_id":       *user.ID,
		"email":     *user.Email,
		"name":      *user.Name,
		"updated":   time.Now(),
		"confirmed": false,
	}
}

func deserializeUser(record map[string]interface{}) (*User, error) {
	id := record["_id"].(int)
	name := record["name"].(string)
	email := record["email"].(string)

	return &User{
		ID:    &id,
		Name:  &name,
		Email: &email,
	}, nil
}
