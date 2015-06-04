package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
	"github.com/mailgun/oxy/testutils"
)

type MarathonApps struct {
	Tasks []struct {
		AppId              string `json:"appId"`
		HealthCheckResults []struct {
			Alive               bool        `json:"alive"`
			ConsecutiveFailures int64       `json:"consecutiveFailures"`
			FirstSuccess        string      `json:"firstSuccess"`
			LastFailure         interface{} `json:"lastFailure"`
			LastSuccess         string      `json:"lastSuccess"`
			TaskId              string      `json:"taskId"`
		} `json:"healthCheckResults"`
		Host         string  `json:"host"`
		Id           string  `json:"id"`
		Ports        []int64 `json:"ports"`
		ServicePorts []int64 `json:"servicePorts"`
		StagedAt     string  `json:"stagedAt"`
		StartedAt    string  `json:"startedAt"`
		Version      string  `json:"version"`
	} `json:"tasks"`
}

// buffer of two, because we dont really need more.
var callbackqueue = make(chan bool, 2)

func callbackworker() {
	// a ticker channel to limit reloads to marathon, 1s is enough for now.
	ticker := time.NewTicker(1000 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				<-callbackqueue
				err := config()
				if err != nil {
					log.Println(err.Error())
				} else {
					log.Println("config updated from Marathon")
				}
			}
		}
	}()
}

func config() error {
	marathon := os.Getenv("MARATHONAPI")
	if marathon == "" {
		return errors.New("MARATHONAPI variable not set.")
	}
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	r, err := http.NewRequest("GET", marathon+"/v2/tasks", nil)
	r.Header.Set("Accept", "application/json")
	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	jsonapps := MarathonApps{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&jsonapps)
	if err != nil {
		return err
	}
	apps = Apps{}
	for _, v := range jsonapps.Tasks {
		if len(v.HealthCheckResults) == 1 {
			if v.HealthCheckResults[0].Alive == false {
				continue
			}
		}
		if s, ok := apps[v.AppId[1:]]; ok {
			s.Lb.UpsertServer(testutils.ParseURI("http://" + v.Host + ":" + strconv.FormatInt(v.Ports[0], 10)))
			s.Tasks = append(s.Tasks, v.Host+":"+strconv.FormatInt(v.Ports[0], 10))
			apps[v.AppId[1:]] = s
		} else {
			var s = App{}
			s.Fwd, _ = forward.New()
			s.Lb, _ = roundrobin.New(s.Fwd)
			s.Lb.UpsertServer(testutils.ParseURI("http://" + v.Host + ":" + strconv.FormatInt(v.Ports[0], 10)))
			s.Tasks = []string{v.Host + ":" + strconv.FormatInt(v.Ports[0], 10)}
			apps[v.AppId[1:]] = s
		}
	}
	return nil
}
