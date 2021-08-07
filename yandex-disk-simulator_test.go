package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

var (
	SyncDirPath    = "$HOME/TeSt_Yandex.Disk_TeSt"         // testting synchronisation path
	ConfigFilePath = "$HOME/.config/TeSt_yandex-disk_TeSt" // testting configuration path
)

const (
	// default executable name
	exe = "yandex-disk-simulator"
)

func getOutput() func() string {
	stdOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	return func() string {
		os.Stdout = stdOut
		w.Close()
		out, _ := io.ReadAll(r)
		return string(out)
	}
}

// execute command and capture stdout
func execCommand(t *testing.T, command string) string {
	out := getOutput()
	err := doMain(exe, command)
	res := out()
	assert.NoError(t, err)
	return res
}

// try to catch the daemon log update during specified timeout
func getStatusAfterEvent(t *testing.T, timeout time.Duration) string {
	watch, err := fsnotify.NewWatcher()
	assert.NoError(t, err)
	defer watch.Close()
	assert.NoError(t, watch.Add(filepath.Join(SyncDirPath, ".sync/cli.log")))
	select {
	case err := <-watch.Errors:
		t.Fatal(err.Error())
		return ""
	case <-watch.Events:
		return execCommand(t, "status")
	case <-time.After(timeout):
		t.Fatal("No event within timaout")
		return ""
	}
}

func TestMain(m *testing.M) {
	flag.Parse()

	// set environment variables for setup of simulator
	SyncDirPath = os.ExpandEnv(SyncDirPath)
	os.Setenv("Sim_SyncDir", SyncDirPath)
	ConfigFilePath = os.ExpandEnv(ConfigFilePath)
	os.Setenv("Sim_ConfDir", ConfigFilePath)

	// Run tests
	errn := m.Run()

	// Clearance
	os.RemoveAll(ConfigFilePath)
	os.RemoveAll(SyncDirPath)

	os.Exit(errn)
}

// try to start utility without command
func TestDoMain00NoCommand(t *testing.T) {
	err := doMain(exe)
	assert.Error(t, err, "no error for no command")
	assert.Equal(t,
		errors.New("Error: command hasn't been specified. Use the --help command to access help\nor setup to launch the setup wizard."),
		err,
		"incorrect message: "+err.Error())
}

// try to start with 'help' command
func TestDoMain01Help(t *testing.T) {
	out := getOutput()
	args := os.Args
	os.Args = []string{exe, "help"}
	main()
	os.Args = args
	output := out()
	if output != fmt.Sprintf(helpMsg, exe, version) {
		t.Error("incorrect message:", output)
	}
}

// try to ask for utility version
func TestDoMain01Version(t *testing.T) {
	res := execCommand(t, "-v")
	assert.Equalf(t, fmt.Sprintf(verMsg, exe, version), res, "incorrect message: %s", res)
}

// try to start with wrong and long command
func TestDoMain02WrongCommand(t *testing.T) {
	err := doMain(exe, "wrongCMD_cut_it")
	assert.Equalf(t, errors.New("Error: unknown command: 'wrongCMD'"), err, "incorrect message: %v", err)
}

// try to start without configuration
func TestDoMain04StartNoConfig(t *testing.T) {
	err := doMain(exe, "start")
	assert.Error(t, err, "no error for start without config")
	assert.Equalf(t, errors.New("Error: option 'dir' is missing"), err, "incorrect message: %v", err)
}

// try to setup the configuration
func TestDoMain05Setup(t *testing.T) {
	assert.NoError(t, doMain(exe, "setup"))
}

// try 'ststus' command with not started daemon
func TestDoMain07Command2NotStarted(t *testing.T) {
	err := doMain(exe, "status")
	assert.Error(t, err, "no error for command to not started")
	assert.Equalf(t, errors.New("Error: daemon not started"), err, "incorrect message: %v", err)
}

