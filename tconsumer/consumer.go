package tconsumer

import (
	"encoding/json"
	"fmt"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"os"
	"sync/atomic"
	"time"
	"torn/model"
	"torn/rethinkdb"
)

type Args struct {
	BootstrapServer string
	RethinkdbServer string
}

func SetUpConsumer(bootstrapServer string) (*kafka.Consumer, func()) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServer,
		"group.id" : "rethinkdb-tconsumer-v2",
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
	consumer, closer := SetUpConsumer(args.BootstrapServer)
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