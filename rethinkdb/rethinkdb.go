package rethinkdb

import (
	"errors"
	"fmt"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
	"log"
	"time"
	"torn/model"
)

type RethinkTornUser struct {
	Id        int64      `r:"id"`
	Offset    int64      `r:"offset"`
	Timestamp time.Time  `r:"timestamp,omitempty"`
	Document  model.User `r:"document,omitempty"`
}

type UserDao struct {
	Session *r.Session
}

func (dao UserDao) GetUserIds() ([]int64, error) {
	m := make(map[string]string)
	m["document"] = "userId"
	cursor, err := r.DB("TornEnergy").Table("User").
		Distinct(r.DistinctOpts{Index: "userId"}).
		Run(dao.Session)
	if err != nil {
		return nil, err
	}
	var rows []int64
	if err = cursor.All(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (dao UserDao) GetInRange(id int64, earliest time.Time, latest time.Time) ([]RethinkTornUser, error) {
	cursor, err := r.DB("TornEnergy").Table("User").
		Between([]interface{}{id, earliest}, []interface{}{id, latest}, r.BetweenOpts{LeftBound: "closed", RightBound: "open", Index: "userIdTimestamp"}).
		OrderBy(r.OrderByOpts{Index: "userIdTimestamp"}).
		Run(dao.Session)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	var rows []RethinkTornUser
	err = cursor.All(&rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (dao UserDao) Exists(id int64) (bool, error) {
	// TODO: Replace with channel
	cursor, err := r.DB("TornEnergy").Table("User").Get(id).
		Field("id").
		Default(nil).
		Run(dao.Session)
	if err != nil {
		return false, err
	}
	var row interface{}
	err = cursor.One(&row)
	return err != r.ErrEmptyResult, nil
}

func (dao UserDao) Insert(user RethinkTornUser) error {
	response, err := r.DB("TornEnergy").Table("User").
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

func SetUpDb(server string) *r.Session {
	r.SetTags("r", "json")
	session, err := r.Connect(r.ConnectOpts{
		Address: server,
		ReadTimeout: time.Second * 6,
		WriteTimeout: time.Second * 30,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return session
}