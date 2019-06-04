package main

import (
	"flag"
	"net/http"
	"regexp"
	"strings"

	docker_client "docker.io/go-docker"
	docker_types "docker.io/go-docker/api/types"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
	"k8s.io/node-problem-detector/pkg/systemlogmonitor/logwatchers/kmsg"
	"k8s.io/node-problem-detector/pkg/systemlogmonitor/logwatchers/types"
)

var (
	kmesgRE = regexp.MustCompile("/pod(\\w+\\-\\w+\\-\\w+\\-\\w+\\-\\w+)/([a-f0-9]+) killed as a result of limit of /kubepods")
)

var (
	kubernetesCounterVec      *prometheus.CounterVec
	prometheusContainerLabels = map[string]string{
		"io.kubernetes.container.name": "container_name",
		"io.kubernetes.pod.namespace":  "namespace",
		"io.kubernetes.pod.uid":        "pod_uid",
		"io.kubernetes.pod.name":       "pod_name",
	}
	metricsAddr  string
	dockerClient *docker_client.Client
)

func init() {
	var err error
	flag.StringVar(&metricsAddr, "listen-address", ":9102", "The address to listen on for HTTP requests.")
	dockerClient, err = docker_client.NewEnvClient()
	if err != nil {
		glog.Fatal(err)
	}
	dockerClient.NegotiateAPIVersion(context.Background())
}

func main() {
	flag.Parse()

	var labels []string
	for _, label := range prometheusContainerLabels {
		labels = append(labels, strings.Replace(label, ".", "_", -1))
	}
	kubernetesCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "klog_pod_oomkill",
		Help: "Extract metrics for OOMKilled pods from kernel log",
	}, labels)

	prometheus.MustRegister(kubernetesCounterVec)

	go func() {
		glog.Info("Starting prometheus metrics")
		http.Handle("/metrics", promhttp.Handler())
		glog.Warning(http.ListenAndServe(metricsAddr, nil))
	}()

	kmsgWatcher := kmsg.NewKmsgWatcher(types.WatcherConfig{Plugin: "kmsg"})
	logCh, err := kmsgWatcher.Watch()

	if err != nil {
		glog.Fatal("Could not create log watcher")
	}

	for log := range logCh {
		podUID, containerID := getContainerIDFromLog(log.Message)
		if containerID != "" {
			container, err := getContainer(containerID, dockerClient)
			if err != nil {
				glog.Warningf("Could not get container %s for pod %s: %v", containerID, podUID, err)
			} else {
				prometheusCount(container.Config.Labels)
			}
		}
	}
}

func getContainerIDFromLog(log string) (string, string) {
	if matches := kmesgRE.FindStringSubmatch(log); matches != nil {
		return matches[1], matches[2]
	}

	return "", ""
}

func getContainer(containerID string, cli *docker_client.Client) (docker_types.ContainerJSON, error) {
	container, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return docker_types.ContainerJSON{}, err
	}
	return container, nil

}

func prometheusCount(containerLabels map[string]string) {
	var counter prometheus.Counter
	var err error

	var labels map[string]string
	labels = make(map[string]string)
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
