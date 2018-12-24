package main

import (
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

var (
	socketPath = "/tmp/yandexdiskmock.socket"
	startTime  = 1000

	message          = " "
	msgLock, symLock sync.Mutex

	msgIdle = "Synchronization core status: idle\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			Total: 43.50 GB\n			Used: 2.89 GB\n			Available: 40.61 GB\n		Max file size: 50 GB\n			Trash size: 0 B\n	\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	"

	// start sequence
	startSequence = []Event{
		Event{" ",
			time.Duration(1600) * time.Millisecond,
			""},
		Event{"Synchronization core status: paused\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			The quota has not been received yet.\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	",
			time.Duration(250) * time.Millisecond,
			"Start simulation 1"},
		Event{"Synchronization core status: index\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			The quota has not been received yet.\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	",
			time.Duration(600) * time.Millisecond,
			"Start simulation 2"},
		Event{"Synchronization core status: busy\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			The quota has not been received yet.\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	",
			time.Duration(100) * time.Millisecond,
			"Start simulation 3"},
		Event{"Synchronization core status: index\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			The quota has not been received yet.\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	",
			time.Duration(4600) * time.Millisecond,
			"Start simulation 4"},
	}

	// sync sequence
	syncSequence = []Event{
		Event{"Synchronization core status: index\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			Total: 43.50 GB\n			Used: 2.89 GB\n			Available: 40.61 GB\n		Max file size: 50 GB\n			Trash size: 0 B\n	\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	",
			time.Duration(900) * time.Millisecond,
			"Sync simulation started"},
		Event{"Sync progress: 65.34 MB/ 139.38 MB (46 %)\nSynchronization core status: index\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			Total: 43.50 GB\n			Used: 2.89 GB\n			Available: 40.61 GB\n		Max file size: 50 GB\n			Trash size: 0 B\n	\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	",
			time.Duration(100) * time.Millisecond,
			"Sync simulation 1"},
		Event{"Sync progress: 65.34 MB/ 139.38 MB (46 %)\nSynchronization core status: busy\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			Total: 43.50 GB\n			Used: 2.89 GB\n			Available: 40.61 GB\n		Max file size: 50 GB\n			Trash size: 0 B\n	\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	",
			time.Duration(1500) * time.Millisecond,
			"Sync simulation 2"},
		Event{"Sync progress: 139.38 MB/ 139.38 MB (100 %)\nSynchronization core status: index\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			Total: 43.50 GB\n			Used: 2.89 GB\n			Available: 40.61 GB\n		Max file size: 50 GB\n			Trash size: 0 B\n	\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	",
			time.Duration(500) * time.Millisecond,
			"Sync simulation 3"},
	}
)

func simulateSync(l io.Writer) {
	symLock.Lock()
	for _, e := range syncSequence {
		setMsg(e.msg)
		l.Write([]byte(e.log + "\n"))
		time.Sleep(e.duration)
	}
	setMsg(msgIdle)
	l.Write([]byte("Sync finished\n"))
	symLock.Unlock()
}

func simulateStart(l io.Writer) {
	// start sequence
	symLock.Lock()
	for _, e := range startSequence {
		setMsg(e.msg)
		l.Write([]byte(e.log + "\n"))
		time.Sleep(e.duration)
	}
	setMsg(msgIdle)
	l.Write([]byte("Start simulation finished\n"))
	symLock.Unlock()
}

func setMsg(m string) {
	// thread sa
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

func main() {
	// check config file
	if len(os.Args) == 1 {
		fmt.Println("Error: command hasn't been specified. Use the --help command to access help\nor setup to launch the setup wizard.")
		os.Exit(1)
	} else {
		cmd := os.Args[1]
		if len(cmd) > 16 {
			cmd = cmd[0:16]
		}
		switch cmd {
		case "daemon":
			daemon()
		case "start":
			checkCfg()
			if notExists(socketPath) {
				daemonize()
				fmt.Print("Starting daemon process...")
				time.Sleep(time.Duration(startTime) * time.Millisecond)
				fmt.Println("Done")
			} else {
				fmt.Println("Daemon is already running.")
			}
		case "--help", "help":
			fmt.Println(`Usage:
	yandex-disk-similator <cmd>
Commands:
	start	starts the daemon and begin starting events simulation
	stop	stops the daemon
	status	get the daemon status
	sync	begin the synchronisation events simulation 
Simulator commands:
	daemon	start as a daemon (don't use it)
Environment variables:
	DEBUG_SyncDir	can be used to set synchronized directory path (default: ~/Yandex.Disk)
	DEBUG_ConfDir	can be used to set configuration directory path (default: ~/.config/yandex-disk)`)
			//	case "status", "stop":
			//		if notExists(expandHome("~/Yandex.Disk")) {
			//			fmt.Println("Error: Indicated directory does not exist")
			//			os.Exit(1)
			//		} else {
			//			socketSend(cmd)
			//		}
		default:
			socketSend(cmd)
		}
	}
}

func daemonize() {
	// get executable name
	_, exe := filepath.Split(os.Args[0])
	// execute it with daemon command
	cmd := exec.Command(exe, "daemon")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func daemon() {
	dlog, err := os.OpenFile("/tmp/yandexdiskmock.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal("/tmp/yandexdiskmock.log opening error:", err)
	}
	defer dlog.Close()
	log.SetOutput(dlog)
	log.Println("Daemon started")
	// disconnect from terminal
	_, err = syscall.Setsid()
	if err != nil {
		log.Fatal("syscall.Setsid() error:", err)
	}
	// create ~/<SyncDir>/.sync/cli.log if it is not exists
	syncDir := os.Getenv("DEBUG_SyncDir")
	if syncDir == "" {
		syncDir = "$HOME/Yandex.Disk"
	}
	logPath := filepath.Join(os.ExpandEnv(syncDir), ".sync")
	err = os.MkdirAll(logPath, 0755)
	if err != nil {
		log.Fatal(logPath+" creation error:", err)
	}
	// open logfile
	logFilePath := filepath.Join(logPath, "cli.log")
	logfile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal(logFilePath+" opening error:", err)
	}
	defer func() {
		logfile.Write([]byte("exit\n"))
		logfile.Close()
	}()
	// open socket as server
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}
	// defer closing of socket and log file
	defer ln.Close()

	// start simulation of start
	go simulateStart(logfile)

	buf := make([]byte, 16)
	for {
		// read next command from socket
		fd, err := ln.Accept()
		if err != nil {
			return
		}
		nr, err := fd.Read(buf)
		if err != nil {
			return
		}
		cmd := string(buf[0:nr])
		log.Println("Received:", cmd)
		// react on command ...
		switch cmd {
		case "stop":
			return
		case "status":
			msgLock.Lock()
			fd.Write([]byte(message))
			msgLock.Unlock()
		case "sync":
			go simulateSync(logfile)
			fd.Write([]byte(" "))
		default:
			fd.Write([]byte("Error: unknown command: '" + cmd + "'"))
		}
	}
}

func socketSend(cmd string) {
	if notExists(socketPath) {
		// output error to stdot and exit with nonzero error code
		fmt.Println("Error: daemon not started")
		os.Exit(1)
	} else {
		// open socket as client
		c, err := net.Dial("unix", socketPath)
		if err != nil {
			log.Fatal("Socket dial error", err)
		}
		defer c.Close()
		// send cmd to socket
		_, err = c.Write([]byte(cmd))
		if err != nil {
			log.Fatal("Socket write error", err)
		}
		buf := make([]byte, 512)
		// wait for replay
		n, err := c.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				fmt.Println("Daemon stopped.")
				return
			}
			log.Fatal("Socket read error", err)
		}
		// output replay to stdout
		m := string(buf[0:n])
		if m != " " {
			fmt.Println(m)
		}
	}
	return
}