// try to start daemon with wrong executable name
func TestDoMain08FailWrongDaemonStart(t *testing.T) {
	assert.NoError(t, doMain(exe, "setup"))
	out := getOutput()
	err := doMain("wrong-simulator", "start")
	res := out()
	assert.Error(t, err, "No error with starting of incorrect daemon")
	assert.Equalf(t, "Starting daemon process...Fail\n", res, "incorrect message: %s", res)
}

// try to start echo daemon
func TestDoMain09StartEcho(t *testing.T) {
	assert.NoError(t, doMain(exe, "setup"))
	out := getOutput()
	err := doMain("echo", "start")
	res := out()
	assert.NoError(t, err)
	assert.Equal(t, "Starting daemon process...Done\n", res)
}

// try to start configured daemon
func TestDoMain10StartSuccess(t *testing.T) {
	assert.NoError(t, doMain(exe, "setup"))
	// start daemon in seporate goroutine
	go doMain(exe, "daemon", SyncDirPath)
	time.Sleep(time.Millisecond)
	t.Run("second start", func(t *testing.T) {
		res := execCommand(t, "start")
		assert.Equalf(t, "Daemon is already running.\n", res, "incorrect message: %s"+res)
	})

	t.Run("empty status after start", func(t *testing.T) {
		res := execCommand(t, "status")
		assert.Equalf(t, "", res, "incorrect message: %s", res)
	})

	t.Run("status after event #1", func(t *testing.T) {
		assert.Equal(t,
			`Synchronization core status: paused
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	The quota has not been received yet.

Last synchronized items:
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'
	file: 'n'`+"\n\n\n",
			getStatusAfterEvent(t, 2*time.Second))
	})

	t.Run("status after event #2", func(t *testing.T) {
		assert.Equal(t,
			`Synchronization core status: index
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	The quota has not been received yet.`+"\n\n\n",
			getStatusAfterEvent(t, 2*time.Second))
	})

	t.Run("status after event #3", func(t *testing.T) {
		assert.Equal(t,
			`Synchronization core status: busy
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	The quota has not been received yet.

Last synchronized items:
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'
	file: 'n'`+"\n\n\n",
			getStatusAfterEvent(t, 2*time.Second))
	})

	t.Run("status after event #4", func(t *testing.T) {
		assert.Equal(t,
			`Synchronization core status: index
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	The quota has not been received yet.

Last synchronized items:
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'
	file: 'n'`+"\n\n\n",
			getStatusAfterEvent(t, 2*time.Second))
	})

	t.Run("status after event #5", func(t *testing.T) {
		assert.Equal(t,
			`Synchronization core status: idle
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	Total: 43.50 GB
	Used: 2.89 GB
	Available: 40.61 GB
	Max file size: 50 GB
	Trash size: 0 B

Last synchronized items:
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'
	file: 'n'`+"\n\n\n",
			getStatusAfterEvent(t, 6*time.Second))
	})

	t.Run("sunc", func(t *testing.T) {
		assert.Empty(t, execCommand(t, "sync"))
	})

	t.Run("status after sync", func(t *testing.T) {
		assert.Equal(t,
			`Synchronization core status: index
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	Total: 43.50 GB
	Used: 2.89 GB
	Available: 40.61 GB
	Max file size: 50 GB
	Trash size: 0 B

Last synchronized items:
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'
	file: 'n'`+"\n\n\n",
			execCommand(t, "status"))
	})

	t.Run("status after sync event #2", func(t *testing.T) {
		assert.Equal(t,
			`Sync progress: 0 MB/ 139.38 MB (0 %)
Synchronization core status: busy
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	Total: 43.50 GB
	Used: 2.89 GB
	Available: 40.61 GB
	Max file size: 50 GB
	Trash size: 0 B

Last synchronized items:
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'
	file: 'n'`+"\n\n\n",
			getStatusAfterEvent(t, 1*time.Second))
	})

	t.Run("status after sync event #3", func(t *testing.T) {
		assert.Equal(t,
			`Sync progress: 65.34 MB/ 139.38 MB (46 %)
Synchronization core status: busy
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	Total: 43.50 GB
	Used: 2.89 GB
	Available: 40.61 GB
	Max file size: 50 GB
	Trash size: 0 B

Last synchronized items:
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'
	file: 'n'`+"\n\n\n",
			getStatusAfterEvent(t, 2*time.Second))
	})

	t.Run("status after sync event #4", func(t *testing.T) {
		assert.Equal(t,
			`Sync progress: 139.38 MB/ 139.38 MB (100 %)
Synchronization core status: index
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	Total: 43.50 GB
	Used: 2.89 GB
	Available: 40.61 GB
	Max file size: 50 GB
	Trash size: 0 B

Last synchronized items:
	file: 'NewFile'
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'`+"\n\n\n",
			getStatusAfterEvent(t, 2*time.Second))
	})

	t.Run("status after sync event #5", func(t *testing.T) {
		assert.Equal(t,
			`Synchronization core status: idle
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	Total: 43.50 GB
	Used: 2.89 GB
	Available: 40.61 GB
	Max file size: 50 GB
	Trash size: 0 B

Last synchronized items:
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'
	file: 'n'`+"\n\n\n",
			getStatusAfterEvent(t, 2*time.Second))
	})

	t.Run("error", func(t *testing.T) {
		assert.Empty(t, execCommand(t, "error"))
	})

	t.Run("status after error", func(t *testing.T) {
		assert.Equal(t,
			`Synchronization core status: error
Error: access error
Path: 'downloads/test1'
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	Total: 43.50 GB
	Used: 2.88 GB
	Available: 40.62 GB
	Max file size: 50 GB
	Trash size: 654.48 MB

Last synchronized items:
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'
	file: 'n'`+"\n\n\n",
			execCommand(t, "status"))
	})

	t.Run("status after error event #1", func(t *testing.T) {
		assert.Equal(t,
			`Synchronization core status: idle
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	Total: 43.50 GB
	Used: 2.89 GB
	Available: 40.61 GB
	Max file size: 50 GB
	Trash size: 0 B

Last synchronized items:
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'
	file: 'n'`+"\n\n\n",
			getStatusAfterEvent(t, 2*time.Second))
	})

	t.Run("status in empty environmen", func(t *testing.T) {
		exe, _ := exec.LookPath(exe)
		cmd := exec.Command("env", "-i", exe, "status")
		bufer := &bytes.Buffer{}
		cmd.Stdout = bufer
		cmd.Stderr = bufer
		cmd.Run()
		res := bufer.String()
		assert.Equal(t,
			`Synchronization core status: idle
Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'
	Total: 43.50 GB
	Used: 2.89 GB
	Available: 40.61 GB
	Max file size: 50 GB
	Trash size: 0 B

Last synchronized items:
	file: 'File.ods'
	file: 'downloads/file.deb'
	file: 'downloads/setup'
	file: 'download'
	file: 'down'
	file: 'do'
	file: 'd'
	file: 'o'
	file: 'w'
	file: 'n'`+"\n\n\n",
			res)
	})

	t.Run("status with removed sinc path", func(t *testing.T) {
		os.RemoveAll(SyncDirPath)
		out := getOutput()
		err := doMain(exe, "status")
		res := out()
		assert.Equal(t, "Error: Indicated directory does not exist", err.Error())
		assert.Empty(t, res)
	})

	t.Run("stop", func(t *testing.T) {
		out := getOutput()
		err := doMain(exe, "stop")
		time.Sleep(time.Millisecond * 150)
		res := out()
		assert.NoError(t, err)
		assert.Equal(t, "Daemon stopped.\n", res)
	})

	t.Run("second stop", func(t *testing.T) {
		out := getOutput()
		err := doMain(exe, "stop")
		time.Sleep(time.Millisecond * 150)
		res := out()
		assert.Equal(t, "Error: daemon not started", err.Error())
		assert.Empty(t, res)
	})

}
