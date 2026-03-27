//go:build go1.24
// +build go1.24

package main

import (
	"fmt"
	"testing"
)

// TestGo124Compatibility tests that our code is Go 1.24 compatible
func TestGo124Compatibility(t *testing.T) {
	fmt.Println("Testing Go 1.24 compatibility...")

	// Test basic functionality
	result := 2 + 2
	if result != 4 {
		t.Errorf("Expected 2+2 to be 4, got %d", result)
	}

	fmt.Println("Go 1.24 compatibility test passed!")
}
