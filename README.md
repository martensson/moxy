# moxy
[![Build Status](https://travis-ci.org/martensson/moxy.svg?branch=master)](https://travis-ci.org/martensson/moxy)

moxy is a HTTP Reverse Proxy and Load Balancer that automatically configures itself for web services deployed on [Apache Mesos](http://mesos.apache.org) and [Marathon](https://mesosphere.github.io/marathon/).

Features:

* Reverse proxy and rr-loadbalancer for your apps running inside Marathon/Mesos
* Single binary for easy deployment
* Support for TLS termination
* Event callback listener to always be up-to-date with changes inside Marathon
* Local file backups of Marathon states, moxy keeps serving your apps even if Marathon goes down
* + more on the works...

## Compatibility

Tested againts Marathon 0.8.1 and Mesos 0.22.1

## Getting started

Edit moxy.toml:

``` toml
port = "7000"
marathon = "http://localhost:8080"
tls = false
cert = "cert.pem"
key = "key.pem"
```

Add the moxy url + `/moxy_callback` to your callbacks in Marathon.

Start moxy!

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
