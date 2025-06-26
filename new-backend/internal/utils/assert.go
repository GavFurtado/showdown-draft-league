package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// Assert panics if the condition is false.
func Assert(cond bool, msg string) {
	if !cond {
		_, file, line, _ := runtime.Caller(1)
		panic(fmt.Sprintf("Assertion failed at %s:%d: %s", filepath.Base(file), line, msg))
	}
}
