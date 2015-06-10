# moxy
moxy - HTTP Reverse Proxy / Load Balancer for Marathon+Mesos

## Getting started

- Set MARATHONAPI env to your MARATHON API Endpoint `http://localhost:8080`
- Run moxy.
- Done!

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
