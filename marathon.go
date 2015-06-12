package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
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
				err := reload()
				if err != nil {
					log.Println(err.Error())

				} else {
					log.Println("config updated")
				}
			}
		}
	}()
}

func loadbackup(jsonapps *MarathonApps) error {
	file, err := ioutil.ReadFile(".moxy.tmp")
	if err != nil {
		log.Println("unable to open temp backup.")
		return err
	}
	err = json.Unmarshal(file, jsonapps)
	if err != nil {
		log.Println("unable to unmarshal temp backup.")
		return err
	}
	log.Println("successfully loaded backup config.")
	return nil
}

func reload() error {
	jsonapps := MarathonApps{}
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	r, err := http.NewRequest("GET", config.Marathon+"/v2/tasks", nil)
	r.Header.Set("Accept", "application/json")
	resp, err := client.Do(r)
	if err != nil {
		log.Println("Unable to contact Marathon:", err)
		err := loadbackup(&jsonapps)
		if err != nil {
			return err
		}
	} else {
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&jsonapps)
		if err != nil {
			loadbackup(&jsonapps)
			if err != nil {
				return err
			}
			log.Println("unable to parse json from Marathon, API changes?", err)
		} else {
			// We write a backup to disk, this permits us to restart moxy even if Marathon is down using last working config.
			config, err := json.MarshalIndent(jsonapps, "", "  ")
			if err != nil {
				log.Println(err)
			}
			err = ioutil.WriteFile(".moxy.tmp", config, 0644)
			if err != nil {
				log.Println("unable to write temp backup to disk:", err)
			}
		}
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
