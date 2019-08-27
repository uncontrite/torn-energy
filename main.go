package main

import (
	"context"
	"flag"
	"github.com/patrickmn/go-cache"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"torn/rethinkdb"
	"torn/thttp"
	"torn/treporter"
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
	Report   *treporter.Args
	Server   *ServerArgs
}

type ServerArgs struct {
	RethinkdbServer string
	Port string
}

func ParseCliArgs() Args {
	// Parse args
	var bootstrapServer string
	var rethinkDbServer string
	var consumer bool
	var reporter bool
	var server bool
	var port string
	flag.StringVar(&bootstrapServer, "bootstrap-server", "127.0.0.1", "Kafka bootstrap server")
	flag.StringVar(&rethinkDbServer, "rethinkdb-server", "127.0.0.1", "RethinkDB server")
	flag.StringVar(&port, "port", ":80", "Server port")
	flag.BoolVar(&consumer, "consumer", false, "Runs app in consumer mode")
	flag.BoolVar(&reporter, "reporter", false, "Runs app in reporter mode")
	flag.BoolVar(&server, "server", false, "Runs app in server mode")
	flag.Parse()
	if reporter {
		args := treporter.Args{RethinkdbServer: rethinkDbServer}
		return Args{Report: &args}
	} else if server {
		args := ServerArgs{
			RethinkdbServer: rethinkDbServer,
			Port:            port,
		}
		return Args{Server: &args}
	}
	return Args{}
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
	if args.Report != nil {
		log.Println("Running in reporter mode.")
		treporter.RunReport(*args.Report, intTermChan)
	} else if args.Server != nil {
		log.Println("Running in server mode.")
		cash := cache.New(time.Second * 3, time.Second * 3)
		session := rethinkdb.SetUpDb(args.Server.RethinkdbServer)
		defer session.Close()
		userDao := rethinkdb.UserDao{Session: session}
		reporter := treporter.Reporter{UserDao: &userDao}
		server := thttp.Server{Cache: cash, Reporter: &reporter}
		server.RefreshCachePeriodically()
		mux := http.NewServeMux()
		mux.HandleFunc("/", server.Handler)
		srv := &http.Server{Addr: args.Server.Port, Handler: mux}
		go func() {
			// returns ErrServerClosed on graceful close
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				// NOTE: there is a chance that next line won't have time to run,
				// as main() doesn't wait for this goroutine to stop. don't use
				// code with race conditions like these for production. see post
				// comments below on more discussion on how to handle this.
				log.Fatalf("ListenAndServe(): %s", err)
			}
		}()
		<- intTermChan
		if err := srv.Shutdown(context.TODO()); err != nil {
			panic(err) // failure/timeout shutting down the server gracefully
		}
	} else {
		log.Println("Invalid arguments provided")
	}
	log.Println("Application stopping.")
}