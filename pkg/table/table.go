package table

import (
	"fmt"
	"io"

	"github.com/makocchi-git/kubectl-free/pkg/util"

	"k8s.io/kubernetes/pkg/printers"
)

// OutputTable is struct of tables for outputs
type OutputTable struct {
	Header []string
	Rows   []string
	Output io.Writer
}

// NewOutputTable is an instance of OutputTable
func NewOutputTable(o io.Writer) *OutputTable {
	return &OutputTable{
		Output: o,
	}
}

// Print shows table output
func (t *OutputTable) Print() {

	// get printer
	printer := printers.GetNewTabWriter(t.Output)

	// write header
	if len(t.Header) > 0 {
		fmt.Fprintln(printer, util.JoinTab(t.Header))
	}

	// write rows
	for _, row := range t.Rows {
		fmt.Fprintln(printer, row)
	}

	// finish
	printer.Flush()
}

// AddRow adds row to table
func (t *OutputTable) AddRow(s []string) {
	t.Rows = append(t.Rows, util.JoinTab(s))
}
