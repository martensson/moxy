# moxy
moxy - HTTP Reverse Proxy / Load Balancer for Marathon+Mesos

Set MARATHONAPI env to your MARATHON API Endpoint `http://localhost:8080`

Routing is based on the HTTP Host matching application.* so app.example.com and app.domain.org both route to the same app.

`/moxystats` for statistics
`/moxyapps` lists apps it proxies
`/moxycallback` add this url to your callbacks in Marathon.
