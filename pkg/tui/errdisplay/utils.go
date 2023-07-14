package errdisplay

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/DomBlack/bubble-shell/internal/config"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errbase"
)

type Frame struct {
	Path     string
	Filename string
	Line     int
	Package  string
	Function string
}

// DeepestStack returns the deepest stack for the error.
func DeepestStack(err error) []Frame {
	// First get the raw stack from the error which will be
	// based on program counters
	stack := getStackFromError(err)

	unwrapped := errbase.UnwrapOnce(err)
	for unwrapped != nil {
		if thisStack := getStackFromError(unwrapped); len(thisStack) > 0 {
			stack = thisStack
		}

		unwrapped = errbase.UnwrapOnce(unwrapped)
	}

	// Then convert the program counters to frames
	frames := make([]Frame, 0, len(stack))
	for _, pc := range stack {
		// For historical reasons if Frame is interpreted as a uintptr
		// its value represents the program counter + 1.
		fn := runtime.FuncForPC(uintptr(pc) - 1)
		if fn == nil {
			continue
		}

		name := fn.Name()
		file, line := fn.FileLine(uintptr(pc) - 1)

		pkgName := path.Dir(name)
		actualPackge, fnName, _ := strings.Cut(path.Base(name), ".")
		if fnName == "" {
			fnName = actualPackge
		} else {
			pkgName = path.Join(pkgName, actualPackge)
		}

		frames = append(frames, Frame{
			Path:     file,
			Filename: path.Base(file),
			Line:     line,
			Package:  pkgName,
			Function: fnName,
		})
	}

	return frames
}

// FilterCommonFrames filters out the common frames from the stack
// that are not useful for the user to see.
//
// Such as internal frames from the runtime, coachcrockdb/errors, bubbletea, etc.
func FilterCommonFrames(cfg *config.Config, frames []Frame) []Frame {
	if len(cfg.PackagesToFilterFromStack) == 0 {
		return frames
	}

	modulePackage, _ := modulePackageAndFilePath()

	filtered := make([]Frame, 0, len(frames))
nextFrame:
	for _, frame := range frames {
		switch {
		case frame.Package == "main":
			// always show the main package
		case frame.Package == modulePackage || strings.HasPrefix(frame.Package, modulePackage+"/"):
			// always show code within the compiled module
		default:
			for _, filter := range cfg.PackagesToFilterFromStack {
				// Filter that package exactly or any subpackage
				if frame.Package == filter || strings.HasPrefix(frame.Package, filter+"/") {
					continue nextFrame
				}
			}
		}

		// If here then the frame is not filtered
		filtered = append(filtered, frame)
	}

	return filtered
}

// getStackFromError returns the stack from the error if it is available.
func getStackFromError(err error) errbase.StackTrace {
	switch err := err.(type) {
	case errbase.StackTraceProvider:
		return err.StackTrace()
	}

	return nil
}

var (
	cachedModuleName, cachedModuleFilePath string
	cacheOnce                              sync.Once
)

// modulePackageAndFilePath returns the module name and filepath on the
// machine the binary was built on.
func modulePackageAndFilePath() (packageName, filePathToPackage string) {
	cacheOnce.Do(func() {
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			return
		}

		// Find where the entry point is on a filepath basis
		var entryPointFilePath string
		base := DeepestStack(errors.New("not an error"))
		for j := len(base) - 1; j >= 0; j-- {
			if base[j].Package == "main" && base[j].Function == "main" {
				// We found the entry point.
				entryPointFilePath = filepath.Dir(base[j].Path)
			}
		}

		// We know this is the entry point package
		entryPointPkg := bi.Path

		// Now we know the path to the main function we can work out the package within the module
		// for the main function
		entryPointPkgWithinModule := strings.TrimPrefix(entryPointPkg, bi.Main.Path)
		filePathToPackage = strings.TrimSuffix(entryPointFilePath, entryPointPkgWithinModule)

		cachedModuleName = bi.Main.Path
		cachedModuleFilePath = fmt.Sprintf("%s%c", filePathToPackage, filepath.Separator)
	})

	return cachedModuleName, cachedModuleFilePath
}

// FilePathFromFrame returns the file path from the frame.
//
// If the file is part of the current application then it will be relative
// to the go.mod file of the application.
//
// Otherwise it will be the package path.
func FilePathFromFrame(frame Frame) string {
	thisModule, filePath := modulePackageAndFilePath()

	if frame.Package == "main" || frame.Package == thisModule || strings.HasPrefix(frame.Package, thisModule+"/") {
		return strings.TrimPrefix(frame.Path, filePath)
	} else {
		return fmt.Sprintf("%s/%s", frame.Package, frame.Filename)
	}
}
