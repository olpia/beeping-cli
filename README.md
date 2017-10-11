# beeping-cli

A GO client for [beeping](https://github.com/yanc0/beeping).
For now, his only jobs is to query BeePing, and send his response to [greedee](https://github.com/yanc0/greedee) in collectd Metric format.
It's Greedee jobs to write then in the TSDB.


## Install (Linux)

1. Download last version on [release page](https://github.com/olpia/beeping-cli/releases) 
2. Set execute bit
3. Move it on your $PATH (i.e. /usr/bin or /usr/local/bin)

## Build

Beeping-cli was made on Golang 1.9 but should compile on 1.8.x

```shell
go get -u github.com/olpia/beeping-cli
cd $GOPATH/src/github.com/olpia/beeping-cli
go build
```

## Usage

```shell

$ beeping-cli
Usage:

  -beeping string
    	URL of your BeePing instance (default "http://localhost:8080")
  -check string
    	URL we want to check
  -greedee string
    	URL of your Greedee instance (default "http://localhost:9223")
  -greedeePass string
    	Greedee password if configured with basic auth
  -greedeeUser string
    	Greedee user if configured with basic auth
  -host string
    	Collectd metric's host entry, for example:'customer.app.env.servername' (default "test.test.prod.host.beeping")
  -pattern string
    	pattern that's need to be found in the body
  -timeout int
    	BeePing check timeout (default 20)
  
 $ beeping-cli -beeping http://localhost:8081 \
               -check https://www.google.fr/ \
               -host google.fr.beeping
               -greedee http://localhost:9223
```


## To Do

- [ ] Use a BeePing package
- [ ] Forward Collectd Metrics from BeePing instead of creating them
- [ ] ...

