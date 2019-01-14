package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Event - the stucture for change event
type Event struct {
	msg      string        // status message
	duration time.Duration // event duration
	log      string        // cli.log message or no message if it is ""
}

const (
	daemonLogFile = "/tmp/yandexdisksymulator.log"
	socketPath    = "/tmp/yandexdisksymulator.socket"
	helpMsg       = `Usage:
	yandex-disk-similator <cmd>
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
Simulator commands:
	daemon	start as a daemon (Don't use it !!!)
Environment variables:
	Sim_SyncDir	can be used to set synchronized directory path (default: ~/Yandex.Disk)
	Sim_ConfDir	can be used to set configuration directory path (default: ~/.config/yandex-disk)`
)

var (
	cfgPath          = ""
	syncPath         = ""
	message          = " "
	msgLock, symLock sync.Mutex

	msgIdle   = "Synchronization core status: idle\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tTotal: 43.50 GB\n\tUsed: 2.89 GB\n\tAvailable: 40.61 GB\n\tMax file size: 50 GB\n\tTrash size: 0 B\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n"
	startTime = 1000

	// start events sequence
	startSequence = &[]Event{
		Event{" ",
			time.Duration(1600) * time.Millisecond,
			""},
		Event{"Synchronization core status: paused\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tThe quota has not been received yet.\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
			time.Duration(250) * time.Millisecond,
			"Start simulation 1"},
		Event{"Synchronization core status: index\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tThe quota has not been received yet.\n\n",
			time.Duration(600) * time.Millisecond,
			"Start simulation 2"},
		Event{"Synchronization core status: busy\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tThe quota has not been received yet.\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
			time.Duration(100) * time.Millisecond,
			"Start simulation 3"},
		Event{"Synchronization core status: index\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tThe quota has not been received yet.\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
			time.Duration(4600) * time.Millisecond,
			"Start simulation 4"},
	}

	// synchronization events sequence
	syncSequence = &[]Event{
		Event{"Synchronization core status: index\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tTotal: 43.50 GB\n\tUsed: 2.89 GB\n\tAvailable: 40.61 GB\n\tMax file size: 50 GB\n\tTrash size: 0 B\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
			time.Duration(900) * time.Millisecond,
			"Synchronization simulation started"},
		Event{"Sync progress: 0 MB/ 139.38 MB (0 %)\nSynchronization core status: busy\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tTotal: 43.50 GB\n\tUsed: 2.89 GB\n\tAvailable: 40.61 GB\n\tMax file size: 50 GB\n\tTrash size: 0 B\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
			time.Duration(100) * time.Millisecond,
			"Synchronization simulation 1"},
		Event{"Sync progress: 65.34 MB/ 139.38 MB (46 %)\nSynchronization core status: busy\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tTotal: 43.50 GB\n\tUsed: 2.89 GB\n\tAvailable: 40.61 GB\n\tMax file size: 50 GB\n\tTrash size: 0 B\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
			time.Duration(1500) * time.Millisecond,
			"Synchronization simulation 2"},
		Event{"Sync progress: 139.38 MB/ 139.38 MB (100 %)\nSynchronization core status: index\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tTotal: 43.50 GB\n\tUsed: 2.89 GB\n\tAvailable: 40.61 GB\n\tMax file size: 50 GB\n\tTrash size: 0 B\n\nLast synchronized items:\n\tfile: 'NewFile'\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\n",
			time.Duration(500) * time.Millisecond,
			"Synchronization simulation 3"},
	}
)

func simulate(name string, seq *[]Event, l io.Writer) {
	symLock.Lock()
	for _, e := range *seq {
		setMsg(e.msg)
		if e.log != "" {
			l.Write([]byte(e.log + "\n"))
		}
		time.Sleep(e.duration)
	}
	setMsg(msgIdle)
	l.Write([]byte(name + " simulation finished\n"))
	symLock.Unlock()
}

func setMsg(m string) {
	// thread safe message update
	msgLock.Lock()
	message = m
	msgLock.Unlock()
}

func notExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsNotExist(err)
	}
	return false
}

func initLog() (*os.File, error) {
	dlog, err := os.OpenFile(daemonLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return nil, fmt.Errorf("%s opening error: %v", daemonLogFile, err)
	}
	log.SetOutput(dlog)
	return dlog, nil
}

