package tconsumer

import (
	"encoding/json"
	"fmt"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"torn/model"
	"torn/rethinkdb"
	"torn/treporter"
)

type Args struct {
	BootstrapServer string
	RethinkdbServer string
}

const GroupIdV1 = "rethinkdb-tconsumer"
const GroupIdV3 = "rethinkdb-tconsumer-v3"

func SetUpConsumer(bootstrapServer string, groupId string) (*kafka.Consumer, func()) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServer,
		"group.id" : groupId,
		"auto.offset.reset": "earliest",
		"enable.auto.commit": "false",
	})
	if err != nil {
		log.Printf("Failed to create tconsumer: %s\n", err)
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
			log.Printf("Unable to close tconsumer: %s\n", err)
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

func ToRethinkTornUser(msg *kafka.Message) (*rethinkdb.RethinkTornUser, error) {
	var tUser model.User
	err := json.Unmarshal(msg.Value, &tUser)
	if err != nil {
		return nil, err
	}
	return &rethinkdb.RethinkTornUser{
		Id: int64(msg.TopicPartition.Offset),
		Offset: int64(msg.TopicPartition.Offset),
		Timestamp: msg.Timestamp,
		Document: tUser,
	}, nil
}

func RethinkdbStoringConsumer(consumer *kafka.Consumer, userDao rethinkdb.UserDao) {
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
			stored, err := userDao.Exists(int64(offset))
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
			err = userDao.Insert(*dbUser)
			if err != nil {
				log.Printf("ERR: Unable to insert User into db: user=%+v\n", dbUser)
				continue
			}
			log.Printf("Wrote User to db: user=%+v\n", dbUser)
			time.Sleep(time.Millisecond * 50)
		}
	}()
}

func RunConsumer(args Args, done chan bool) {
	RunConsumerV1(args, done)
}

func RunConsumerV1(args Args, done chan bool) {
	consumer, closer := SetUpConsumer(args.BootstrapServer, GroupIdV1)
	defer closer()
	session := rethinkdb.SetUpDb(args.RethinkdbServer)
	userDao := rethinkdb.UserDao{Session: session}
	RethinkdbStoringConsumer(consumer, userDao)
	<-done
	partitions, err := consumer.Commit()
	if err != nil {
		log.Printf("Unable to commit to Kafka: %s\n", err)
	} else {
		log.Printf("Committing to Kafka: %+v\n", partitions)
	}
}

type UserPair struct {
	Prev *rethinkdb.RethinkTornUser
	Curr *rethinkdb.RethinkTornUser
}

// WIP
func RunConsumerV3(args Args, done chan bool) {
	consumer, closer := SetUpConsumer(args.BootstrapServer, GroupIdV1)
	defer closer()
	//session := rethinkdb.SetUpDb(args.RethinkdbServer)
	//userDao := rethinkdb.UserDao{Session: session}

	messages := make(chan *kafka.Message, 4)

	// Produce Kafka Message
	go func() {
		for {
			//log.Println("Reading message")
			msg, err := consumer.ReadMessage(time.Second * 5)
			if err != nil {
				log.Printf("Unable to read message: %v", err)
				continue
			}
			messages <- msg
		}
	}()

	// Adapt Kafka Message to User
	users := make(chan *rethinkdb.RethinkTornUser, 4)
	go func() {
		for {
			//log.Println("Converting message to User")
			msg := <-messages
			dbUser, err := ToRethinkTornUser(msg)
			if err != nil {
				log.Printf("ERR: Unable to convert Kafka message to Rethink model: offset=%d, err=%s\n", msg.TopicPartition.Offset, err)
				continue
			}
			users <- dbUser
		}
	}()

	// Produces UserPair or sets previous User to compare with next match
	pairs := make(chan UserPair, 4)
	prevs := make(map[uint]*rethinkdb.RethinkTornUser)
	go func() {
		for {
			//log.Println("Pairing User updates")
			user := <-users
			prev, exists := prevs[user.Document.UserId]
			if exists {
				pairs <- UserPair{Prev: prev, Curr: user}
			}
			prevs[user.Document.UserId] = user
		}
	}()

	// Test it out
	var pc int64
	usertrained := make(map[uint]int)
	var mux sync.Mutex
	min := time.Date(2019, time.August, 24, 0, 0, 0, 0, time.UTC)
	go func() {
		for {
			atomic.AddInt64(&pc, 1)
			pair := <-pairs
			if pair.Prev.Timestamp.Before(min) {
				continue
			}
			if pair.Curr != nil {
				udiff := pair.Prev.Document.Diff(pair.Curr.Document)
				diff, _ := json.Marshal(udiff)
				trained := udiff.CalculateEnergyTrained()
				events := udiff.GetEvents()
				if trained > 0 || len(events) > 0 {
					log.Printf("Diff (id=%d): %s (t=%d) (e=%d)\n", pair.Prev.Document.UserId, diff, trained, len(events))
					for _, e := range events {
						log.Printf("  %s\n", e)
					}
				}
				mux.Lock()
				usertrained[pair.Prev.Document.UserId] += trained
				mux.Unlock()
			}

			if pc % 100 == 0 {
				log.Printf("Processed: %d\n", pc)
			}
		}
	}()
	<-done
	log.Printf("Processed: %d\n", pc)
	sorted := treporter.SortMapByValue(usertrained)
	for _, v := range sorted {
		fmt.Printf("%d: %d energy trained\n", v.Key, v.Value)
	}
}