package util

import (
	"bytes"
	"strconv"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/makocchi-git/kubectl-free/pkg/constants"

	color "github.com/gookit/color"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake "k8s.io/client-go/kubernetes/fake"
	metricsapiv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// test node object
var testNodes = []v1.Node{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node1",
			Labels: map[string]string{"hostname": "node1"},
		},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourceEphemeralStorage: *resource.NewQuantity(10*1000*1000*1000, resource.DecimalSI),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceEphemeralStorage: *resource.NewQuantity(5*1000*1000*1000, resource.DecimalSI),
			},
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{Name: "node2"},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourceEphemeralStorage: *resource.NewQuantity(10*1000*1000*1000, resource.DecimalSI),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceEphemeralStorage: *resource.NewQuantity(5*1000*1000*1000, resource.DecimalSI),
			},
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionFalse,
				},
			},
		},
	},
}

// test pod object
var testPods = []v1.Pod{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
			Containers: []v1.Container{
				{
					Name: "container1",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(2000, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(2000, resource.DecimalSI),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(1000, resource.DecimalSI),
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: "node2",
			Containers: []v1.Container{
				{
					Name: "container2a",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(1000, resource.DecimalSI),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(1000, resource.DecimalSI),
						},
					},
				},
				{
					Name: "container2b",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(50, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(100, resource.DecimalSI),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(50, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(100, resource.DecimalSI),
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod3",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: "node3",
			Containers: []v1.Container{
				{
					Name: "container3",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(300, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(300, resource.DecimalSI),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(300, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(300, resource.DecimalSI),
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodFailed,
		},
	},
}

var testMetrics = &metricsapiv1beta1.PodMetricsList{
	Items: []metricsapiv1beta1.PodMetrics{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "default",
				Labels: map[string]string{
					"key": "value",
				},
			},
			Timestamp: metav1.Now(),
			Containers: []metricsapiv1beta1.ContainerMetrics{
				{
					Name: "container1",
					Usage: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(10, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(10, resource.DecimalSI),
					},
				},
			},
		},
	},
}

func TestGetSiUnit(t *testing.T) {
	var tests = []struct {
		description string
		b           bool
		k           bool
		m           bool
		g           bool
		expectedInt int64
		expectedStr string
	}{
		{"all false", false, false, false, false, constants.UnitKiloBytes, constants.UnitKiloBytesStr},
		{"all true", true, true, true, true, constants.UnitGigaBytes, constants.UnitGigaBytesStr},
		{"b only", true, false, false, false, constants.UnitBytes, constants.UnitBytesStr},
		{"g only", false, false, false, true, constants.UnitGigaBytes, constants.UnitGigaBytesStr},
		{"b and k", true, true, false, false, constants.UnitKiloBytes, constants.UnitKiloBytesStr},
		{"k and m", false, true, true, false, constants.UnitMegaBytes, constants.UnitMegaBytesStr},
		{"k and m and g", false, true, true, true, constants.UnitGigaBytes, constants.UnitGigaBytesStr},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualInt, actualStr := GetSiUnit(test.b, test.k, test.m, test.g)
			if actualInt != test.expectedInt || actualStr != test.expectedStr {
				t.Errorf(
					"[%s] expected(%d, %s) differ (got: %d, %s)",
					test.description,
					test.expectedInt,
					test.expectedStr,
					actualInt,
					actualStr,
				)
				return
			}
		})
	}
}

func TestGetBinUnit(t *testing.T) {
	var tests = []struct {
		description string
		b           bool
		k           bool
		m           bool
		g           bool
		expectedInt int64
		expectedStr string
	}{
		{"all false", false, false, false, false, constants.UnitKibiBytes, constants.UnitKibiBytesStr},
		{"all true", true, true, true, true, constants.UnitGibiBytes, constants.UnitGibiBytesStr},
		{"b only", true, false, false, false, constants.UnitBytes, constants.UnitBytesStr},
		{"g only", false, false, false, true, constants.UnitGibiBytes, constants.UnitGibiBytesStr},
		{"b and k", true, true, false, false, constants.UnitKibiBytes, constants.UnitKibiBytesStr},
		{"k and m", false, true, true, false, constants.UnitMibiBytes, constants.UnitMibiBytesStr},
		{"k and m and g", false, true, true, true, constants.UnitGibiBytes, constants.UnitGibiBytesStr},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualInt, actualStr := GetBinUnit(test.b, test.k, test.m, test.g)
			if actualInt != test.expectedInt || actualStr != test.expectedStr {
				t.Errorf(
					"[%s] expected(%d, %s) differ (got: %d, %s)",
					test.description,
					test.expectedInt,
					test.expectedStr,
					actualInt,
					actualStr,
				)
				return
			}
		})
	}
}

func TestJoinTab(t *testing.T) {
	var tests = []struct {
		description string
		words       []string
		expected    string
	}{
		{"1 string", []string{"foo"}, "foo"},
		{"2 strings", []string{"foo", "bar"}, "foo\tbar"},
		{"3 strings", []string{"foo", "bar", "baz"}, "foo\tbar\tbaz"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := JoinTab(test.words)
			if actual != test.expected {
				t.Errorf(
					"[%s] expected(%s) differ (got: %s)",
					test.description,
					test.expected,
					actual,
				)
				return
			}
		})
	}
}

