//This file is only used to speedup the docker build
//We use this to download and compile the go module dependencies before adding our own source code.
//See Dockerfile for more details

package main

import (
	_ "flag"
	_ "net/http"
	_ "regexp"
	_ "strings"

	_ "docker.io/go-docker"
	_ "docker.io/go-docker/api/types"
	_ "github.com/golang/glog"
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
	_ "golang.org/x/net/context"
	_ "k8s.io/node-problem-detector/pkg/systemlogmonitor/logwatchers/kmsg"
	_ "k8s.io/node-problem-detector/pkg/systemlogmonitor/logwatchers/types"
)

func main() {

}
