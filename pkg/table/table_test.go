package table

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestNewOutputTable(t *testing.T) {
	expected := &OutputTable{
		Output: os.Stdout,
	}

	actual := NewOutputTable(os.Stdout)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected(%#v) differ (got: %#v)", expected, actual)
		return
	}

}

func TestPrint(t *testing.T) {

	buffer := &bytes.Buffer{}

	var tests = []struct {
		description string
		rows        []string
		expected    string
	}{
		{"1 row", []string{"1\t2"}, "a     b\n1     2\n"},
		{"2 rows", []string{"1\t2", "3\t4"}, "a     b\n1     2\n3     4\n"},
	}

	for _, test := range tests {

		buffer.Reset()

		t.Run(test.description, func(t *testing.T) {
			table := &OutputTable{
				Header: []string{"a", "b"},
				Rows:   test.rows,
				Output: buffer,
			}

			table.Print()

			if buffer.String() != test.expected {
				t.Errorf(
					"[%s] expected(%s) differ (got: %s)",
					test.description,
					test.expected,
					buffer.String(),
				)
				return
			}
		})
	}
}

func TestAddRow(t *testing.T) {
	table := &OutputTable{}
	table.AddRow([]string{"1", "2", "3"})
	table.AddRow([]string{"4", "5", "6"})

	rows := table.Rows

	if rows[0] != "1\t2\t3" {
		t.Errorf("expected(1\t2\t3) differ (got: %s)", rows[0])
		return
	}
	if rows[1] != "4\t5\t6" {
		t.Errorf("expected(4\t5\t6) differ (got: %s)", rows[1])
		return
	}

}