func TestSetPercentageColor(t *testing.T) {

	var tests = []struct {
		description string
		percentage  int64
		warn        int64
		crit        int64
		expected    string
	}{
		{"5 with warn 10 crit 30", 5, 10, 30, color.Green.Sprint("5%")},
		{"15 with warn 10 crit 30", 15, 10, 30, color.Yellow.Sprint("15%")},
		{"50 with warn 10 crit 30", 50, 10, 30, color.Red.Sprint("50%")},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := strconv.FormatInt(test.percentage, 10) + "%"
			SetPercentageColor(&actual, test.percentage, test.warn, test.crit)

			if actual != test.expected {
				t.Errorf(
					"[%s] expected(%v) differ (got: %v)",
					test.description,
					actual,
					test.expected,
				)
				return
			}
		})
	}
}

func TestSetNodeStatusColor(t *testing.T) {

	var tests = []struct {
		description string
		status      string
		nocolor     bool
		expected    string
	}{
		{"green status", "Ready", false, color.Green.Sprint("Ready")},
		{"red status", "NotReady", false, color.Red.Sprint("NotReady")},
		{"green status but nocolor", "Ready", true, "Ready"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			SetNodeStatusColor(&test.status, test.nocolor)
			actual := []byte(test.status)
			expectedB := []byte(test.expected)
			if !bytes.Equal(actual, expectedB) {
				t.Errorf(
					"[%s] expected(%v) differ (got: %v)",
					test.description,
					expectedB,
					actual,
				)
				return
			}
		})
	}
}

func TestGetPodStatus(t *testing.T) {

	var tests = []struct {
		description string
		podStatus   string
		nocolor     bool
		emoji       bool
		expected    string
	}{
		{"pod running", string(v1.PodRunning), false, false, color.Green.Sprint("Running")},
		{"pod succeeded", string(v1.PodSucceeded), false, false, color.Green.Sprint("Succeeded")},
		{"pod pending", string(v1.PodPending), false, false, color.Yellow.Sprint("Pending")},
		{"pod failed", string(v1.PodFailed), false, false, color.Red.Sprint("Failed")},
		{"pod other status", "other", false, false, color.FgDefault.Render("Unknown")},
		{"pod running but nocolor", string(v1.PodRunning), true, false, "Running"},
		{"pod running and emoji", string(v1.PodRunning), false, true, color.Green.Sprint("✅")},
		{"pod succeeded and emoji", string(v1.PodSucceeded), false, true, color.Green.Sprint("⭕")},
		{"pod pending and emoji", string(v1.PodPending), false, true, color.Yellow.Sprint("🚫")},
		{"pod failed and emoji", string(v1.PodFailed), false, true, color.Red.Sprint("❌")},
		{"pod unknown and emoji", "other", false, true, color.FgDefault.Render("❓")},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := GetPodStatus(test.podStatus, test.nocolor, test.emoji)
			actualB := []byte(actual)
			expectedB := []byte(test.expected)
			if !bytes.Equal(actualB, expectedB) {
				t.Errorf(
					"[%s] expected(%v) differ (got: %v)",
					test.description,
					expectedB,
					actualB,
				)
				return
			}
		})
	}
}

