// Command lamp3d-cli is the named binary entry point; the actual CLI
// orchestration lives in internal/cli (see also the root entry point,
// runnable with "go run .").
package main

import "lamp3d-cli/internal/cli"

func main() {
	cli.Run()
}
