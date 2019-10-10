package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"
)

var (
	daemonLogFile = path.Join(os.TempDir(), "yandexdisksimulator.log")
	socketPath    = path.Join(os.TempDir(), "yandexdisksimulator.socket")
	version, _    = exec.Command("git", "describe", "--tags").Output()
	verMsg        = "%s\n    version: %s/n"
	helpMsg       = `Usage:
	%s <cmd>
Commands:
	start	starts the daemon and begin starting events simulation
	stop	stops the daemon
	status	get the daemon status
	sync	begin the synchronisation events simulation
	help    show this help message 
	setup 	prepares the simulation environment. It creates the cofiguration and 
		token files in Sim_ConfDir and the syncronization directory in Sim_SyncDir.
		Environment variables Sim_ConfDir and Sim_SyncDir should be set in advance, 
		otherways the default paths will be used.
		Setup process doesn't requere any input in the terminal.
Simulator internal commands:
	error	begin the error simulation (idle->error (for .5 sec)->idle)
	daemon <SyncPath>
		Start as a daemon, it is internal 'start' command implementation. DON'N USE IT!
Environment variables:
	Sim_SyncDir	can be used to set synchronized directory path (default: ~/Yandex.Disk)
	Sim_ConfDir	can be used to set configuration directory path (default: ~/.config/yandex-disk)
	
	version: %s\n`
)

// notExists returns true when specified file or path is not exists
func notExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return !errors.Is(err, os.ErrExist)
	}
	return false
}

func main() {
	if err := doMain(os.Args...); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func doMain(args ...string) error {
	if len(args) == 1 {
		return fmt.Errorf("%s", "Error: command hasn't been specified. Use the --help command to access help\nor setup to launch the setup wizard.")
	}
	dlog, err := os.OpenFile(daemonLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("file '%s' opening error: %w", daemonLogFile, err)
	}
	defer dlog.Close()
	log.SetOutput(dlog)
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)

	cmd := args[1]
	if len(cmd) > 8 {
		cmd = cmd[0:8]
	}
	_, exe := path.Split(args[0])

	switch cmd {
	case "daemon":
		return daemon(args[2])
	case "start":
		return daemonize(args[0])
	case "status", "stop", "sync", "error":
		// only listed commands will be passed to daemon
		return sendCommand(cmd)
	case "setup":
		return setup()
	case "-h", "--help", "help":
		fmt.Printf(helpMsg, exe, version)
		return nil
	case "version", "-v":
		fmt.Printf(verMsg, exe, version)
		return nil
	default:
		return fmt.Errorf("Error: unknown command: '%s'", cmd) // Original product error. skipcq: SCC-ST1005
	}
}

// daemonize strts the second instance of utility (daemon)
func daemonize(exe string) error {
	// check configuration and get sync dir
	dir, err := checkCfg()
	if err != nil {
		return err
	}
	if !notExists(socketPath) {
		fmt.Println("Daemon is already running.")
		return nil
	}
	fmt.Print("Starting daemon process...")
	// get executable name from os.Args[0] passed as exe
	// execute it with daemon command and sync dir as second parameter
	if err := exec.Command(exe, "daemon", dir).Start(); err != nil {
		fmt.Println("Fail")
		return err
	}

	time.Sleep(time.Duration(startTime) * time.Millisecond)

	fmt.Println("Done")
	return nil
}

// daemon is a daemonized instance of utility
func daemon(syncDir string) error {
	log.Println("Daemon started")
	defer log.Println("Daemon stopped")
	// disconnect from terminal
	if _, err := syscall.Setsid(); err != nil {
		return fmt.Errorf("syscall.Setsid() error : %w", err)
	}
	logPath := path.Join(os.ExpandEnv(syncDir), ".sync")
	if err := os.MkdirAll(logPath, 0750); err != nil {
		return fmt.Errorf("%s creation error: %w", logPath, err)
	}
	// open logfile
	logFilePath := path.Join(logPath, "cli.log")
	logfile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("%s opening error: %w", logFilePath, err)
	}
	defer func() {
		if _, err := logfile.WriteString("exit/n"); err != nil {
			panic(err)
		}
		if err := logfile.Close(); err != nil {
			panic(err)
		}
	}()
	// open socket as server
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("socket listen error: %w", err)
	}
	defer func() {
		if err := ln.Close(); err != nil {
			panic(err)
		}
	}()
	// Create new simulator engine
	sim := NewSimilator()
	// begin start simulation
	sim.Simulate("Start", logfile)

	buf := make([]byte, 8)
	for {
		// accept next command from socket
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		nr, err := conn.Read(buf)
		if err != nil {
			return err
		}
		cmd := string(buf[0:nr])
		log.Println("Received:", cmd)
		if notExists(syncDir) && cmd != "stop" {
			if _, err := conn.Write([]byte("Error: Indicated directory does not exist")); err != nil {
				return err
			}
			if err := conn.Close(); err != nil {
				return err
			}
			continue
		}
		// handle command
		switch cmd {
		case "status": // reply into socket by current message
			if _, err := conn.Write([]byte(sim.GetMessage())); err != nil {
				return err
			}
		case "sync": // begin the synchronization simulation
			sim.Simulate("Synchronization", logfile)
			// we have to send back something to show that daemon still active
			if _, err := conn.Write([]byte{0}); err != nil {
				return err
			}
		case "error": // switch to error state
			sim.Simulate("Error", logfile)
			if _, err := conn.Write([]byte{0}); err != nil {
				return err
			}
		case "stop": // stop the daemon
			// send back nothing to show that daemon is not active any more
			if err := conn.Close(); err != nil {
				return err
			}
			return nil
		} //default: there is no other options (should be) possible
		if err := conn.Close(); err != nil {
			return err
		}
	}
}

