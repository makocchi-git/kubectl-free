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
NAME    STATUS   CPU/req   CPU/alloc   CPU/%   MEM/req   MEM/alloc   MEM/%
node1   Ready    350m      3600m       9%      73400K    5943865K    1%
node2   Ready    553m      3600m       15%     216006K   5943865K    3%
node3   Ready    455m      2           22%     52428K    3503149K    1%
```

And list containers of pod on Kubernetes node(s).

```shell
$ kubectl free --list node1
NODE NAME   POD                                   POD IP        POD STATUS   NAMESPACE     CONTAINER            CPU/req   CPU/lim   MEM/req   MEM/lim
node1       calico-node-sml6z                     10.8.2.87     Running      kube-system   calico-node          250m      -         -         -
node1       coredns-5695fb77c8-knc6d              10.112.1.3    Running      kube-system   coredns              100m      -         73400K    178257K
node1       kube-state-metrics-5f4459f5b8-7hrwz   10.112.1.42   Running      kube-system   kube-state-metrics   103m      103m      111149K   111149K
node1       kube-state-metrics-5f4459f5b8-7hrwz   10.112.1.42   Running      kube-system   addon-resizer        100m      100m      31457K    31457K
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

This plugin shows just sum of requested(limited) resources, **not a real usage**.  
I recommend to use `kubectl free` with `kubectl top`.

## License

This software is released under the MIT License.
