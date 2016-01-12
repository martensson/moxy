package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
	"github.com/mailgun/oxy/testutils"
)

type MarathonTasks struct {
	Tasks []struct {
		AppId              string `json:"appId"`
		HealthCheckResults []struct {
			Alive bool `json:"alive"`
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

type MarathonApps struct {
	Apps []struct {
		Id           string            `json:"id"`
		Labels       map[string]string `json:"labels"`
		HealthChecks []interface{}     `json:"healthChecks"`
	} `json:"apps"`
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

func createBackup(jsontasks *MarathonTasks, jsonapps *MarathonApps) error {
	backup, err := json.MarshalIndent(jsontasks, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("/tmp/.moxy.tasks.tmp", backup, 0644)
	if err != nil {
		return err
	}
	backup, err = json.MarshalIndent(jsonapps, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("/tmp/.moxy.apps.tmp", backup, 0644)
	if err != nil {
		return err
	}
	return nil
}

func loadBackup(jsontasks *MarathonTasks, jsonapps *MarathonApps) error {
	file, err := ioutil.ReadFile("/tmp/.moxy.tasks.tmp")
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, jsontasks)
	if err != nil {
		return err
	}
	file, err = ioutil.ReadFile("/tmp/.moxy.apps.tmp")
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, jsonapps)
	if err != nil {
		return err
	}
	return nil
}

func fetchApps(jsontasks *MarathonTasks, jsonapps *MarathonApps) error {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	r, _ := http.NewRequest("GET", config.Marathon+"/v2/tasks", nil)
	r.Header.Set("Accept", "application/json")
	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&jsontasks)
	if err != nil {
		return err
	}
	r, _ = http.NewRequest("GET", config.Marathon+"/v2/apps", nil)
	r.Header.Set("Accept", "application/json")
	resp, err = client.Do(r)
	if err != nil {
		return err
	}
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&jsonapps)
	if err != nil {
		return err
	}
	return nil
}

func syncApps(jsontasks *MarathonTasks, jsonapps *MarathonApps) {
	apps = Apps{Apps: make(map[string]App)}
	apps.Lock()
	defer apps.Unlock()
	for _, task := range jsontasks.Tasks {
		// Use regex to remove characters that are not allowed in hostnames
		re := regexp.MustCompile("[^0-9a-z-]")
		appid := re.ReplaceAllLiteralString(task.AppId, "")
		apphealth := false
		for _, v := range jsonapps.Apps {
			if v.Id == task.AppId {
				if s, ok := v.Labels["moxy_subdomain"]; ok {
					appid = s
				}
				if len(v.HealthChecks) > 0 {
					apphealth = true
				}
			}
		}
		if apphealth {
			if len(task.HealthCheckResults) == 0 {
				// this means tasks is being deployed but not yet monitored as alive. Assume down.
				continue
			}
			alive := true
			for _, health := range task.HealthCheckResults {
				// check if health check is alive
				if health.Alive == false {
					alive = false
				}
			}
			if alive != true {
				// at least one health check has failed. Assume down.
				continue
			}
		}
		if s, ok := apps.Apps[appid]; ok {
			s.Lb.UpsertServer(testutils.ParseURI("http://" + task.Host + ":" + strconv.FormatInt(task.Ports[0], 10)))
			s.Tasks = append(s.Tasks, task.Host+":"+strconv.FormatInt(task.Ports[0], 10))
			apps.Apps[appid] = s
		} else {
			var s = App{}
			s.Fwd, _ = forward.New(forward.PassHostHeader(true))
			s.Lb, _ = roundrobin.New(s.Fwd)
			s.Lb.UpsertServer(testutils.ParseURI("http://" + task.Host + ":" + strconv.FormatInt(task.Ports[0], 10)))
			s.Tasks = []string{task.Host + ":" + strconv.FormatInt(task.Ports[0], 10)}
			apps.Apps[appid] = s
		}
	}
}

func reload() error {
	jsontasks := MarathonTasks{}
	jsonapps := MarathonApps{}
	err := fetchApps(&jsontasks, &jsonapps)
	if err != nil {
		log.Println("Unable to sync from Marathon:", err)
		err = loadBackup(&jsontasks, &jsonapps)
		if err != nil {
			return err
		}
		log.Println("successfully loaded backup config.")
	}
	syncApps(&jsontasks, &jsonapps)
	// We write a backup to disk, this permits us to restart moxy even if Marathon is down using last working config.
	err = createBackup(&jsontasks, &jsonapps)
	if err != nil {
		log.Println("Unable to create backup:", err)
	}
	return nil
}
