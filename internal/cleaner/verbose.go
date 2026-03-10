package cleaner

import (
	"fmt"
	"io"
	"sync"

	"github.com/carlisle0615/OpenAgentCleaner/internal/cleaner/sessionstore"
)

var verboseState struct {
	mu      sync.Mutex
	enabled bool
	writer  io.Writer
}

func setVerboseLogger(enabled bool, writer io.Writer) {
	verboseState.mu.Lock()
	defer verboseState.mu.Unlock()
	verboseState.enabled = enabled
	verboseState.writer = writer
	sessionstore.SetVerboseLogger(enabled, verbosef)
}

func resetVerboseLogger() {
	setVerboseLogger(false, nil)
}

func verbosef(format string, args ...any) {
	verboseState.mu.Lock()
	defer verboseState.mu.Unlock()
	if !verboseState.enabled || verboseState.writer == nil {
		return
	}
	fmt.Fprintf(verboseState.writer, "[verbose] "+format+"\n", args...)
}
