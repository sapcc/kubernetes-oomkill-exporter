// Copyright 2024 SAP SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/go-api-declarations/bininfo"
	"golang.org/x/net/context"
	"k8s.io/node-problem-detector/pkg/systemlogmonitor/logwatchers/kmsg"
	"k8s.io/node-problem-detector/pkg/systemlogmonitor/logwatchers/types"
)

var (
	defaultPattern            = `^oom-kill.+,task_memcg=\/kubepods(?:\.slice)?\/.+\/(?:kubepods-burstable-)?pod(\w+[-_]\w+[-_]\w+[-_]\w+[-_]\w+)(?:\.slice)?\/(?:cri-containerd-)?([a-f0-9]+)`
	kmesgRE                   = regexp.MustCompile(defaultPattern)
	kubernetesCounterVec      *prometheus.CounterVec
	prometheusContainerLabels = map[string]string{
		"io.kubernetes.container.name": "container_name",
		"io.kubernetes.pod.namespace":  "namespace",
		"io.kubernetes.pod.uid":        "pod_uid",
		"io.kubernetes.pod.name":       "pod_name",
	}
	metricsAddr string
	versionFlag bool
	newPattern  string
)

func init() {
	flag.StringVar(&metricsAddr, "listen-address", ":9102", "The address to listen on for HTTP requests.")
	flag.StringVar(&newPattern, "regexp-pattern", defaultPattern, "Overwrites the default regexp pattern to match and extract Pod UID and Container ID.")
	flag.BoolVar(&versionFlag, "version", false, "Print version info")
}

func main() {
	flag.Parse()

	if versionFlag {
		fmt.Printf("Version: %s\n", bininfo.Version())
		os.Exit(0)
	}

	if newPattern != "" {
		kmesgRE = regexp.MustCompile(newPattern)
	}

	containerdClient, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		glog.Fatal(err)
	}
	defer containerdClient.Close()

	var labels []string
	for _, label := range prometheusContainerLabels {
		labels = append(labels, strings.ReplaceAll(label, ".", "_"))
	}
	kubernetesCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "klog_pod_oomkill",
		Help: "Extract metrics for OOMKilled pods from kernel log",
	}, labels)

	prometheus.MustRegister(kubernetesCounterVec)

	go func() {
		glog.Info("Starting prometheus metrics")
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		server := &http.Server{
			Addr:              metricsAddr,
			ReadHeaderTimeout: 3 * time.Second,
			Handler:           mux,
		}
		glog.Warning(server.ListenAndServe())
	}()

	kmsgWatcher := kmsg.NewKmsgWatcher(types.WatcherConfig{Plugin: "kmsg"})
	logCh, err := kmsgWatcher.Watch()

	if err != nil {
		glog.Fatal("Could not create log watcher")
	}

	for log := range logCh {
		podUID, containerID := getContainerIDFromLog(log.Message)
		if containerID != "" {
			labels, err := getContainerLabels(containerID, containerdClient)
			if err != nil || labels == nil {
				glog.Warningf("Could not get labels for container id %s, pod %s: %v", containerID, podUID, err)
			} else {
				prometheusCount(labels)
			}
		}
	}
}

func getContainerIDFromLog(log string) (podUID, containerID string) {
	podUID = ""
	containerID = ""

	if matches := kmesgRE.FindStringSubmatch(log); matches != nil {
		podUID = matches[1]
		containerID = matches[2]
	}

	return
}

func getContainerLabels(containerID string, cli *containerd.Client) (map[string]string, error) {
	ctx := namespaces.WithNamespace(context.Background(), "k8s.io")
	container, err := cli.ContainerService().Get(ctx, containerID)
	if err != nil {
		return nil, err
	}

	return container.Labels, nil
}

func prometheusCount(containerLabels map[string]string) {
	var counter prometheus.Counter
	var err error

	labels := make(map[string]string)
	for key, label := range prometheusContainerLabels {
		labels[label] = containerLabels[key]
	}

	glog.V(5).Infof("Labels: %v\n", labels)
	counter, err = kubernetesCounterVec.GetMetricWith(labels)

	if err != nil {
		glog.Warning(err)
	} else {
		counter.Add(1)
	}
}
