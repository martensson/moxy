/* moxy - The Marathon+Mesos http reverse proxy - code by <benjamin.martensson@nrk.no> */
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
	"github.com/thoas/stats"
)

type App struct {
	Tasks []string
	Fwd   *forward.Forwarder     `json:"-"`
	Lb    *roundrobin.RoundRobin `json:"-"`
}
type Apps map[string]App

var apps Apps

func main() {
	moxystats := stats.New()
	redirect := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/moxy_callback" {
			log.Println("callback received from Marathon")
			select {
			case callbackqueue <- true: // Add reload to our queue channel, unless it is full of course.
			default:
				w.WriteHeader(202)
				fmt.Fprintln(w, "queue is full")
				return
			}
			w.WriteHeader(202)
			fmt.Fprintln(w, "queued")
			return
		} else if req.URL.Path == "/moxy_stats" {
			stats := moxystats.Data()
			b, _ := json.MarshalIndent(stats, "", "  ")
			w.Write(b)
			return
		} else if req.URL.Path == "/moxy_apps" {
			b, _ := json.MarshalIndent(apps, "", "  ")
			w.Write(b)
			return
		}
		// let us forward this request to another server container
		app := strings.Split(req.Host, ".")[0]
		if s, ok := apps[app]; ok {
			s.Lb.ServeHTTP(w, req)
		}
		fmt.Fprintln(w, "moxy")
	})
	// In case we want to log req/resp.
	//trace, _ := trace.New(redirect, os.Stdout)
	port := "7000"
	handler := moxystats.Handler(redirect)
	s := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
	callbackworker()
	callbackqueue <- true
	log.Println("Starting moxy on :" + port)
	s.ListenAndServe()
}
