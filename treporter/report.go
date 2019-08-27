package treporter

import (
	"fmt"
	"log"
	"sort"
	"time"
	"torn/model"
	"torn/rethinkdb"
)

type Args struct {
	RethinkdbServer string
}

type UserEnergy struct {
	User uint
	Energy int
}

type Reporter struct {
	UserDao *rethinkdb.UserDao
}

func (r Reporter) CalculateEnergyTrained(earliest time.Time, latest time.Time) ([]UserEnergy, error) {
	energyTrainedPerUser := make(map[uint]int)
	start := time.Now()
	userIds, err := r.UserDao.GetUserIds()
	elapsed := time.Since(start)
	log.Printf("GetUserIds took: %s\n", elapsed)
	if err != nil {
		return nil, err
	}
	log.Printf("Found %d distinct User IDs: %v energyTrained", len(userIds), userIds)
	for _, userId := range userIds {
		log.Printf("Consuming UserId: %d", userId)
		userData, err := r.UserDao.GetInRange(userId, earliest, latest)
		if err != nil {
			log.Printf("ERR: Unable to get history for User: id=%d, err=%s\n", userId, err)
		}
		for i := 0; i < len(userData)-1; i++ {
			prev := userData[i]
			next := userData[i+1]
			eTrained := model.CalculateEnergyTrained(prev.Document, next.Document)
			energyTrainedPerUser[prev.Document.UserId] += eTrained
		}
	}
	var userEnergy []UserEnergy
	for userId, energyTrained := range energyTrainedPerUser {
		userEnergy = append(userEnergy, UserEnergy{userId, energyTrained})
	}
	sort.SliceStable(userEnergy, func(i, j int) bool {
		return userEnergy[i].Energy > userEnergy[j].Energy
	})
	return userEnergy, nil
}

func RunReport(args Args, done chan bool) {
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
	session := rethinkdb.SetUpDb(args.RethinkdbServer)
	defer session.Close()
	userDao := rethinkdb.UserDao{session}

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
				eTrained := model.CalculateEnergyTrained(prev.Document, next.Document)
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