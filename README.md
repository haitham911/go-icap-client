<h1 align="center">go-icap-client</h1>

<p align="center">
    <a href="https://github.com/k8-proxy/go-icap-client/actions/workflows/build.yml">
        <img src="https://github.com/k8-proxy/go-icap-client/actions/workflows/build.yml/badge.svg"/>
    </a>
    <a href="https://codecov.io/gh/k8-proxy/go-icap-client">
        <img src="https://codecov.io/gh/k8-proxy/go-icap-client/branch/main/graph/badge.svg"/>
    </a>	    
    <a href="https://goreportcard.com/report/github.com/k8-proxy/go-icap-client">
      <img src="https://goreportcard.com/badge/k8-proxy/go-icap-client" alt="Go Report Card">
    </a>
	<a href="https://github.com/k8-proxy/go-icap-client/pulls">
        <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat" alt="Contributions welcome">
    </a>
    <a href="https://opensource.org/licenses/Apache-2.0">
        <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="Apache License, Version 2.0">
    </a>
</p>

# go-icap-client
The Golang equivalent of c-icap-client:

******OPTIONS******

```go
 -i icap_servername
              The hostname of the icap server,port,service
              icap://<ICAP-Server-IP>:<port>/<service>
 -f filename
              Send this file to the icap server.
 -r filename
              Save output to this file.
 -c check response and file 
   check if icap server return response and file,if true must return response and file
   The default is false

```

**Initiating a plain ICAP request**

```go

./go-icap-client -i=icap://<ICAP-Server-IP>:<port>/<service> -f<input file path>
 -r<out file path> -c<out file path>
example : go run go-icap-client.go  -i=icap://34.242.219.224:1344/gw_rebuild -f=test.pdf -r=retest.pdf -c=true
return healthy server if file return 
example : go run go-icap-client.go  -i=icap://34.242.219.224:1344/gw_rebuild -f=test.pdf -r=retest.pdf -c=false
return healthy server even file not found
```

**Initiating a Secure ICAP request**

```go

./go-icap-client -i=icaps://<ICAP-Server-IP>:<port>/<service>
example : go run go-icap-client.go  -i=icaps://34.242.219.224:1345/gw_rebuild -f=test.pdf -r=retest.pdf -c=true

```