func TestGetNodes(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(&testNodes[0], &testNodes[1])
	fakenode := fakeClient.CoreV1().Nodes()

	// no args and no labels
	t.Run("no args and no labels", func(t *testing.T) {
		nodes, err := GetNodes(fakenode, []string{}, "")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		l := len(nodes)
		if l != 2 {
			t.Errorf("[no args and no labels] expected(2) differ (got: %d)", l)
			return
		}
	})

	// no args with valid labels
	t.Run("no args with valid labels", func(t *testing.T) {
		nodes, err := GetNodes(fakenode, []string{}, "hostname=node1")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		l := len(nodes)
		if l != 1 {
			t.Errorf("[no args and no valid labels] expected(1) differ (got: %d)", l)
			return
		}

		if nodes[0].ObjectMeta.Name != "node1" {
			t.Errorf("[no args and no valid labels] expected(node1) differ (got: %s)", nodes[0].ObjectMeta.Name)
		}
	})

	// no args with invalid labels
	t.Run("no args with invalid labels", func(t *testing.T) {
		nodes, err := GetNodes(fakenode, []string{}, "foo=bar")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		l := len(nodes)
		if l != 0 {
			t.Errorf("[no args and no invalid labels] expected(0) differ (got: %d)", l)
			return
		}
	})

	// one arg
	t.Run("one arg", func(t *testing.T) {
		nodes, err := GetNodes(fakenode, []string{"node2"}, "")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		l := len(nodes)
		if l != 1 {
			t.Errorf("[one arg] expected(1) differ (got: %d)", l)
			return
		}

		if nodes[0].ObjectMeta.Name != "node2" {
			t.Errorf("[one arg] expected(node2) differ (got: %s)", nodes[0].ObjectMeta.Name)
		}
	})

	// one arg but invalid node
	t.Run("one arg but invalid node", func(t *testing.T) {
		_, err := GetNodes(fakenode, []string{"foobar"}, "")

		if err == nil {
			t.Errorf("unexpected error: should return err")
			return
		}

	})
}

func TestGetNodeStatus(t *testing.T) {

	var tests = []struct {
		description string
		node        v1.Node
		emoji       bool
		expected    string
	}{
		{"ready", testNodes[0], false, "Ready"},
		{"notready", testNodes[1], false, "NotReady"},
		{"ready emoji", testNodes[0], true, "😃"},
		{"notready emoji", testNodes[1], true, "😭"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual, err := GetNodeStatus(test.node, test.emoji)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if actual != test.expected {
				t.Errorf(
					"[%s] expected(%s) differ (got: %s)",
					test.description,
					test.expected,
					actual,
				)
				return
			}
		})
	}
}

func TestGetPods(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(&testPods[0], &testPods[1])
	fakepod := fakeClient.CoreV1().Pods("")

	// get pods
	t.Run("get pods", func(t *testing.T) {

		// FieldSelector on fakeclient doesn't work well
		// https://github.com/kubernetes/client-go/issues/326
		pods, err := GetPods(fakepod, "dummy")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		l := len(pods.Items)
		if l != 2 {
			t.Errorf("[correct node name] expected(2) differ (got: %d)", l)
			return
		}
	})
}

func TestGetPodResources(t *testing.T) {

	//                            cpu/ mem           cpu/ mem
	// pod1 container1  : Running [requests 1000/1000] [limits 2000/2000]
	// pod2 container2a : Running [requests  500/1000] [limits  500/1000]
	// pod2 container2b : Running [requests   50/ 100] [limits   50/ 100]
	// pod3 container3  : Failed  [requests  300/ 300] [limits  300/ 300]
	pods := v1.PodList{
		Items: []v1.Pod{
			testPods[0],
			testPods[1],
			testPods[2],
		},
	}

	// get pods resource
	t.Run("get pods resource", func(t *testing.T) {

		rc, rm, lc, lm := GetPodResources(pods)

		// 1000 + 500 + 50
		if rc != 1550 {
			t.Errorf("[get pods resource Requests.Cpu] expected(1550) differ (got: %d)", rc)
			return
		}

		// 1000 + 1000 + 100
		if rm != 2100 {
			t.Errorf("[get pods resource Requests.Memory] expected(2100) differ (got: %d)", rm)
			return
		}

		// 2000 + 500 + 50
		if lc != 2550 {
			t.Errorf("[get pods resource Limits.Cpu] expected(2550) differ (got: %d)", lc)
			return
		}

		// 2000 + 1000 + 100
		if lm != 3100 {
			t.Errorf("[get pods resource Limits.Memeory] expected(3100) differ (got: %d)", lm)
			return
		}
	})
}

