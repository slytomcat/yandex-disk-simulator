package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	socketPath = "/tmp/yandexdiskmock.socket"
	startTime  = 1000

	userHome string
	once     sync.Once

	msgLock sync.Mutex

	// start sequence
	message   = " "
	event1    = 1600
	msgStart1 = "Synchronization core status: paused\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			The quota has not been received yet.\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	"
	event2    = 250
	msgStart2 = "Synchronization core status: index\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			The quota has not been received yet.\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	"
	event3    = 600
	msgStart3 = "Synchronization core status: busy\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			The quota has not been received yet.\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	"
	event4    = 100
	// again msgStart2
	event5 = 4600
	msgOk  = "Synchronization core status: idle\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			Total: 43.50 GB\n			Used: 2.89 GB\n			Available: 40.61 GB\n		Max file size: 50 GB\n			Trash size: 0 B\n	\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	"

	// sync sequence
	msgSync1 = "Synchronization core status: index\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			Total: 43.50 GB\n			Used: 2.89 GB\n			Available: 40.61 GB\n		Max file size: 50 GB\n			Trash size: 0 B\n	\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	"
	event11  = 900
	msgSync2 = "Sync progress: 65.34 MB/ 139.38 MB (46 %)\nSynchronization core status: index\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			Total: 43.50 GB\n			Used: 2.89 GB\n			Available: 40.61 GB\n		Max file size: 50 GB\n			Trash size: 0 B\n	\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	"
	event12  = 100
	msgSync3 = "Sync progress: 65.34 MB/ 139.38 MB (46 %)\nSynchronization core status: busy\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			Total: 43.50 GB\n			Used: 2.89 GB\n			Available: 40.61 GB\n		Max file size: 50 GB\n			Trash size: 0 B\n	\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	"
	event13  = 1500
	msgSync4 = "Sync progress: 139.38 MB/ 139.38 MB (100 %)\nSynchronization core status: index\n	Path to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n			Total: 43.50 GB\n			Used: 2.89 GB\n			Available: 40.61 GB\n		Max file size: 50 GB\n			Trash size: 0 B\n	\n	Last synchronized items:\n			file: 'File.ods'\n			file: 'downloads/file.deb'\n			file: 'downloads/setup'\n			file: 'download'\n			file: 'down'\n			file: 'do'\n			file: 'd'\n			file: 'o'\n			file: 'w'\n			file: 'n'\n	\n	\n	"
	event14  = 1500
	// msgOk
)

func simulateSync(l io.Writer) {
	// sync sequence
	setMsg(msgSync1)
	l.Write([]byte("Sync start\n"))
	time.Sleep(time.Duration(event11) * time.Millisecond)
	setMsg(msgSync2)
	l.Write([]byte("Sync pogress1\n"))
	time.Sleep(time.Duration(event12) * time.Millisecond)
	setMsg(msgSync3)
	l.Write([]byte("Sync pogress2\n"))
	time.Sleep(time.Duration(event14) * time.Millisecond)
	setMsg(msgSync4)
	l.Write([]byte("Sync pogress3\n"))
	time.Sleep(time.Duration(event14) * time.Millisecond)
	setMsg(msgOk)
	l.Write([]byte("Sync finished\n"))
}

func simulateStart(l io.Writer) {
	// start sequence
	setMsg(" ")
	time.Sleep(time.Duration(event1) * time.Millisecond)
	setMsg(msgStart1)
	l.Write([]byte("Start simulation 2\n"))
	time.Sleep(time.Duration(event2) * time.Millisecond)
	setMsg(msgStart2)
	l.Write([]byte("Start simulation 3\n"))
	time.Sleep(time.Duration(event3) * time.Millisecond)
	setMsg(msgStart3)
	l.Write([]byte("Start simulation 4\n"))
	time.Sleep(time.Duration(event4) * time.Millisecond)
	setMsg(msgStart2)
	l.Write([]byte("Start simulation 5\n"))
	time.Sleep(time.Duration(event5) * time.Millisecond)
	setMsg(msgOk)
	l.Write([]byte("Start simulation finished\n"))
}

func setMsg(m string) {
	msgLock.Lock()
	message = m
	msgLock.Unlock()
}

func expandHome(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	once.Do(func() {
		usr, err := user.Current()
		if err != nil {
			log.Fatal("Can't get current user profile:", err)
		}
		userHome = usr.HomeDir
	})
	return filepath.Join(userHome, path[1:])
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
	yandexdiskmock <cmd>
Commands:
	tart	starts the daemon and begin starting events simulation
	stop	stops the daemon
	status	get the daemon status
	sync	begin the synchronisation events simulation 
MOK commands:
	daemon	start as a daemon (dont use it)`)
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
	/* execute os.Args[0] with "daemon" as os.Args[1]*/
	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command(path, "daemon")
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func daemon() {
	dlog, err := os.OpenFile(expandHome("/tmp/yandexdiskmock.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
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
	// create ~/Yandex.Disk/.sync/cli.log if it is not exists
	err = os.MkdirAll(expandHome("~/Yandex.Disk/.sync"), 0755)
	if err != nil {
		log.Fatal("~/Yandex.Disk/.sync creation error:", err)
	}
	// open logfile
	logfile, err := os.OpenFile(expandHome("~/Yandex.Disk/.sync/cli.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal("~/Yandex.Disk/.sync/cli.log opening error:", err)
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
	f, err := os.Open(expandHome("~/.config/yandex-disk/config.cfg"))
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
