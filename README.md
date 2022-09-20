# kubernetes-oomkill-exporter

[![Contributions](https://img.shields.io/badge/contributions-welcome-brightgreen.svg)](https://github.com/sapcc/kubernetes-oomkill-exporter)
[![License](https://img.shields.io/badge/license-Apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.txt)

----

`kubernetes-oomkill-exporter` is parsing kernel log for killed pods, collects information like namespace from ~~docker~~ containerd and exposes them in a Prometheus metric. It can be deployed as a `DaemonSet` to run on every node in your cluster, see [here for an example](yaml/oomkill-exporter.yaml). Exported metric is called `klog_pod_oomkill` and counts the amount of oomkills of a certain pod.

A Prometheus query for alerting could look something like this:
```
sum by(namespace, pod_name) (changes(klog_pod_oomkill[30m])) > 2
```

Note: Recent versions of `kubernetes-oomkill-exporter` (`>=0.5.0`) are only working with nodes running containerd as container runtime. If you are using docker please use a version prior than that.


## License
This project is licensed under the Apache2 License - see the [LICENSE](LICENSE) file for details