func TestGetContainerMetrics(t *testing.T) {

	var tests = []struct {
		description   string
		podName       string
		containerName string
		expectedCPU   int64
		expectedMEM   int64
	}{
		{"10 and 10", "pod1", "container1", 10, 10},
		{"0 and 0", "pod999", "container999", 0, 0},
	}

	for _, test := range tests {
		t.Run("[GetContainerMetrics] cpu and mem", func(t *testing.T) {
			actualCPU, actualMEM := GetContainerMetrics(testMetrics, test.podName, test.containerName)
			if actualCPU != test.expectedCPU {
				t.Errorf("[%s cpu] expected(%d) differ (got: %d)", test.description, test.expectedCPU, actualCPU)
				return
			}
			if actualMEM != test.expectedMEM {
				t.Errorf("[%s mem] expected(%d) differ (got: %d)", test.description, test.expectedMEM, actualMEM)
				return
			}
		})
	}
}

func TestGetPodCount(t *testing.T) {

	var tests = []struct {
		description string
		pods        v1.PodList
		expected    int
	}{
		{"2 pods", v1.PodList{Items: []v1.Pod{testPods[0], testPods[1]}}, 2},
		{"1 pod ", v1.PodList{Items: []v1.Pod{testPods[0]}}, 1},
		{"0 pods", v1.PodList{Items: []v1.Pod{}}, 0},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := GetPodCount(test.pods)
			if actual != test.expected {
				t.Errorf(
					"[%s] expected(%d) differ (got: %d)",
					test.description,
					test.expected,
					actual,
				)
				return
			}
		})
	}
}

func TestGetContainerCount(t *testing.T) {

	var tests = []struct {
		description string
		pods        v1.PodList
		expected    int
	}{
		{"2 pods (1 container + 2 container)", v1.PodList{Items: []v1.Pod{testPods[0], testPods[1]}}, 3},
		{"1 pod (1 container)", v1.PodList{Items: []v1.Pod{testPods[0]}}, 1},
		{"0 pods", v1.PodList{Items: []v1.Pod{}}, 0},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := GetContainerCount(test.pods)
			if actual != test.expected {
				t.Errorf(
					"[%s] expected(%d) differ (got: %d)",
					test.description,
					test.expected,
					actual,
				)
				return
			}
		})
	}
}

func TestGetPercentage(t *testing.T) {

	var tests = []struct {
		description string
		a           int64
		b           int64
		expected    int64
	}{
		{"0/10", 0, 10, 0},
		{"10/20", 10, 20, 50},
		{"30/20", 30, 20, 150},
		{"30/0", 30, 0, 0},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := GetPercentage(test.a, test.b)
			if actual != test.expected {
				t.Errorf(
					"[%s] expected(%d) differ (got: %d)",
					test.description,
					test.expected,
					actual,
				)
				return
			}
		})
	}

}

func TestColor(t *testing.T) {

	t.Run("default color", func(t *testing.T) {
		s := "foo"
		expected := []byte(color.FgDefault.Render(s))
		DefaultColor(&s)
		actual := []byte(s)
		if !bytes.Equal(actual, expected) {
			t.Errorf(
				"[default color] expected(%v) differ (got: %v)",
				expected,
				actual,
			)
			return
		}
	})

	t.Run("green color", func(t *testing.T) {
		s := "foo"
		expected := []byte(color.FgGreen.Render(s))
		Green(&s)
		actual := []byte(s)
		if !bytes.Equal(actual, expected) {
			t.Errorf(
				"[green color] expected(%v) differ (got: %v)",
				expected,
				actual,
			)
			return
		}
	})

	t.Run("red color", func(t *testing.T) {
		s := "foo"
		expected := []byte(color.FgRed.Render(s))
		Red(&s)
		actual := []byte(s)
		if !bytes.Equal(actual, expected) {
			t.Errorf(
				"[red color] expected(%v) differ (got: %v)",
				expected,
				actual,
			)
			return
		}
	})

	t.Run("yellow color", func(t *testing.T) {
		s := "foo"
		expected := []byte(color.FgYellow.Render(s))
		Yellow(&s)
		actual := []byte(s)
		if !bytes.Equal(actual, expected) {
			t.Errorf(
				"[yellow color] expected(%v) differ (got: %v)",
				expected,
				actual,
			)
			return
		}
	})
}
