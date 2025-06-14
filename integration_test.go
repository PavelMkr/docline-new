//go:build integration_test
// +build integration_test

package main

import "fmt"

// startUI is a stub function for integration testing
func startUI() {
	// Do nothing in integration test mode
	fmt.Println("UI disabled in integration test mode")
}
