package main

import "testing"

// Flags work after the spec too, where people tend to type them.
func TestFlagsAfterTheSpec(t *testing.T) {
	spec := "testdata/neighbr/neighbr.philote"
	before := runOK(t, "tree", "--depth", "1", spec)
	after := runOK(t, "tree", spec, "--depth", "1")
	if before != after {
		t.Errorf("tree with --depth after the spec printed:\n%s\nwant:\n%s", after, before)
	}
	if runOK(t, "expand", spec, "api/horses") == "" {
		t.Error("expand with two positional arguments printed nothing")
	}
}
