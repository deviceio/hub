package db

import (
	r "gopkg.in/gorethink/gorethink.v2"
)

type tableName string

const (
	UserTable   tableName = tableName("User")
	DeviceTable tableName = tableName("Device")
	MemberTable tableName = tableName("Member")
)

// Table returns a rethink term to a table by name
func Table(name tableName) r.Term {
	return r.DB(Database).Table(string(name))
}
