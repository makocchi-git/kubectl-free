package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/makocchi-git/kubectl-free/pkg/table"

	v1 "k8s.io/api/core/v1"
	fake "k8s.io/client-go/kubernetes/fake"
)

func TestShowFree(t *testing.T) {

	var tests = []struct {
		description string
		pod         bool
		namespace   string
		noheader    bool
		nometrics   bool
		expected    []string
		expectedErr error
	}{
		{
			"default free",
			false,
			"default",
			true,
			true,
			[]string{
				"node1   Ready   1     2     4     25%   50%   1K    2K    4K    25%   50%",
				"",
			},
			nil,
		},
		{
			"default free with metrics",
			false,
			"default",
			true,
			false,
			[]string{
				"node1   Ready   100m   1     2     4     2%    25%   50%   1K    1K    2K    4K    25%   25%   50%",
				"",
			},
			nil,
		},
		{
			"default free --pod",
			true,
			"default",
			true,
			true,
			[]string{
				"node1   Ready   1     2     4     25%   50%   1K    2K    4K    25%   50%   1     110   1",
				"",
			},
			nil,
		},
		{
			"awesome-ns free",
			true,
			"awesome-ns",
			true,
			true,
			[]string{
				"node1   Ready   200m   200m   4     5%    5%    0K    0K    4K    7%    7%    1     110   2",
				"",
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			fakeNodeClient := fake.NewSimpleClientset(&testNodes[0])
			fakePodClient := fake.NewSimpleClientset(&testPods[0], &testPods[2])
			fakeMetricsPodClient := prepareTestPodMetricsClient()
			fakeMetricsNodeClient := prepareTestNodeMetricsClient()

			buffer := &bytes.Buffer{}
			o := &FreeOptions{
				nocolor:           true,
				table:             table.NewOutputTable(buffer),
				list:              false,
				pod:               test.pod,
				noHeaders:         true,
				noMetrics:         test.nometrics,
				nodeClient:        fakeNodeClient.CoreV1().Nodes(),
				podClient:         fakePodClient.CoreV1().Pods(test.namespace),
				metricsPodClient:  fakeMetricsPodClient.MetricsV1beta1().PodMetricses("default"),
				metricsNodeClient: fakeMetricsNodeClient.MetricsV1beta1().NodeMetricses(),
			}

			if err := o.showFree([]v1.Node{testNodes[0]}); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			e := strings.Join(test.expected, "\n")
			if buffer.String() != e {
				t.Errorf("expected(%s) differ (got: %s)", e, buffer.String())
				return
			}

		})
	}

	t.Run("Allnamespace", func(t *testing.T) {

		fakeNodeClient := fake.NewSimpleClientset(&testNodes[0])
		fakePodClient := fake.NewSimpleClientset(&testPods[0], &testPods[1], &testPods[2])
		fakeMetricsPodClient := prepareTestPodMetricsClient()
		fakeMetricsNodeClient := prepareTestNodeMetricsClient()

		buffer := &bytes.Buffer{}
		o := &FreeOptions{
			nocolor:           true,
			table:             table.NewOutputTable(buffer),
			list:              false,
			pod:               false,
			allNamespaces:     true,
			noHeaders:         true,
			noMetrics:         true,
			nodeClient:        fakeNodeClient.CoreV1().Nodes(),
			podClient:         fakePodClient.CoreV1().Pods(""),
			metricsPodClient:  fakeMetricsPodClient.MetricsV1beta1().PodMetricses("default"),
			metricsNodeClient: fakeMetricsNodeClient.MetricsV1beta1().NodeMetricses(),
		}

		if err := o.showFree([]v1.Node{testNodes[0]}); err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		expected := []string{
			"node1   Ready   1700m   2700m   4     42%   67%   2K    3K    4K    57%   82%",
			"",
		}
		e := strings.Join(expected, "\n")
		if buffer.String() != e {
			t.Errorf("expected(%s) differ (got: %s)", e, buffer.String())
			return
		}

	})
}
