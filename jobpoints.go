package main

import (
	"encoding/json"
	"sort"
)

type RawJobPoints struct {
	Jobs map[string]*json.RawMessage
	Companies map[string]*json.RawMessage
}

type Job struct {
	Name string
	Points int
}

func (jp RawJobPoints) ToJobs() ([]Job, error) {
	var jobs []Job
	var points int
	for jobName, msg := range jp.Jobs {
		if err := json.Unmarshal([]byte(*msg), &points); err != nil {
			return nil, err
		}
		jobs = append(jobs, Job{jobName, points})
	}

	for _, msg := range jp.Companies {
		var jp struct{
			Name string `json:"name,omitempty"`;
			JobPoints int `json:"jobpoints,omitempty"`;
		}
		if err := json.Unmarshal([]byte(*msg), &jp); err != nil {
			return nil, err
		}
		jobs = append(jobs, Job{jp.Name, jp.JobPoints})
	}
	sort.SliceStable(jobs, func(i, j int) bool {
		return jobs[i].Name < jobs[j].Name
	})
	return jobs, nil
}

func Eq(a, b []Job) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func Diff(l, r []Job) []Job {
	rm := make(map[string]int)
	for _, j := range r {
		rm[j.Name] = j.Points
	}
	lm := make(map[string]int)
	for _, j := range l {
		lm[j.Name] = j.Points
	}
	diff := make(map[string]int)
	for k, _ := range rm {
		jpdiff := lm[k] - rm[k]
		if jpdiff != 0 {
			diff[k] = jpdiff
		}
	}
	var jobs []Job
	for name, jp := range diff {
		jobs = append(jobs, Job{name, jp})
	}
	sort.SliceStable(jobs, func(i, j int) bool {
		return jobs[i].Name < jobs[j].Name
	})
	return jobs
}