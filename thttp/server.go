package thttp

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"log"
	"net/http"
	"time"
	"torn/treporter"
)

const IsoLayout = "2006-01-02"
const DefaultEarliest = "2019-08-24"
const DefaultLatest = "2019-08-31"

type Server struct {
	Cache *cache.Cache
	Reporter *treporter.Reporter
}

func (s Server) RefreshCachePeriodically() {
	earliest, _ := time.Parse(IsoLayout, DefaultEarliest)
	latest, _ := time.Parse(IsoLayout, DefaultLatest)
	go func() {
		log.Println("Starting ET cache update")
		userEnergy, err := s.Reporter.CalculateEnergyTrained(earliest, latest)
		if err != nil {
			log.Printf("ERR: Unable to refresh cache on interval: %v", err)
			time.Sleep(time.Second * 10)
		} else {
			log.Println("Successfully updated ET cache")
			s.Cache.Set("t", userEnergy, time.Minute * 1)
			time.Sleep(time.Second * 30)
		}
	}()
}

func (s Server) Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Page requested: %s\n", r.Header.Get("X-Forwarded-For"))
	sEarliest := "2019-08-24"
	sLatest := "2019-08-31"
	earliest, eerr := time.Parse(IsoLayout, sEarliest)
	latest, lerr := time.Parse(IsoLayout, sLatest)
	if eerr != nil || lerr != nil {
		log.Printf("WARN: Invalid report date range: earliest=%s, latest=%s, err1=%s, err2=%s\n",
			earliest, latest, eerr, lerr)
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Add("Content-Type", "plain/text")
		_, err := w.Write([]byte("Unable to parse date range, check [earliest] and [latest] query parameters."))
		if err != nil {
			log.Printf("ERR: Unable to write bad date range to response: %v", err)
		}
		return
	}
	var userEnergy []treporter.UserEnergy
	cached, _ := s.Cache.Get("t")
	if cached != nil {
		userEnergy = cached.([]treporter.UserEnergy)
	} else {
		var err error
		userEnergy, err = s.Reporter.CalculateEnergyTrained(earliest, latest)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "plain/text")
			_, err := w.Write([]byte("Oops, something went horribly wrong. Please ping Epi :D"))
			if err != nil {
				log.Printf("ERR: Unable to write 500 response: %v", err)
			}
			return
		}
		s.Cache.Set("t", userEnergy, time.Minute * 1)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "plain/text")
	for _, ue := range userEnergy {
		_, err := w.Write([]byte(fmt.Sprintf("User [%d] has trained [%d] energy at the gym.\n", ue.User, ue.Energy)))
		if err != nil {
			log.Printf("ERR: Unable to write UserEnergy to response: %v", err)
		}
	}
}