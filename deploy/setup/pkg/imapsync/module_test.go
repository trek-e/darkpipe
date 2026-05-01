package imapsync

import (
	"testing"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/mailmigrate"
)

func TestNewModuleAndConfig(t *testing.T) {
	state := &mailmigrate.MigrationState{}
	mapper := &mailmigrate.FolderMapper{}

	m := New(nil, nil, state, mapper, "")
	if m == nil {
		t.Fatal("expected module, got nil")
	}

	// Contract: config methods are safe to call through the interface.
	m.SetBatchSize(25)
	m.SetProgressCallbacks(nil, nil, nil)
}
