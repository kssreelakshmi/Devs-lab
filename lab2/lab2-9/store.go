package main

import (
	"fmt"
	"strings"
)

// User represents a row in our (fake) users table.
type User struct {
	ID      int
	Name    string
	Email   string
	IsAdmin bool
}

// DB is a tiny in-memory stand-in for a real *sql.DB. It exists so this repo
// has zero external dependencies and builds offline, while still mirroring
// the shape of real database/sql code (Query takes a query string), so the
// vulnerability -- and later, the fix -- looks like what you'd see against a
// real driver.
type DB struct {
	users []User
}

func NewDB() *DB {
	return &DB{
		users: []User{
			{ID: 1, Name: "alice", Email: "alice@example.com", IsAdmin: true},
			{ID: 2, Name: "bob", Email: "bob@example.com", IsAdmin: false},
			{ID: 3, Name: "carol", Email: "carol@example.com", IsAdmin: false},
		},
	}
}

// FindUserByName -- FIXED (was CWE-89, SQL Injection)
//
// Before: the query was built with fmt.Sprintf, splicing the untrusted
// `name` directly into the string ("... WHERE name = '" + name + "'"). An
// attacker could pass a value like `x' OR '1'='1` to break out of the
// string literal and inject a tautology, which any SQL engine (real or, as
// here, a naive stand-in) would then evaluate as part of the WHERE clause --
// returning every row instead of one.
//
// After: `?` is a placeholder, and `name` travels alongside the query as a
// bound parameter, never concatenated into it. The value is only ever
// compared against as data -- it has no way to change what the query means,
// no matter what characters it contains.
func (db *DB) FindUserByName(name string) ([]User, error) {
	query := "SELECT id, name, email, is_admin FROM users WHERE name = ?"
	return db.Query(query, name)
}

// Query mirrors the (*sql.DB).Query(query string, args ...any) signature so
// this fix -- passing args separately instead of interpolating -- is a
// direct, realistic parallel to what you'd change against a real driver
// (e.g. db.QueryRow(query, arg1, arg2) with database/sql).
func (db *DB) Query(query string, args ...any) ([]User, error) {
	if !strings.Contains(query, "?") {
		return nil, fmt.Errorf("unparameterized query not supported")
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("expected exactly one bound parameter")
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("expected a string parameter")
	}

	var out []User
	for _, u := range db.users {
		if u.Name == name {
			out = append(out, u)
		}
	}
	return out, nil
}
