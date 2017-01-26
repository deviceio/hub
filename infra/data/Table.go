package data

import (
	r "gopkg.in/gorethink/gorethink.v2"
)

// Table returns a rethink term to a table by name
func Table(name string) r.Term {
	return r.DB(Database).Table(name)
}
