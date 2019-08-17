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
	Points uint
}

func (jp RawJobPoints) ToJobs() ([]Job, error) {
	var jobs []Job
	var points uint
	for jobName, msg := range jp.Jobs {
		if err := json.Unmarshal([]byte(*msg), &points); err != nil {
			return nil, err
		}
		jobs = append(jobs, Job{jobName, points})
	}

	for _, msg := range jp.Companies {
		var jp struct{
			Name string `json:"name,omitempty"`;
			JobPoints uint `json:"jobpoints,omitempty"`;
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