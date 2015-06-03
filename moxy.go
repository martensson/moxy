/* moxy - The Marathon+Mesos http reverse proxy - code by <benjamin.martensson@nrk.no> */
package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
)

type App struct {
	fwd *forward.Forwarder
	lb  *roundrobin.RoundRobin
}
type Apps map[string]App

var apps Apps

func main() {
	redirect := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/.reload" {
			select {
			case callbackqueue <- true: // Add reload to our queue channel, unless it is full of course.
			default:
				w.WriteHeader(202)
				return
			}
			w.WriteHeader(200)
			return
		}
		// let us forward this request to another server
		app := strings.Split(req.Host, ".")[0]
		if s, ok := apps[app]; ok {
			s.lb.ServeHTTP(w, req)
		}
	})
	// that's it! our reverse proxy is ready!
	// In case we want to log req/resp.
	//trace, _ := trace.New(redirect, os.Stdout)
	port := "8080"
	s := &http.Server{
		Addr:    ":" + port,
		Handler: redirect,
	}
	callbackworker()
	callbackqueue <- true
	log.Println("Starting moxy on :" + port)
	s.ListenAndServe()
}
