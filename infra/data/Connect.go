package data

import (
	"log"

	r "gopkg.in/gorethink/gorethink.v2"
)

// Connect establishes our connection to rethinkdb
func Connect(opts *Options) {
	s, err := r.Connect(r.ConnectOpts{
		Address:  opts.DBHost,
		Database: opts.DBName,
	})

	if err != nil {
		log.Fatalln(err)
	}

	Session = s

	Database = opts.DBName
}
