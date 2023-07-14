package errdisplay

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DomBlack/bubble-shell/internal/config"
	"github.com/cockroachdb/errors"
)

// Note these functions are first _before_ tests so that the line numbers match expectations
func one() error {
	return errors.Wrap(two(), "one")
}

func two() error {
	return errors.Wrap(three(), "two")
}

func three() error {
	return errors.New("three")
}

func TestDeepestStack(t *testing.T) {
	cfg := config.Default()
	cfg.PackagesToFilterFromStack = []string{"runtime", "testing"}
	stack := FilterCommonFrames(
		cfg,
		DeepestStack(one()),
	)

	expected := []Frame{
		{Filename: "utils_test.go", Line: 22, Package: "github.com/DomBlack/bubble-shell/pkg/tui/errdisplay", Function: "three"},
		{Filename: "utils_test.go", Line: 18, Package: "github.com/DomBlack/bubble-shell/pkg/tui/errdisplay", Function: "two"},
		{Filename: "utils_test.go", Line: 14, Package: "github.com/DomBlack/bubble-shell/pkg/tui/errdisplay", Function: "one"},
		{Filename: "utils_test.go", Line: 30, Package: "github.com/DomBlack/bubble-shell/pkg/tui/errdisplay", Function: "TestDeepestStack"},
	}

	if len(stack) != len(expected) {
		var got strings.Builder
		for _, frame := range stack {
			got.WriteString("\t")
			got.WriteString(fmt.Sprintf("%+v", frame))
			got.WriteString("\n")
		}
		t.Errorf("Expected %d frames but got %d\n\n%s", len(expected), len(stack), got.String())

		if len(stack) < len(expected) {
			return
		}
	}

	for i, frame := range expected {
		if frame.Package != stack[i].Package {
			t.Errorf("Expected package %s but got %s", frame.Package, stack[i].Package)
		}

		if frame.Function != stack[i].Function {
			t.Errorf("Expected function %s but got %s", frame.Function, stack[i].Function)
		}

		if frame.Filename != stack[i].Filename {
			t.Errorf("Expected filename %s but got %s", frame.Filename, stack[i].Filename)
		}

		if frame.Line != stack[i].Line {
			t.Errorf("Expected line %d but got %d", frame.Line, stack[i].Line)

		}
	}

}
