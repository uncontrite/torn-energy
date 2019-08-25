package main

import (
	"encoding/json"
	"flag"
	gcache "github.com/patrickmn/go-cache"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

/*
Features:
	* Register API key
	* Unregister API key
	* Retrieve User
	* Determine If User Updated
	* Publish User to Kafka
	* Alert: Torn API Down
	* Alert: Unable to Publish to Kafka
	* Alert: Haven't Published to Kafka in 30 Seconds
*/

type TrackerUser struct {
	TornApiKey string `json:"apiKey,omitempty"`
	Frequency time.Duration `json:"frequency,omitempty"`
}

func UpdateUser(tornClient *TornClient, cache *gcache.Cache, producer *kafka.Producer, TornApiKey string) (userId *uint, err error) {
	// Get User
	user, tornError, err := tornClient.GetUser(TornApiKey)
	if err != nil {
		return nil, err
	} else if tornError != nil {
		return nil, tornError.GetError()
	}
	userKey := strconv.FormatUint(uint64(user.UserId), 10)

	cachedUser, _ := cache.Get(userKey)
	cache.Set(userKey, *user, gcache.NoExpiration)
	if !user.Equals(cachedUser) {
		log.Printf("User updated:\n  Old:%+v\n  New:%+v\n\n", cachedUser, *user)
		// Produce messages to topic (asynchronously)
		userJson, err := json.Marshal(user)
		if err != nil {
			return &user.UserId, err
		}
		topic := "TornEnergy"
		err = producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key: 			[]byte(userKey),
			Value:          userJson,
		}, nil)
		if err != nil {
			return &user.UserId, err
		}
	}
	return &user.UserId, nil
}

func RunProducer(bootstrapServer string, apiKeys []string, done chan bool) {
	// Global setup
	var trackerUsers []TrackerUser
	for _, apiKey := range apiKeys {
		trackerUsers = append(trackerUsers, TrackerUser{apiKey, time.Second * 5})
	}
	var tornClient = NewTornClient()
	var cache = gcache.New(gcache.NoExpiration, gcache.NoExpiration)

	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": bootstrapServer})
	if err != nil {
		log.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
				break
			default:
				log.Printf("Ignored event: %s\n", ev)
				break
			}
		}
	}()

	// Set up repeat poller per TrackerUser
	tickersByUser := make(map[TrackerUser]*time.Ticker)

	for _, tu := range trackerUsers {
		ticker := time.NewTicker(tu.Frequency)
		tickersByUser[tu] = ticker
	}

	for _, tu := range trackerUsers {
		go func(tu TrackerUser) {
			ticker := tickersByUser[tu]
			for t := range ticker.C {
				truncatedApiKey := tu.TornApiKey[:4]
				log.Printf("Job started: %s at %s\n", truncatedApiKey, t)
				userId, err := UpdateUser(tornClient, cache, producer, tu.TornApiKey)
				if err != nil {
					if errExt, ok := err.(*TornErrorExt); ok {
						if errExt.Remove {
							log.Printf("Job failed permanently: error=%s, key=%s\n", errExt.Text, truncatedApiKey)
							ticker.Stop()
							tickersByUser[tu] = nil
							return
						} else if errExt.Delay {
							log.Printf("Job failed, delay requested: error=%s, key=%s\n", errExt.Text, truncatedApiKey)
							time.Sleep(tu.Frequency * 2)
						}
					} else {
						log.Printf("Job failed: Unable to fetch user: %s\n", err)
					}
				} else {
					log.Printf("Job succeeded: user=%d\n", *userId)
				}
			}
		}(tu)
	}

	// Replace this with chan latch
	go func() {
		for {
			activeTickers := 0
			for _, t := range tickersByUser {
				if t != nil {
					activeTickers += 1
				}
			}
			if activeTickers == 0 {
				log.Println("Signaling exit as all jobs have been stopped.")
				done <- true
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	<-done
	log.Println("Flushing Kafka producer before returning...")
	unflushedEvents := producer.Flush(15000)
	log.Printf("Flushed events: remaining=%d\n", unflushedEvents)
}

type Args struct {
	BootstrapServer string
	ApiKeys []string
}

func ParseCliArgs() Args {
	// Parse args
	bootstrapServer := flag.String("bootstrap-server", "127.0.0.1", "Kafka bootstrap server")
	flag.Parse()
	apiKeys := flag.Args()
	args := Args{BootstrapServer: *bootstrapServer, ApiKeys: apiKeys}
	return args
}
func CreateIntTermChannel() chan bool {
	// Signal handling
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Println()
		log.Println("Received termination signal: ", sig)
		done <- true
	}()
	return done
}

func main() {
	args := ParseCliArgs()
	intTermChan := CreateIntTermChannel()

	log.Println("Application initialised; awaiting termination signal.")
	RunProducer(args.BootstrapServer, args.ApiKeys, intTermChan)
	log.Println("Application stopping.")
}