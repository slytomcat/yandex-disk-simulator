package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	SyncDirPath    = "$HOME/TeSt_Yandex.Disk_TeSt"         // testting synchronisation path
	ConfigFilePath = "$HOME/.config/TeSt_yandex-disk_TeSt" // testting configuration path
)

const (
	// default executable name
	exe = "yandex-disk-simulator"
)

func TestMain(m *testing.M) {
	flag.Parse()

	// build the simulator
	err := exec.Command("go", "build").Run()
	if err != nil {
		fmt.Printf("simulator building error: %v", err)
		return
	}

	// update the PATH to find executble simulator in it
	cwd, _ := os.Getwd()
	os.Setenv("PATH", cwd+":"+os.Getenv("PATH"))

	// set environment variables for setup of simulator
	SyncDirPath = os.ExpandEnv(SyncDirPath)
	os.Setenv("Sim_SyncDir", SyncDirPath)
	ConfigFilePath = os.ExpandEnv(ConfigFilePath)
	os.Setenv("Sim_ConfDir", ConfigFilePath)

	// try to stop daemon if it left running
	exec.Command(exe, "stop").Run()

	// Run tests
	errn := m.Run()

	// Clearance
	exec.Command(exe, "stop").Run()
	os.RemoveAll(ConfigFilePath)
	os.RemoveAll(SyncDirPath)

	os.Exit(errn)
}

// try to start utility without command
func TestDoMain00NoCommand(t *testing.T) {
	err := doMain(exe)
	if err == nil {
		t.Error("no error for no command")
		return
	}
	if err.Error() != "Error: command hasn't been specified. Use the --help command to access help\nor setup to launch the setup wizard." {
		t.Error("incorrect message: " + err.Error())
	}
}

// execute some command and capture stdout
func testCmdWithCapture(cmd string, t *testing.T) string {
	stdOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := doMain(exe, cmd)
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = stdOut
	if err != nil {
		t.Errorf("error while executing command '%s %s': %s", exe, cmd, err)
	}
	return string(out)
}

// try to start with 'help' command
func TestDoMain01Help(t *testing.T) {
	stdOut := os.Stdout
	args := os.Args
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Args = []string{exe, "help"}
	main()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = stdOut
	os.Args = args
	if string(out) != fmt.Sprintf(helpMsg, exe, version) {
		t.Error("incorrect message:", out)
	}
}

// try to ask for utility version
func TestDoMain01Version(t *testing.T) {
	res := testCmdWithCapture("-v", t)
	if res != fmt.Sprintf(verMsg, exe, version) {
		t.Error("incorrect message:", res)
	}
}

// try to start with wrong and long command
func TestDoMain02WrongCommand(t *testing.T) {
	err := doMain(exe, "wrongCMD_cut_it")
	if err == nil {
		t.Error("no error for wrong command")
		return
	}
	if err.Error() != "Error: unknown command: 'wrongCMD'" {
		t.Error("incorrect message: " + err.Error())
	}
}

// try to start without configuration
func TestDoMain04StartNoConfig(t *testing.T) {
	err := doMain(exe, "start")
	if err == nil {
		t.Error("no error for start without config")
		return
	}
	if err.Error() != "Error: option 'dir' is missing" {
		t.Error("incorrect message: " + err.Error())
	}
}

// try to setup the configuration
func TestDoMain05Setup(t *testing.T) {
	err := doMain(exe, "setup")
	if err != nil {
		t.Error("error for setup :", err)
	}
}

// try 'ststus' command with not started daemon
func TestDoMain07Command2NotStarted(t *testing.T) {
	err := doMain(exe, "status")
	if err == nil {
		t.Error("no error for command to not started")
		return
	}
	if err.Error() != "Error: daemon not started" {
		t.Error("incorrect message: " + err.Error())
	}
}

