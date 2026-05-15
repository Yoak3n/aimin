package controller

import (
	"github.com/Yoak3n/aimin/gut/dao/implements"
)

var db *implements.Database

func init() {
	db, _ = implements.NewDatabase()
}

func GetDB() *implements.Database {
	return db
}
