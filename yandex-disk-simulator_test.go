package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	SyncDirPath    = "$HOME/TeSt_Yandex.Disk_TeSt"
	ConfigFilePath = "$HOME/.config/TeSt_Yandex.Disk_TeSt"
)

func TestMain(m *testing.M) {
	flag.Parse()

	exec.Command("go", "build", "yandex-disk-simulator.go").Run()
	cwd, _ := os.Getwd()
	os.Setenv("PATH", cwd+":"+os.Getenv("PATH"))
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

func TestDoMain00NoCommand(t *testing.T) {
	err := doMain([]string{"yandex-disk-simulator"})
	if err == nil {
		t.Error("no error for no command")
	}
	if err.Error() != "Error: command hasn't been specified. Use the --help command to access help\nor setup to launch the setup wizard." {
		t.Error("incorrect message for no command case")
	}
}

func TestDoMain01Help(t *testing.T) {
	res := testCmdWithCapture("help", t)
	if res != helpMsg+"\n" {
		t.Error("incorrect message for help case:", res)
	}
}

func TestDoMain02WrongCommand(t *testing.T) {
	err := doMain([]string{"yandex-disk-simulator", "wrongCMD_cut_it"})
	if err == nil {
		t.Error("no error for wrong command")
	}
	if err.Error() != "Error: unknown command: 'wrongCMD'" {
		t.Error("incorrect message for wrong command case")
	}
}

func TestDoMain04StartNoConfig(t *testing.T) {
	err := doMain([]string{"yandex-disk-simulator", "start"})
	if err == nil {
		t.Error("no error for start without config")
	}
	if err.Error() != "Error: option 'dir' is missing" {
		t.Error("incorrect message for start without config case")
	}
}

func TestDoMain05Setup(t *testing.T) {
	err := doMain([]string{"yandex-disk-simulator", "setup"})
	if err != nil {
		t.Error("error for setup :", err)
	}
}

func TestDoMain07Command2NotStarted(t *testing.T) {
	err := doMain([]string{"yandex-disk-simulator", "status"})
	if err == nil {
		t.Error("no error for command to not started")
	}
	if err.Error() != "Error: daemon not started" {
		t.Error("incorrect message for command to not started case: " + err.Error())
	}
}

func testCmdWithCapture(cmd string, t *testing.T) string {
	stdOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := doMain([]string{"yandex-disk-simulator", cmd})
	if err != nil {
		t.Error("error for ", cmd, ":", err)
	}
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdOut
	return string(out)
}
func TestDoMain10StartSuccess(t *testing.T) {
	res := testCmdWithCapture("start", t)
	if res != "Starting daemon process...Done\n" {
		t.Error("incorrect message for start without config case:", res)
	}
}

func TestDoMain11StartSecondary(t *testing.T) {
	res := testCmdWithCapture("start", t)
	if res != "Daemon is already running.\n" {
		t.Error("incorrect message for secondary start case:", res)
	}
}

func TestDoMain15StartDaemon(t *testing.T) {
	// stop already executed daemon
	exec.Command("yandex-disk-simulator", "stop").Run()
	// start daemon in separate gorutine
	go doMain([]string{"yandex-disk-simulator", "daemon"})
	time.Sleep(10 * time.Millisecond)
}

func TestDoMain17Status(t *testing.T) {
	res := testCmdWithCapture("status", t)
	if res != "" {
		t.Error("incorrect message for status case:", res)
	}
}

func execCommand(command string) {
	err := doMain([]string{"yandex-disk-simulator", command})
	if err != nil {
		fmt.Println(err.Error())
	}
}

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

func Example45StatusAfter1stEvent() {
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

func Example50StatusAfter2ndEvent() {
	getStatusAfterEvent(time.Duration(2 * time.Second))
	// Output:
	// Synchronization core status: index
	// Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	// 	The quota has not been received yet.
}

func Example55StatusAfter3rdEvent() {
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

func Example60StatusAfter4rdEvent() {
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

func Example65StatusAfter5rdEvent() {
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

func Example70Sync() {
	execCommand("sync")
	// Output:
	//
}

func Example75StatusAfterSyncStart() {
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

func Example80StatusAfter2ndEvent() {
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

func Example85StatusAfter3rdEvent() {
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

func Example87StatusAfter4rdEvent() {
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

func Example88StatusAfter5rdEvent() {
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

func Example89StatusInEnv() {
	exe, _ := exec.LookPath("yandex-disk-simulator")
	cmd := exec.Command("env", "-i", exe, "status")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
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

func Example90Stop() {
	execCommand("stop")
	// Output:
	// Daemon stopped.
}

func Example90SecondaryStop() {
	execCommand("stop")
	// Output:
	// Error: daemon not started
}
