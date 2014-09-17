package base

import "testing"

func TestSumPath(t *testing.T) {
	expected := "/var/dedup/blobs/sha1/de/deadbeef"
	b := Base{Path: "/var/dedup", HashName: "sha1"}
	if bp := b.blobPath("deadbeef"); bp != expected {
		t.Errorf("expected %q, got %q", expected, bp)
	}
}