func main() {
	err := doMain(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// doMain is main(). It is for tests
func doMain(args []string) error {
	if len(args) == 1 {
		return fmt.Errorf("%s", "Error: command hasn't been specified. Use the --help command to access help\nor setup to launch the setup wizard.")
	}
	daemonlfile, err := initLog()
	if err != nil {
		return err
	}
	defer daemonlfile.Close()
	cmd := args[1]
	if len(cmd) > 8 {
		cmd = cmd[0:8]
	}
	switch cmd {
	case "daemon":
		return daemon()
	case "start":
		return daemonize(args[0])
	case "status", "stop", "sync":
		// only listed commands will be passed to daemon
		return socketIneract(cmd)
	case "setup":
		return setup()
	case "-h", "--help", "help":
		fmt.Println(helpMsg)
		return nil
	default:
		return fmt.Errorf("Error: unknown command: '" + cmd + "'")
	}
}

func daemonize(exe string) error {
	err := checkCfg()
	if err != nil {
		return err
	}
	if !notExists(socketPath) {
		fmt.Println("Daemon is already running.")
		return nil
	}
	fmt.Print("Starting daemon process...")
	// get executable name
	_, exe = filepath.Split(exe)
	// execute it with daemon command
	cmd := exec.Command(exe, "daemon")
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(startTime) * time.Millisecond)
	fmt.Println("Done")
	return nil
}

func daemon() error {
	log.Println("Daemon started")
	defer log.Println("Daemon stopped")
	// disconnect from terminal
	_, err := syscall.Setsid()
	if err != nil {
		return fmt.Errorf("syscall.Setsid() error : %v", err)
	}
	err = checkCfg()
	if err != nil {
		return err
	}
	// create ~/<SyncDir>/.sync/cli.log if it is not exists
	syncDir := os.Getenv("Sim_SyncDir")
	if syncDir == "" {
		syncDir = "$HOME/Yandex.Disk"
	}
	logPath := filepath.Join(os.ExpandEnv(syncDir), ".sync")
	err = os.MkdirAll(logPath, 0755)
	if err != nil {
		return fmt.Errorf(logPath+" creation error:", err)
	}
	// open logfile
	logFilePath := filepath.Join(logPath, "cli.log")
	logfile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf(logFilePath+" opening error:", err)
	}
	defer func() {
		logfile.Write([]byte("exit/n"))
		logfile.Close()
	}()
	// open socket as server
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listen error: %v", err)
	}
	defer ln.Close()

	// begin start simulation
	go simulate("Start", startSequence, logfile)

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
		// react on command ...
		switch cmd {
		case "status": // replay to socket by current message
			msgLock.Lock()
			conn.Write([]byte(message))
			msgLock.Unlock()
		case "sync": // begin the synchronization simulation
			go simulate("Synchronization", syncSequence, logfile)
			conn.Write([]byte(" "))
		case "stop": // stop the daemon
			conn.Close()
			return nil
			//default: there is no other options (should be) possible
		}
		conn.Close()
	}
}

func socketIneract(cmd string) error {
	if notExists(socketPath) {
		// output error to stdout and exit with nonzero error code
		return fmt.Errorf("Error: daemon not started")
	}
	// open socket as client
	c, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("Socket dial error: %v", err)
	}
	defer c.Close()
	// send cmd to socket
	_, err = c.Write([]byte(cmd))
	if err != nil {
		return fmt.Errorf("Socket write error: %v", err)
	}
	// read reply
	buf := make([]byte, 512)
	n, err := c.Read(buf[:])
	if err != nil {
		if err == io.EOF { // closed socket mean that daemon was stopped
			fmt.Println("Daemon stopped.")
			return nil
		}
		return fmt.Errorf("Socket read error: %v ", err)
	}
	// output non-empty reply to stdout
	m := string(buf[0:n])
	if m != " " {
		fmt.Println(m)
	}
	return nil
}

func checkCfg() error {
	confDir := os.Getenv("Sim_ConfDir")
	if confDir == "" {
		confDir = "$HOME/.config/yandex-disk"
	}
	confFile := filepath.Join(os.ExpandEnv(confDir), "config.cfg")
	log.Println("Config file: ", confFile)
	f, err := os.Open(confFile)
	if err != nil {
		return fmt.Errorf("%s", "Error: option 'dir' is missing")
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
		return err
	}
	// for empty value DIR return "Error: option 'dir' is missing"
	if notExists(dir) {
		return fmt.Errorf("%s", "Error: option 'dir' is missing")
	}
	// for empty value AUTH return "Error: file with OAuth token hasn't been found.\nUse 'token' command to authenticate and create this file"
	if notExists(auth) {
		return fmt.Errorf("%s", "Error: file with OAuth token hasn't been found.\nUse 'token' command to authenticate and create this file")
	}
	return nil
}

func setup() error {
	cfgPath = os.Getenv("Sim_ConfDir")
	if cfgPath == "" {
		cfgPath = os.ExpandEnv("$HOME/Yandex.Disk")
	}
	syncPath = os.Getenv("Sim_SyncDir")
	if syncPath == "" {
		syncPath = os.ExpandEnv("$HOME/.config/yandex-disk")
	}
	err := os.MkdirAll(cfgPath, 0777)
	if err != nil {
		return fmt.Errorf("Config path creation error")
	}
	auth := filepath.Join(cfgPath, "passwd")
	if notExists(auth) {
		tfile, err := os.OpenFile(auth, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return fmt.Errorf("yandex-disk token file creation error: %v", err)
		}
		defer tfile.Close()
		_, err = tfile.Write([]byte("token")) // yandex-disk-simulator doesn't require the real token
		if err != nil {
			return fmt.Errorf("yandex-disk token file write error: %v", err)
		}
	}
	cfg := filepath.Join(cfgPath, "config.cfg")
	cfile, err := os.OpenFile(cfg, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer cfile.Close()
	_, err = cfile.Write([]byte("proxy=\"no\"\n\nauth=\"" + auth + "\"\ndir=\"" + syncPath + "\"\n\n"))
	if err != nil {
		return fmt.Errorf("Can't create config file: %v", err)
	}
	err = os.MkdirAll(syncPath, 0777)
	if err != nil {
		return fmt.Errorf("synchronization Dir creation error: %v", err)
	}
	return nil
}
