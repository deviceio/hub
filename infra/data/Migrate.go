package data

import (
	"log"

	"github.com/deviceio/shared/try"
	"github.com/deviceio/shared/types"

	r "gopkg.in/gorethink/gorethink.v2"
)

// Migrate conducts any required data migrations
func Migrate() {
	try.Call(func() error {
		c, err := r.DBList().Run(Session)

		if err != nil {
			return err
		}

		var dblist types.StringSlice

		c.All(&dblist)

		log.Println("Available Databases", dblist)

		if !dblist.Contains(Database) {
			log.Println("Creating Database", Database)
			r.DBCreate(Database).RunWrite(Session)
		}

		return nil
	}, func(e error, stack string) {
		log.Fatal(e, stack)
	})

	try.Call(func() error {
		tables := &types.StringSlice{
			"Device",
			"Hub",
			"User",
			"ApiKey",
			"Config",
		}

		c, err := r.TableList().Run(Session)

		if err != nil {
			return err
		}

		var tablelist types.StringSlice

		c.All(&tablelist)

		log.Println("Available Tables", tablelist)

		for _, table := range tables.ToSlice() {
			if !tablelist.Contains(table) {
				log.Println("Creating Table", table)
				r.TableCreate(table).RunWrite(Session)
			}
		}

		return nil
	}, func(e error, stack string) {
		log.Fatal(e, stack)
	})
}
