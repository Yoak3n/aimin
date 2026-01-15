package controller

import (
	"github.com/Yoak3n/aimin/blood/dao/implements"
)

var db *implements.Database

func init() {
	db, _ = implements.NewDatabase(implements.DefaultDatabaseConfig())
}

func GetDB() *implements.Database {
	return db
}
