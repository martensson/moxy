# moxy
moxy - HTTP Reverse Proxy / Load Balancer for Marathon+Mesos

## Getting started

Edit moxy.toml:

``` toml
port = "7000"
marathon = "http://localhost:8080"
tls = false
cert = "cert.pem"
key = "key.pem"
```

And run moxy!

## Using Moxy 

Routing is based on the HTTP Host header matching app.* 
Example: app.example.com and app.example.org both route to the same tasks running that app.

Example to access your apps app1,app2,app3 running in Mesos and Marathon:

    curl -i localhost:7000/ -H 'Host: app1.example.com'
    curl -i localhost:7000/ -H 'Host: app2.example.com'
    curl -i localhost:7000/ -H 'Host: app3.example.com'


- `/moxy_callback` add this url to your callbacks in Marathon.

- `/moxy_stats` for traffic statistics

- `/moxy_apps` list apps and running tasks for load balancing