func checkCfg() {
	confDir := os.Getenv("DEBUG_ConfDir")
	if confDir == "" {
		confDir = "$HOME/.config/yandex-disk"
	}
	confFile := filepath.Join(os.ExpandEnv(confDir), "config.cfg")
	log.Println("Config file: ", confFile)
	f, err := os.Open(confFile)
	if err != nil {
		fmt.Println("Error: option 'dir' is missing --")
		os.Exit(1)
	}
	defer f.Close()
	reader := io.Reader(f)
	var line, dir, auth string
	var n int
	for {
		n, err = fmt.Fscanln(reader, &line)
		if n == 0 {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if strings.HasPrefix(line, "dir") {
			dir = line[5 : len(line)-1]
		}
		if strings.HasPrefix(line, "auth") {
			auth = line[6 : len(line)-1]
		}
		if dir != "" && auth != "" {
			break
		}
	}
	// for empty value DIR return "Error: option 'dir' is missing"
	if notExists(dir) {
		fmt.Println("Error: option 'dir' is missing")
		os.Exit(1)
	}
	// for empty value AUTH return "Error: file with OAuth token hasn't been found.\nUse 'token' command to authenticate and create this file"
	if notExists(auth) {
		fmt.Println("Error: file with OAuth token hasn't been found.\nUse 'token' command to authenticate and create this file")
		os.Exit(1)
	}
}
