// Test Pcloud filesystem interface
package pcloud_test

import (
	"testing"

	"github.com/dceldran/rclone/backend/pcloud"
	"github.com/dceldran/rclone/fstest/fstests"
)

// TestIntegration runs integration tests against the remote
func TestIntegration(t *testing.T) {
	fstests.Run(t, &fstests.Opt{
		RemoteName: "TestPcloud:",
		NilObject:  (*pcloud.Object)(nil),
	})
}
