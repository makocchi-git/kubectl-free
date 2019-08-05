package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/makocchi-git/kubectl-free/pkg/table"

	v1 "k8s.io/api/core/v1"
	fake "k8s.io/client-go/kubernetes/fake"
	fakemetrics "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

func TestShowPodsOnNode(t *testing.T) {

	var tests = []struct {
		description   string
		listContainer bool
		listAll       bool
		nometrics     bool
		expected      []string
	}{
		{
			"list container with metrics",
			false,
			false,
			false,
			[]string{
				"node2   default   pod2   <unknown>   2.3.4.5   Running   container2a   -     500m   500m   -     1K    1K",
				"",
			},
		},
		{
			"list container: false, list all: false",
			false,
			false,
			true,
			[]string{
				"node2   default   pod2   <unknown>   2.3.4.5   Running   container2a   500m   500m   1K    1K",
				"",
			},
		},
		{
			"list container: true, list all: false",
			true,
			false,
			true,
			[]string{
				"node2   default   pod2   <unknown>   2.3.4.5   Running   container2a   500m   500m   1K    1K    nginx:latest",
				"",
			},
		},
		{
			"list container: false, list all: true",
			false,
			true,
			true,
			[]string{
				"node2   default   pod2   <unknown>   2.3.4.5   Running   container2a   500m   500m   1K    1K",
				"node2   default   pod2   <unknown>   2.3.4.5   Running   container2b   -      -      -     -",
				"",
			},
		},
		{
			"list container: true, list all: true",
			true,
			true,
			true,
			[]string{
				"node2   default   pod2   <unknown>   2.3.4.5   Running   container2a   500m   500m   1K    1K    nginx:latest",
				"node2   default   pod2   <unknown>   2.3.4.5   Running   container2b   -      -      -     -     busybox:latest",
				"",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			fakeClient := fake.NewSimpleClientset(&testPods[1])
			fakeMetricsNodeClient := fakemetrics.NewSimpleClientset(&testNodeMetrics.Items[0])
			fakeMetricsPodClient := fakemetrics.NewSimpleClientset(testPodMetrics)

			o := &FreeOptions{
				table:              table.NewOutputTable(buffer),
				noHeaders:          true,
				noMetrics:          test.nometrics,
				nocolor:            true,
				listContainerImage: test.listContainer,
				listAll:            test.listAll,
				podClient:          fakeClient.CoreV1().Pods(""),
				metricsPodClient:   fakeMetricsPodClient.MetricsV1beta1().PodMetricses("default"),
				metricsNodeClient:  fakeMetricsNodeClient.MetricsV1beta1().NodeMetricses(),
			}

			if err := o.showPodsOnNode([]v1.Node{testNodes[1]}); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			expected := strings.Join(test.expected, "\n")
			actual := buffer.String()
			if actual != expected {
				t.Errorf("[%s] expected(%s) differ (got: %s)", test.description, expected, actual)
				return
			}
		})
	}
}
