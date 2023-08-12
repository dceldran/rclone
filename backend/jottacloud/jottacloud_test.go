// Test Box filesystem interface
package jottacloud_test

import (
	"testing"

	"github.com/dceldran/rclone/backend/jottacloud"
	"github.com/dceldran/rclone/fstest/fstests"
)

// TestIntegration runs integration tests against the remote
func TestIntegration(t *testing.T) {
	fstests.Run(t, &fstests.Opt{
		RemoteName: "TestJottacloud:",
		NilObject:  (*jottacloud.Object)(nil),
	})
}
