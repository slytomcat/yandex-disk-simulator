package main

import (
	"os"
	"os/exec"
)

func init() {
	exec.Command("go", "build", "yandex-disk-simulator.go").Run()
	cwd, _ := os.Getwd()
	path := os.Getenv("PATH")
	os.Setenv("PATH", cwd+":"+path)
}

func Example1() {
	cmd := exec.Command("go", "run", "yandex-disk-simulator.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
	// Output:
	// Error: command hasn't been specified. Use the --help command to access help
	// or setup to launch the setup wizard.
	// exit status 1
}
