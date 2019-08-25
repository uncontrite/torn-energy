package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
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

type Args struct {
	Producer *ProducerArgs
	Consumer *ConsumerArgs
	ReportArgs *ReportArgs
}

type ConsumerArgs struct {
	BootstrapServer string
	RethinkdbServer string
}

type ProducerArgs struct {
	BootstrapServer string
	ApiKeys []string
}

type ReportArgs struct {
	RethinkdbServer string
}

func ParseCliArgs() Args {
	// Parse args
	var bootstrapServer string
	var rethinkDbServer string
	var consumer bool
	var report bool
	flag.StringVar(&bootstrapServer, "bootstrap-server", "127.0.0.1", "Kafka bootstrap server")
	flag.StringVar(&rethinkDbServer, "rethinkdb-server", "127.0.0.1", "RethinkDB server")
	flag.BoolVar(&consumer, "consumer", false, "Runs app in consumer mode")
	flag.BoolVar(&report, "report", false, "Runs app in report mode")
	flag.Parse()
	if consumer {
		consumerArgs := ConsumerArgs{BootstrapServer: bootstrapServer, RethinkdbServer: rethinkDbServer}
		return Args{Consumer: &consumerArgs}
	} else if report {
		args := ReportArgs{RethinkdbServer: rethinkDbServer}
		return Args{ReportArgs: &args}
	}
	// Producer mode
	apiKeys := flag.Args()
	producerArgs := ProducerArgs{BootstrapServer: bootstrapServer, ApiKeys: apiKeys}
	return Args{Producer: &producerArgs}
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
	if args.Producer != nil {
		log.Println("Running in producer mode.")
		RunProducer(args.Producer.BootstrapServer, args.Producer.ApiKeys, intTermChan)
	} else if args.Consumer != nil {
		log.Println("Running in consumer mode.")
		RunConsumer(*args.Consumer, intTermChan)
	} else {
		log.Println("Invalid arguments provided")
	}
	log.Println("Application stopping.")
}