// try to start daemon with wrong executable name
func TestDoMain08FailWrongDaemonStart(t *testing.T) {
	stdOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := doMain("wrong-simulator", "start")
	w.Close()
	os.Stdout = stdOut
	out, _ := io.ReadAll(r)
	res := string(out)
	if err == nil {
		t.Error("No error with starting of incorrect daemon")
	}
	if res != "Starting daemon process...Fail\n" {
		t.Errorf("incorrect message: %s", res)
	}
}

// try to start configured daemon
func TestDoMain10StartSuccess(t *testing.T) {
	res := testCmdWithCapture("start", t)
	if res != "Starting daemon process...Done\n" {
		t.Error("incorrect message: " + res)
	}
}

// try to start again
func TestDoMain11StartSecondary(t *testing.T) {
	res := testCmdWithCapture("start", t)
	if res != "Daemon is already running.\n" {
		t.Error("incorrect message: " + res)
	}
}

// try to restart daemon (preparation for next test)
func TestDoMain15StartDaemon(t *testing.T) {
	// stop already executed daemon
	exec.Command(exe, "stop").Run()
	// start daemon in separate gorutine
	go doMain(exe, "daemon", SyncDirPath)
	time.Sleep(10 * time.Millisecond)
}

// try to get status of daemon right after start, it should be empty
func TestDoMain17Status(t *testing.T) {
	res := testCmdWithCapture("status", t)
	if res != "" {
		t.Error("incorrect message: " + res)
	}
}

