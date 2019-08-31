package thttp

import (
	"errors"
	"fmt"
	"github.com/patrickmn/go-cache"
	"log"
	"net/http"
	"time"
	"torn/model"
	"torn/treporter"
)

type DateRange struct {
	Begin time.Time // inclusive
	End time.Time // exclusive
}

func GetDateRangeForCompetition(in string) (*DateRange, error) {
	times := []time.Time{
		time.Date(2019, time.August, 24, 0, 0, 0, 0, time.UTC),
		time.Date(2019, time.August, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2019, time.September, 7, 0, 0, 0, 0, time.UTC),
	}
	switch in {
	case "1":
		return &DateRange{times[0], times[1]}, nil
	case "2":
		return &DateRange{times[1], times[2]}, nil
	case "overall":
		return &DateRange{times[0], times[2]}, nil
	default:
		return nil, errors.New("invalid input, expected one of: [1, 2, overall]")
	}

}

type Server struct {
	Cache *cache.Cache
	Reporter *treporter.Reporter
}

func (s Server) RefreshCachePeriodically() {
	cacheKeys := []string{"1", "2", "overall"}
	go func() {
		for {
			for _, key := range cacheKeys {
				log.Println("Starting ET cache update: key=" + key)
				dateRange, _ := GetDateRangeForCompetition(key)
				userEnergy, err := s.Reporter.CalculateEnergyTrained(dateRange.Begin, dateRange.End)
				if err != nil {
					log.Printf("ERR: Unable to refresh cache (key=%s) on interval: %v", key, err)
				} else {
					log.Println("Successfully updated ET cache: key=" + key)
					s.Cache.Set(key, userEnergy, cache.NoExpiration)
				}
			}
			time.Sleep(time.Second * 5)
		}
	}()
}

func WritePlaintextResponse(statusCode int, message string, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	w.Header().Add("Content-Type", "plain/text")
	_, err := w.Write([]byte(message))
	if err != nil {
		log.Printf("ERR: Unable to write %d response: %v", statusCode, err)
	}
}

func (s Server) Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Page requested: %v\n", r.Header)
	week := r.URL.Query().Get("week")
	if week == "" {
		week = "2"
	}
	_, err := GetDateRangeForCompetition(week)
	if err != nil {
		WritePlaintextResponse(http.StatusBadRequest, err.Error(), w)
		return
	}
	var userSummary []model.UserSummary
	cached, _ := s.Cache.Get(week)
	if cached == nil {
		WritePlaintextResponse(http.StatusInternalServerError, "Oops, something went horribly wrong. Please ping Epi :D", w)
		return
	}
	userSummary = cached.([]model.UserSummary)
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "plain/text")
	for rank, ue := range userSummary {
		_, err := fmt.Fprintf(w, "#%d [%d (%s)] %d trained [fhc=%d, xan=%d, prf=%d, lsd=%d, cans=%d, edvds=%d, jpEnergy=%d, attacks=%d]\n",
			rank+1, ue.User, ue.Name, ue.Energy, ue.FHCs, ue.Xanax, ue.EnergyRefills, ue.LSD, ue.EnergyDrinks, ue.EDVDs, ue.JpEnergy, ue.Attacks)
		if err != nil {
			log.Printf("ERR: Unable to write UserSummary to response: %v", err)
		}
	}
}