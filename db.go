package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

const DB_CONNECT_STRING = "host=198.199.109.47 port=5432 user=postgres password=zel123foot dbname=gk_prod sslmode=disable"

func db_connect() *gorm.DB {
	db, err := gorm.Open("postgres", DB_CONNECT_STRING)
	if err != nil {
		fmt.Printf("Database opening error -->%v\n", err)
		panic("Database error")
	}
	fmt.Printf("Connected to DB  \r\n")
	return &db
}

func init_database(pdb **gorm.DB) {
	err := pdb.CreateTable(&Location{})
	if err != nil {
		fmt.Printf("Create table error -->%v\n", err)
		panic("Create table error")
	}
}
