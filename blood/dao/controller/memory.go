package controller

import "blood/dao/implements"

var db *implements.Database

func init() {
	db, _ = implements.NewDatabase(implements.DefaultDatabaseConfig())
}
