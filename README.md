# moxy
[![Build Status](https://travis-ci.org/martensson/moxy.svg?branch=master)](https://travis-ci.org/martensson/moxy)

moxy is a HTTP Reverse Proxy and Load Balancer that automatically configures itself for microservices deployed on [Apache Mesos](http://mesos.apache.org) and [Marathon](https://mesosphere.github.io/marathon/). It is inspired by [Vulcand](https://github.com/mailgun/vulcand) and moxy does in fact use the same proxy library written by the nice people at Mailgun.

Features:

* Reverse proxy and load balancer for your microservices running inside Mesos and Marathon
* Single binary with no other dependencies for easy deployment
* Supports TLS termination
* Statistics of req/s per application via statsd
* Event callback listener to automatically be up-to-date with Marathon
* Local file backups of Marathon states, so moxy will keep serving your apps even if Marathon goes down
* + more on the works...

## Compatibility

Tested againts Marathon 0.8.1 and Mesos 0.22.1

## Getting started

1. Easiest is to install moxy from pre-compiled packages. Check `releases` page.

2. Edit config (default on ubuntu is /etc/moxy.toml):

``` toml
# moxy listening port
port = "7000"

# marathon api
marathon = "http://localhost:8080"

# statsd settings
statsd = "localhost:8125" # optional if you want to graph req/s per app
prefix = "moxy."

# tls settings
tls = false # optional if you want moxy to terminate tls
cert = "cert.pem"
key = "key.pem"
```

3. Add the moxy url + `/moxy_callback` to your callbacks in Marathon.

4. Run moxy!

## Using Moxy

Routing is based on the HTTP Host header matching app.*
Example: app.example.com and app.example.org both route to the same tasks running that app.

Example to access your apps app1,app2,app3 running in Mesos and Marathon:

    curl -i localhost:7000/ -H 'Host: app1.example.com'
    curl -i localhost:7000/ -H 'Host: app2.example.com'
    curl -i localhost:7000/ -H 'Host: app3.example.com'

### To set custom subdomain for an application

Deploy your app to Marathon setting a custom label called `moxy_subdomain`:

    "labels": {
        "moxy_subdomain": "foobar"
    },

This will override the application name and replace it with `foobar` as the new subdomain/host-header.

### Check state of Moxy

- `/moxy_stats` for traffic statistics

- `/moxy_apps` list apps and running tasks for load balancing
