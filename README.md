# kubernetes-oomkill-exporter

[![Build Status](https://travis-ci.org/sapcc/kubernetes-oomkill-exporter.svg?branch=master)](https://travis-ci.org/sapcc/kubernetes-oomkill-exporter)
[![Contributions](https://img.shields.io/badge/contributions-welcome-brightgreen.svg)](https://travis-ci.org/sapcc/kubernetes-oomkill-exporter.svg?branch=master)
[![License](https://img.shields.io/badge/license-Apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.txt)

----

kubernetes-oomkill-exporter is parsing kernel log for killed pods, collects information like namespace from docker and exposes them in a metric. It can be deployed as a `DaemonSet` to run on every node in your cluster, see [here for an example](yaml/oomkill-exporter.yaml). Exported metric is called `klog_pod_oomkill` and counts the amount of oomkills of a certain pod.



## License
This project is licensed under the Apache2 License - see the [LICENSE](LICENSE) file for details
