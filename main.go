package main

import (
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	docker_client "docker.io/go-docker"
	docker_types "docker.io/go-docker/api/types"
	docker_filters "docker.io/go-docker/api/types/filters"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
	"k8s.io/node-problem-detector/pkg/systemlogmonitor/logwatchers/kmsg"
	"k8s.io/node-problem-detector/pkg/systemlogmonitor/logwatchers/types"
)

const (
	OOMMatchExpression   = ".*killed as a result of limit of.*"
	PodExtractExpression = "^.+/pod(\\w+\\-\\w+\\-\\w+\\-\\w+\\-\\w+)/.+$"
	PodUIDLabel          = "io.kubernetes.pod.uid"
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
		podUID := getPodUIDFromLog(log.Message)
		if podUID != "" {
			container, err := getContainerFromPod(podUID, dockerClient)

			if err != nil {
				glog.Warningf("Could not get container for pod UID %s: %v", podUID, err)
			} else {
				prometheusCount(container)
			}
		}
	}
}

func getPodUIDFromLog(log string) string {
	match, err := regexp.MatchString(OOMMatchExpression, log)
	if err != nil {
		return ""
	}

	var ret []string
	if match {
		re := regexp.MustCompile(PodExtractExpression)
		ret = re.FindStringSubmatch(log)
		if len(ret) == 2 {
			return ret[1]
		}
	}

	return ""
}

func getContainerFromPod(podUID string, cli *docker_client.Client) (docker_types.Container, error) {
	filters := docker_filters.NewArgs()
	filters.Add("label", fmt.Sprintf("%s=%s", PodUIDLabel, podUID))
	filters.Add("label", fmt.Sprintf("%s=%s", "io.kubernetes.docker.type", "container"))

	listOpts := docker_types.ContainerListOptions{
		Filters: filters,
	}

	containers, err := cli.ContainerList(context.Background(), listOpts)
	if err != nil {
		return docker_types.Container{}, err
	}

	if len(containers) < 1 {
		return docker_types.Container{}, fmt.Errorf("There should be at least one container with UID %s", podUID)
	}

	return containers[0], nil
}

func prometheusCount(container docker_types.Container) {
	var counter prometheus.Counter
	var err error

	var labels map[string]string
	labels = make(map[string]string)
	for key, label := range prometheusContainerLabels {
		labels[label] = container.Labels[key]
	}

	glog.V(5).Infof("Labels: %v\n", labels)
	counter, err = kubernetesCounterVec.GetMetricWith(labels)

	if err != nil {
		glog.Warning(err)
	} else {
		counter.Add(1)
	}
}
