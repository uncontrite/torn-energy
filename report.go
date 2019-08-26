package main

import (
	"fmt"
	"log"
	"sort"
	"time"
)

func RunReport(args ReportArgs, done chan bool) {
	// Basic data setup
	isoLayout := "2006-01-02"
	sEarliest := "2019-08-24"
	sLatest := "2019-08-31"
	earliest, aerr := time.Parse(isoLayout, sEarliest)
	latest, lerr := time.Parse(isoLayout, sLatest)
	if aerr != nil || lerr != nil {
		log.Panicf("Invalid report date range: earliest=%s, latest=%s, err1=%s, err2=%s\n",
			earliest, latest, aerr, lerr)
	}

	// DI setup
	session := SetUpDb(args.RethinkdbServer)
	defer session.Close()
	userDao := RethinkdbUserDao{session}

	energyTrainedPerUser := make(map[uint]int)
	go func() {
		userIds, err := userDao.GetUserIds()
		if err != nil {
			log.Panicf("Unable to get User IDs: %s", err)
		}
		log.Printf("Found %d distinct User IDs: %v", len(userIds), userIds)

		for u := 0; u < len(userIds); u++ {
			userId := userIds[u]
			userData, err := userDao.GetInRange(userId, earliest, latest)
			if err != nil {
				log.Printf("Unable to get history for User: id=%d, err=%s\n", userId, err)
			}
			for i := 0; i < len(userData) - 1; i++ {
				prev := userData[i]
				next := userData[i+1]
				eTrained := CalculateEnergyTrained(prev.Document, next.Document)
				energyTrainedPerUser[prev.Document.UserId] += eTrained
				//diff := prev.Document.Diff(next.Document)
				//_, cats := diff.IsDiffRelevant()
				//if _, jp := cats["jp"]; jp && eTrained > 0 {
				//	pp, _ := json.MarshalIndent(diff, "", "  ")
				//	log.Println(string(pp))
				//}
			}
		}
		done <- true
	}()

	<- done

	type kv struct {
		Key uint
		Value int
	}
	var sorted []kv
	for k, v := range energyTrainedPerUser {
		sorted = append(sorted, kv{k, v})
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	for _, v := range sorted {
		fmt.Printf("%d: %d energy trained\n", v.Key, v.Value)
	}
}