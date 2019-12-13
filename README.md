# Measure and Graph Network Latency

### Overview

Under Construction WIP - This project uses the same concept as [/nerdalert/cloud-bandwidth](https://github.com/nerdalert/cloud-bandwidth) of measuring performance metrics and writing them out to a TSDB backend and then visualizing them into Grafana. The primary difference is this measures latency to the specified targets.

TODO:
- Currently only measures ICMP RTT averages. Need to add TCP connection latency to targets hosts specified in the config file.
- Update the example Grafana template to measure latency.
- Handle naming if one is not specified
- Replace s/./-/g if the target name contains `.` (or is left as an IP or DNS name) since the TSDB prefixes use dots as well.

### QuickStart Demo

Start the TSDB and Grafana:

```sh
docker run -d\
 --name go-graphite\
 --restart=always\
 -p 80:80\
 -p 2003-2004:2003-2004\
 gographite/go-graphite
```

This maps the following ports:

Host | Container | Service
---- | --------- | -------------------------------------------------------------------------------------------------------------------
  80 |        80 | [grafana](http://docs.grafana.org/)
2003 |      2003 | [carbon receiver - plaintext](http://graphite.readthedocs.io/en/latest/feeding-carbon.html#the-plaintext-protocol)
2004 |      2004 | [carbon receiver - pickle](http://graphite.readthedocs.io/en/latest/feeding-carbon.html#the-pickle-protocol)

Verify you can reach the grafana/graphite server running by pointing your browser to the container IP. If you're running Docker for desktop on a Mac, [http://localhost](http://localhost). On Linux just point to the host IP since the port is getting mapped with `-p 80:80`. The default login is `username: admin` and `password: admin`

```sh
git clone https://github.com/nerdalert/cloud-latency.git
cd cloud-latency/
```
