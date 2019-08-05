# kubectl free

[![Build Status](https://travis-ci.org/makocchi-git/kubectl-free.svg?branch=master)](https://travis-ci.org/makocchi-git/kubectl-free)
[![Maintainability](https://api.codeclimate.com/v1/badges/b92591d00becc95b11ca/maintainability)](https://codeclimate.com/github/makocchi-git/kubectl-free/maintainability)
[![Go Report Card](https://goreportcard.com/badge/github.com/makocchi-git/kubectl-free)](https://goreportcard.com/report/github.com/makocchi-git/kubectl-free)
[![codecov](https://codecov.io/gh/makocchi-git/kubectl-free/branch/master/graph/badge.svg)](https://codecov.io/gh/makocchi-git/kubectl-free)
[![kubectl plugin](https://img.shields.io/badge/kubectl-plugin-blue.svg)](https://github.com/topics/kubectl-plugin)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

Print pod resources/limits usage on Kubernetes node(s) like a linux "free" command.  

```shell
$ kubectl free
NAME    STATUS   CPU/req   CPU/lim   CPU/alloc   CPU/req%   CPU/lim%   MEM/req    MEM/lim    MEM/alloc   MEM/req%   MEM/lim%
node1   Ready    704m      304m      3600m       19%        8%         807403K    375390K    5943857K    13%        6%
node2   Ready    350m      2100m     3600m       9%         58%        260046K    1304428K   5943857K    4%         21%
node3   Ready    2030m     12900m    3600m       56%        358%       3736783K   8347396K   5943865K    62%        140%
```

And list containers of pod on Kubernetes node(s).

```shell
$ kubectl free --list node1 --all-namespaces
NODE NAME  NAMESPACE     POD NAME                               POD AGE   POD IP       POD STATUS   CONTAINER            CPU/use   CPU/req   CPU/lim   MEM/use   MEM/req   MEM/lim
node1      default       nginx-7cdbd8cdc9-q2bbg                 3d22h     10.112.2.43  Running      nginx                2m        100m      2         27455K    134217K   1073741K
node1      kube-system   coredns-69dc677c56-chfcm               9d        10.112.3.2   Running      coredns              3m        100m      -         17420K    73400K    178257K
node1      kube-system   kube-flannel-ds-amd64-4b4s2            9d        10.1.2.3     Running      kube-flannel         4m        100m      100m      13877K    52428K    52428K
node1      kube-system   kube-state-metrics-69bcc79474-wvmmk    9d        10.112.3.3   Running      kube-state-metrics   11m       104m      104m      33382K    113246K   113246K
node1      kube-system   kube-state-metrics-69bcc79474-wvmmk    9d        10.112.3.3   Running      addon-resizer        1m        100m      100m      8511K     31457K    31457K
...
```

## Install

```shell
$ make
$ mv _output/kubectl-free /usr/local/bin/.

# Happy free time!
$ kubectl free
```

## Usage

```shell
# Show pod resource usage of Kubernetes nodes (default namespace is "default").
kubectl free

# Show pod resource usage of Kubernetes nodes (all namespaces).
kubectl free --all-namespaces

# Show pod resource usage of Kubernetes nodes with number of pods and containers.
kubectl free --pod

# Using label selector.
kubectl free -l key=value

# Print raw(bytes) usage.
kubectl free --bytes --without-unit

# Using binary prefix unit (GiB, MiB, etc)
kubectl free -g -B

# List resources of containers in pods on nodes.
kubectl free --list

# List resources of containers in pods on nodes with image information.
kubectl free --list --list-image

# Print container even if that has no resources/limits.
kubectl free --list --list-all

# Do you like emoji? 😃
kubectl free --emoji
kubectl free --list --emoji
```

## Notice

~~This plugin shows just sum of requested(limited) resources, **not a real usage**.  
I recommend to use `kubectl free` with `kubectl top`.~~

kubectl free v0.2.0 supports printing real usages from metrics server in a target cluster.

## License

This software is released under the MIT License.
