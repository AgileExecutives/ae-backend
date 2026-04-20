//go:build ignore
// +build ignore

package main

import "log"

// This file is an executable scratch/example and is intentionally excluded from
// normal builds and tests.
func main() {
	log.Println("examples/settings_with_invoice_integration.go is excluded via //go:build ignore")
}
