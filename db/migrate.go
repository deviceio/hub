package db

import (
	"github.com/Sirupsen/logrus"
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

		logrus.Println("Available Databases", dblist)

		if !dblist.Contains(Database) {
			logrus.Println("Creating Database", Database)
			r.DBCreate(Database).RunWrite(Session)
		}

		return nil
	}, func(e error, stack string) {
		logrus.Fatal(e, stack)
	})

	try.Call(func() error {
		tables := &types.StringSlice{
			string(DeviceTable),
			string(UserTable),
			string(MemberTable),
		}

		c, err := r.TableList().Run(Session)

		if err != nil {
			return err
		}

		var tablelist types.StringSlice

		c.All(&tablelist)

		logrus.Println("Available Tables", tablelist)

		for _, table := range tables.ToSlice() {
			if !tablelist.Contains(table) {
				logrus.Println("Creating Table", table)
				r.TableCreate(table).RunWrite(Session)
			}
		}

		return nil
	}, func(e error, stack string) {
		logrus.Fatal(e, stack)
	})
}
