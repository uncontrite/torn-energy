package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"os"
	"sync/atomic"
	"time"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"

)

func SetUpConsumer(bootstrapServer string) (*kafka.Consumer, func()) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServer,
		"group.id" : "rethinkdb-consumer",
		"auto.offset.reset": "earliest",
		"enable.auto.commit": "false",
	})
	if err != nil {
		log.Printf("Failed to create consumer: %s\n", err)
		os.Exit(1)
	}
	err = consumer.SubscribeTopics([]string{"TornEnergy"}, nil)
	if err != nil {
		log.Printf("Unable to subscribe to topic: %s\n", err)
		os.Exit(1)
	}
	return consumer, func() {
		err := consumer.Close()
		if err != nil {
			log.Printf("Unable to close consumer: %s\n", err)
		}
	}
}

func CountingConsumer(consumer *kafka.Consumer) {
	var c uint64
	go func() {
		for {
			msg, err := consumer.ReadMessage(time.Second * 5)
			if err == nil {
				//fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
				atomic.AddUint64(&c, 1)
				if c % 100 == 0 {
					log.Printf("Processed: %d records\n", c)
				}
			} else {
				// The client will automatically try to recover from all errors.
				fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			}
		}
	}()
}

type RethinkTornUser struct {
	Id int64 `rethinkdb:"id"`
	Offset int64 `rethinkdb:"offset"`
	Timestamp time.Time `rethinkdb:"timestamp,omitempty"`
	Document interface{} `rethinkdb:"document,omitempty"`
}

func ToRethinkTornUser(msg *kafka.Message) (*RethinkTornUser, error) {
	var tUser User
	err := json.Unmarshal(msg.Value, &tUser)
	if err != nil {
		return nil, err
	}
	return &RethinkTornUser{
		Id: int64(msg.TopicPartition.Offset),
		Offset: int64(msg.TopicPartition.Offset),
		Timestamp: msg.Timestamp,
		Document: tUser,
	}, nil
}

func UserExistsInDb(session *r.Session, offset kafka.Offset) (bool, error) {
	// TODO: Replace with channel
	cursor, err := r.DB("TornEnergy").Table("User").Get(offset).
		Field("id").
		Default(nil).
		Run(session)
	if err != nil {
		return false, err
	}
	var row interface{}
	err = cursor.One(&row)
	return err != r.ErrEmptyResult, nil
}

func RethinkdbStoringConsumer(consumer *kafka.Consumer, session *r.Session) {
	var kerrs uint64
	go func() {
		for {
			msg, err := consumer.ReadMessage(time.Second * 5)
			// TODO: Retry errors
			if err != nil {
				atomic.AddUint64(&kerrs, 1)
				log.Printf("Consumer error: %v (%v)\n", err, msg)
				continue
			}
			offset := msg.TopicPartition.Offset
			stored, err := UserExistsInDb(session, offset)
			if err != nil {
				log.Printf("ERR: Unable to determine if User already exists: offset=%d, err=%s\n", offset, err)
				continue
			}
			if stored {
				log.Printf("User already stored, skipping: offset=%d\n", offset)
				continue
			}
			dbUser, err := ToRethinkTornUser(msg)
			if err != nil {
				log.Printf("ERR: Unable to convert Kafka message to Rethink model: offset=%d, err=%s\n", offset, err)
				continue
			}
			response, err := r.DB("TornEnergy").Table("User").
				Insert(dbUser).
				RunWrite(session)
			if err != nil {
				log.Printf("ERR: Unable to insert User into db: offset=%d, err=%s\n", offset, err)
				continue
			} else if response.Inserted > 0 {
				log.Printf("Wrote User to db: offset=%d\n", offset)
			} else {
				log.Printf("ERR: Unable to insert User into db (?): offset=%+v\n", response)
			}
			time.Sleep(time.Millisecond * 75)
		}
	}()
}

func RunConsumer(args ConsumerArgs, done chan bool) {
	consumer, closer := SetUpConsumer(args.BootstrapServer)
	defer closer()
	r.SetTags("rethinkdb", "json")
	session, err := r.Connect(r.ConnectOpts{
		Address: args.RethinkdbServer,
	})
	if err != nil {
		log.Fatalln(err)
	}
	RethinkdbStoringConsumer(consumer, session)
	<-done
	partitions, err := consumer.Commit()
	if err != nil {
		log.Printf("Unable to commit to Kafka: %s\n", err)
	} else {
		log.Printf("Committing to Kafka: %+v\n", partitions)
	}
}