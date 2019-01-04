package main

import (
	"flag"
	"os"
	"os/exec"
	"testing"
)

var (
	SyncDirPath    = "$HOME/TeSt_Yandex.Disk_TeSt"
	ConfigFilePath = "$HOME/.config/TeSt_Yandex.Disk_TeSt"
)

func TestMain(m *testing.M) {
	flag.Parse()

	exec.Command("go", "build", "yandex-disk-simulator.go").Run()
	cwd, _ := os.Getwd()
	path := os.Getenv("PATH")
	os.Setenv("PATH", cwd+":"+path)
	SyncDirPath = os.ExpandEnv(SyncDirPath)
	os.Setenv("Sim_SyncDir", SyncDirPath)
	ConfigFilePath = os.ExpandEnv(ConfigFilePath)
	os.Setenv("Sim_ConfDir", ConfigFilePath)

	// Run tests
	e := m.Run()

	// Clearance
	exec.Command("yandex-disk-simulator", "stop").Run()
	os.RemoveAll(ConfigFilePath)
	os.RemoveAll(SyncDirPath)

	os.Exit(e)
}

func Example01StartWithoutCommand() {
	cmd := exec.Command("yandex-disk-simulator")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
	// Output:
	// Error: command hasn't been specified. Use the --help command to access help
	// or setup to launch the setup wizard.
}

func Example05StartUnconfigured() {
	cmd := exec.Command("yandex-disk-simulator", "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
	// Output:
	// Error: option 'dir' is missing --
}

func Example10StartSetup() {
	cmd := exec.Command("yandex-disk-simulator", "setup")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
	// Output:
	//
}

func Example20StartSuccess() {
	cmd := exec.Command("yandex-disk-simulator", "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
	// Output:
	// Starting daemon process...Done
}

func Example25SecondStart() {
	cmd := exec.Command("yandex-disk-simulator", "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
	// Output:
	// Daemon is already running.
}

func Example90Stop() {
	cmd := exec.Command("yandex-disk-simulator", "stop")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
	// Output:
	// Daemon stopped.
}

func Example90SecondaryStop() {
	cmd := exec.Command("yandex-disk-simulator", "stop")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
	// Output:
	// Error: daemon not started
}
