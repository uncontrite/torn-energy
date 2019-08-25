package main

import (
	"errors"
	"fmt"
	"gopkg.in/rethinkdb/rethinkdb-go.v5"
	"log"
	"time"
)

type RethinkTornUser struct {
	Id int64 `rethinkdb:"id"`
	Offset int64 `rethinkdb:"offset"`
	Timestamp time.Time `rethinkdb:"timestamp,omitempty"`
	Document interface{} `rethinkdb:"document,omitempty"`
}

type UserDao interface {
	Exists(id int64) (bool, error)
	Insert(user RethinkTornUser) error
}

type RethinkdbUserDao struct {
	Session *rethinkdb.Session
}

func (dao RethinkdbUserDao) Exists(id int64) (bool, error) {
	// TODO: Replace with channel
	cursor, err := rethinkdb.DB("TornEnergy").Table("User").Get(id).
		Field("id").
		Default(nil).
		Run(dao.Session)
	if err != nil {
		return false, err
	}
	var row interface{}
	err = cursor.One(&row)
	return err != rethinkdb.ErrEmptyResult, nil
}

func (dao RethinkdbUserDao) Insert(user RethinkTornUser) error {
	response, err := rethinkdb.DB("TornEnergy").Table("User").
		Insert(user).
		RunWrite(dao.Session)
	if err != nil {
		return err
	}
	if response.Inserted < 1 {
		return errors.New(fmt.Sprintf("ERR: Insert failed (?): response=%+v", response))
	}
	return nil
}

func SetUpDb(server string) *rethinkdb.Session {
	rethinkdb.SetTags("rethinkdb", "json")
	session, err := rethinkdb.Connect(rethinkdb.ConnectOpts{
		Address: server,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return session
}