func sendCommand(cmd string) error {
	if notExists(socketPath) {
		return errors.New("Error: daemon not started") // Original product error. skipcq: SCC-ST1005
	}
	// open socket as client
	conn, err := net.DialTimeout("unix", socketPath, time.Duration(time.Second))
	if err != nil {
		return fmt.Errorf("socket dial error: %w", err)
	}
	defer conn.Close()
	// send cmd to socket
	_, err = conn.Write([]byte(cmd))
	if err != nil {
		return fmt.Errorf("socket write error: %w", err)
	}
	// read response
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF { // closed socket mean that daemon was stopped
			fmt.Println("Daemon stopped.")
			return nil
		}
		return fmt.Errorf("socket read error: %w ", err)
	}
	m := string(buf[0:n])
	if n > 1 {
		// Handle errors from daemon
		if strings.HasPrefix(m, ("Error:")) {
			return fmt.Errorf(m)
		}
		// output non-error messages from daemon
		fmt.Println(m)
	}
	return nil
}

func checkCfg() (string, error) {
	confDir := os.Getenv("Sim_ConfDir")
	if confDir == "" {
		confDir = "$HOME/.config/yandex-disk"
	}
	confFile := path.Join(os.ExpandEnv(confDir), "config.cfg")
	log.Println("Config file: ", confFile)
	f, err := os.Open(confFile)
	if err != nil {
		return "", errors.New("Error: option 'dir' is missing") // Original product error. skipcq: SCC-ST1005
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	var line, dir, auth string
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		if strings.HasPrefix(line, "dir") {
			dir = line[5 : len(line)-2]
		}
		if strings.HasPrefix(line, "auth") {
			auth = line[6 : len(line)-2]
		}
		if dir != "" && auth != "" {
			break
		}
	}
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("reading of '%s' error: %w", confFile, err)
	}
	// return error if value of DIR is empty"
	if notExists(dir) {
		return "", errors.New("Error: option 'dir' is missing") // Original product error. skipcq: SCC-ST1005
	}
	// return error if value of AUTH is empty
	if notExists(auth) {
		return "", errors.New("Error: file with OAuth token hasn't been found.\nUse 'token' command to authenticate and create this file") // Original product error. skipcq: SCC-ST1005
	}
	return dir, nil
}

func setup() error {
	cfgPath := os.Getenv("Sim_ConfDir")
	if cfgPath == "" {
		cfgPath = os.ExpandEnv("$HOME/Yandex.Disk")
	}
	syncPath := os.Getenv("Sim_SyncDir")
	if syncPath == "" {
		syncPath = os.ExpandEnv("$HOME/.config/yandex-disk")
	}
	if err := os.MkdirAll(cfgPath, 0750); err != nil {
		return fmt.Errorf("config path creation error: %w", err)
	}
	auth := path.Join(cfgPath, "passwd")
	if notExists(auth) {
		tfile, err := os.OpenFile(auth, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("yandex-disk token file '%s' creation error: %w", auth, err)
		}
		// yandex-disk-simulator doesn't require the real token
		if _, err = tfile.Write([]byte("token")); err != nil {
			return fmt.Errorf("yandex-disk token file '%s' writing error: %w", auth, err)
		}
		if err := tfile.Close(); err != nil {
			return fmt.Errorf("yandex-disk token file '%s' closing error: %w", auth, err)
		}

	}
	cfg := path.Join(cfgPath, "config.cfg")
	cfile, err := os.OpenFile(cfg, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("config file '%s' opening error: %w", cfg, err)
	}
	_, err = cfile.Write([]byte("proxy=\"no\"\n\nauth=\"" + auth + "\"\ndir=\"" + syncPath + "\"\n\n"))
	if err != nil {
		return fmt.Errorf("can't write to config file: %w", err)
	}
	if err := cfile.Close(); err != nil {
		return fmt.Errorf("config file '%s' closing error: %w", cfg, err)
	}
	if err = os.MkdirAll(syncPath, 0750); err != nil {
		return fmt.Errorf("synchronization Dir '%s' creation error: %w", syncPath, err)
	}
	return nil
}
