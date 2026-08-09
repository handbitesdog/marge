package main

import "testing"

func TestBuildArgs(t *testing.T) {
	if src, dist, ok := buildArgs(nil); !ok || src != defaultSrcDir || dist != defaultDistDir {
		t.Errorf("buildArgs(nil) = %q, %q, %v, want %q, %q, true", src, dist, ok, defaultSrcDir, defaultDistDir)
	}
	if src, dist, ok := buildArgs([]string{"s", "d"}); !ok || src != "s" || dist != "d" {
		t.Errorf(`buildArgs(["s", "d"]) = %q, %q, %v, want "s", "d", true`, src, dist, ok)
	}
	for _, args := range [][]string{{"s"}, {"s", "d", "extra"}} {
		if _, _, ok := buildArgs(args); ok {
			t.Errorf("buildArgs(%v) ok = true, want false", args)
		}
	}
}

func TestServeArgs(t *testing.T) {
	if src, dist, addr, ok := serveArgs([]string{":8080"}); !ok || src != defaultSrcDir || dist != defaultDistDir || addr != ":8080" {
		t.Errorf("serveArgs([:8080]) = %q, %q, %q, %v, want %q, %q, :8080, true", src, dist, addr, ok, defaultSrcDir, defaultDistDir)
	}
	if src, dist, addr, ok := serveArgs([]string{"s", "d", ":8080"}); !ok || src != "s" || dist != "d" || addr != ":8080" {
		t.Errorf(`serveArgs(["s", "d", ":8080"]) = %q, %q, %q, %v, want "s", "d", ":8080", true`, src, dist, addr, ok)
	}
	for _, args := range [][]string{{}, {"s", "d"}, {"s", "d", "e", "extra"}} {
		if _, _, _, ok := serveArgs(args); ok {
			t.Errorf("serveArgs(%v) ok = true, want false", args)
		}
	}
}