// command execution with error handling
func execCommand(command string) {
	err := doMain(exe, command)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// try to catch the daemon log update during specified timeout
func getStatusAfterEvent(timeout time.Duration) {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watch.Close()
	err = watch.Add(filepath.Join(SyncDirPath, ".sync/cli.log"))
	if err != nil {
		return
	}
	select {
	case <-watch.Errors:
		return
	case <-watch.Events:
		execCommand("status")
		return
	case <-time.After(timeout):
		return
	}
}

// catch 1-st status change after start
func Example20StatusAfter1stEvent() {
	getStatusAfterEvent(time.Duration(2 * time.Second))
	// Output:
	// Synchronization core status: paused
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	The quota has not been received yet.
	//
	// Last synchronized items:
	// 	file: 'File.ods'
	// 	file: 'downloads/file.deb'
	// 	file: 'downloads/setup'
	// 	file: 'download'
	// 	file: 'down'
	// 	file: 'do'
	// 	file: 'd'
	// 	file: 'o'
	// 	file: 'w'
	// 	file: 'n'
}

// catch 2-nd status change after start
func Example22StatusAfter2ndEvent() {
	getStatusAfterEvent(time.Duration(2 * time.Second))
	// Output:
	// Synchronization core status: index
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	The quota has not been received yet.
}

// catch 3-rd status change after start
func Example24StatusAfter3rdEvent() {
	getStatusAfterEvent(time.Duration(2 * time.Second))
	// Output:
	// Synchronization core status: busy
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	The quota has not been received yet.
	//
	// Last synchronized items:
	// 	file: 'File.ods'
	// 	file: 'downloads/file.deb'
	// 	file: 'downloads/setup'
	// 	file: 'download'
	// 	file: 'down'
	// 	file: 'do'
	// 	file: 'd'
	// 	file: 'o'
	// 	file: 'w'
	// 	file: 'n'
}

// catch 4-th status change after start
func Example26StatusAfter4thEvent() {
	getStatusAfterEvent(time.Duration(2 * time.Second))
	// Output:
	// Synchronization core status: index
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	The quota has not been received yet.
	//
	// Last synchronized items:
	// 	file: 'File.ods'
	// 	file: 'downloads/file.deb'
	// 	file: 'downloads/setup'
	// 	file: 'download'
	// 	file: 'down'
	// 	file: 'do'
	// 	file: 'd'
	// 	file: 'o'
	// 	file: 'w'
	// 	file: 'n'
}

// catch 5-th status change after start
func Example28StatusAfter5thEvent() {
	getStatusAfterEvent(time.Duration(6 * time.Second))
	// Output:
	// Synchronization core status: idle
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	Total: 43.50 GB
	// 	Used: 2.89 GB
	// 	Available: 40.61 GB
	// 	Max file size: 50 GB
	// 	Trash size: 0 B
	//
	// Last synchronized items:
	// 	file: 'File.ods'
	// 	file: 'downloads/file.deb'
	// 	file: 'downloads/setup'
	// 	file: 'download'
	// 	file: 'down'
	// 	file: 'do'
	// 	file: 'd'
	// 	file: 'o'
	// 	file: 'w'
	// 	file: 'n'
}

// call the 'sync' command
func Example40Sync() {
	// call it
	execCommand("sync")
	// Output:
	//
}

// catch status after synchronisation start
func Example42StatusAfterSyncStart() {
	execCommand("status")
	// Output:
	// Synchronization core status: index
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	Total: 43.50 GB
	// 	Used: 2.89 GB
	// 	Available: 40.61 GB
	// 	Max file size: 50 GB
	// 	Trash size: 0 B
	//
	// Last synchronized items:
	// 	file: 'File.ods'
	// 	file: 'downloads/file.deb'
	// 	file: 'downloads/setup'
	// 	file: 'download'
	// 	file: 'down'
	// 	file: 'do'
	// 	file: 'd'
	// 	file: 'o'
	// 	file: 'w'
	// 	file: 'n'
}

// catch 2-nd status change after start of synchronisation
func Example44StatusAfter2ndEvent() {
	getStatusAfterEvent(time.Duration(2 * time.Second))
	// Output:
	// Sync progress: 0 MB/ 139.38 MB (0 %)
	// Synchronization core status: busy
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	Total: 43.50 GB
	// 	Used: 2.89 GB
	// 	Available: 40.61 GB
	// 	Max file size: 50 GB
	// 	Trash size: 0 B
	//
	// Last synchronized items:
	// 	file: 'File.ods'
	// 	file: 'downloads/file.deb'
	// 	file: 'downloads/setup'
	// 	file: 'download'
	// 	file: 'down'
	// 	file: 'do'
	// 	file: 'd'
	// 	file: 'o'
	// 	file: 'w'
	// 	file: 'n'
}

// catch 3-rd status change after start of synchronisation
func Example46StatusAfter3rdEvent() {
	getStatusAfterEvent(time.Duration(1 * time.Second))
	// Output:
	// Sync progress: 65.34 MB/ 139.38 MB (46 %)
	// Synchronization core status: busy
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	Total: 43.50 GB
	// 	Used: 2.89 GB
	// 	Available: 40.61 GB
	// 	Max file size: 50 GB
	// 	Trash size: 0 B
	//
	// Last synchronized items:
	// 	file: 'File.ods'
	// 	file: 'downloads/file.deb'
	// 	file: 'downloads/setup'
	// 	file: 'download'
	// 	file: 'down'
	// 	file: 'do'
	// 	file: 'd'
	// 	file: 'o'
	// 	file: 'w'
	// 	file: 'n'
}

// catch 4-th status change after start of synchronisation
func Example48StatusAfter4thEvent() {
	getStatusAfterEvent(time.Duration(3 * time.Second))
	// Output:
	// Sync progress: 139.38 MB/ 139.38 MB (100 %)
	// Synchronization core status: index
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	Total: 43.50 GB
	// 	Used: 2.89 GB
	// 	Available: 40.61 GB
	// 	Max file size: 50 GB
	// 	Trash size: 0 B
	//
	// Last synchronized items:
	// 	file: 'NewFile'
	// 	file: 'File.ods'
	// 	file: 'downloads/file.deb'
	// 	file: 'downloads/setup'
	// 	file: 'download'
	// 	file: 'down'
	// 	file: 'do'
	// 	file: 'd'
	// 	file: 'o'
	// 	file: 'w'
}

// catch 5-th status change after start of synchronisation
func Example50StatusAfter5thEvent() {
	getStatusAfterEvent(time.Duration(1 * time.Second))
	// Output:
	// Synchronization core status: idle
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	Total: 43.50 GB
	// 	Used: 2.89 GB
	// 	Available: 40.61 GB
	// 	Max file size: 50 GB
	// 	Trash size: 0 B
	//
	// Last synchronized items:
	// 	file: 'File.ods'
	// 	file: 'downloads/file.deb'
	// 	file: 'downloads/setup'
	// 	file: 'download'
	// 	file: 'down'
	// 	file: 'do'
	// 	file: 'd'
	// 	file: 'o'
	// 	file: 'w'
	// 	file: 'n'
}

// start error simulation
func Example60Error() {
	execCommand("error")
	// Output:
	//
}

// catch status change right after start of error synchronisation
func Example62StatusAfterError() {
	execCommand("status")
	// Output:
	// Synchronization core status: error
	// Error: access error
	// Path: 'downloads/test1'
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	Total: 43.50 GB
	// 	Used: 2.88 GB
	// 	Available: 40.62 GB
	// 	Max file size: 50 GB
	// 	Trash size: 654.48 MB
	//
	// Last synchronized items:
	// 	file: 'File.ods'
	// 	file: 'downloads/file.deb'
	// 	file: 'downloads/setup'
	// 	file: 'download'
	// 	file: 'down'
	// 	file: 'do'
	// 	file: 'd'
	// 	file: 'o'
	// 	file: 'w'
	// 	file: 'n'
}

// catch 2-nd status change after start of error synchronisation
func Example64StatusAfter1stEvent() {
	getStatusAfterEvent(time.Duration(1 * time.Second))
	// Output:
	// Synchronization core status: idle
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	Total: 43.50 GB
	// 	Used: 2.89 GB
	// 	Available: 40.61 GB
	// 	Max file size: 50 GB
	// 	Trash size: 0 B
	//
	// Last synchronized items:
	// 	file: 'File.ods'
	// 	file: 'downloads/file.deb'
	// 	file: 'downloads/setup'
	// 	file: 'download'
	// 	file: 'down'
	// 	file: 'do'
	// 	file: 'd'
	// 	file: 'o'
	// 	file: 'w'
	// 	file: 'n'
}

// try to get status in empty enviroment
// it starts to fail on CircleCI - env is not installed in the image
// need some alternative solution to perform this test
//func Example80StatusInEnv() {
// 	exe, _ := exec.LookPath(exe)
// 	cmd := exec.Command("env", "-i", exe, "status")
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stdout
// 	cmd.Run()
// 	// Output:
// 	// Synchronization core status: idle
// 	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
// 	// 	Total: 43.50 GB
// 	// 	Used: 2.89 GB
// 	// 	Available: 40.61 GB
// 	// 	Max file size: 50 GB
// 	// 	Trash size: 0 B
// 	//
// 	// Last synchronized items:
// 	// 	file: 'File.ods'
// 	// 	file: 'downloads/file.deb'
// 	// 	file: 'downloads/setup'
// 	// 	file: 'download'
// 	// 	file: 'down'
// 	// 	file: 'do'
// 	// 	file: 'd'
// 	// 	file: 'o'
// 	// 	file: 'w'
// 	// 	file: 'n'
// }

// try to get status with removed sinc path
func Example90CommandWithoutDir() {
	os.RemoveAll(SyncDirPath)
	execCommand("status")
	// Output:
	// Error: Indicated directory does not exist
}

// try to stop daemon
func Example95Stop() {
	execCommand("stop")
	// Output:
	// Daemon stopped.
}

// try to stop daemon again
func Example97SecondaryStop() {
	execCommand("stop")
	// Output:
	// Error: daemon not started
}
