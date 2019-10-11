package main

import (
	"io"
	"log"
	"sync"
	"time"
)

const (
	msgIdle   = "Synchronization core status: idle\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tTotal: 43.50 GB\n\tUsed: 2.89 GB\n\tAvailable: 40.61 GB\n\tMax file size: 50 GB\n\tTrash size: 0 B\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n"
	startTime = 500
)

// event - the stucture for change event
type event struct {
	msg      string        // status message
	duration time.Duration // event duration
	logm     string        // cli.log message or no message if it is ""
}

// Simulator - the itnterface to simulator engine
type Simulator struct {
	message     string
	msgLock     sync.RWMutex
	symLock     sync.Mutex
	simulations map[string][]event
	logger      io.Writer
}

// NewSimilator - constructor of new Simulator
func NewSimilator(logger io.Writer) *Simulator {
	return &Simulator{
		logger:  logger,
		message: " ",
		simulations: map[string][]event{
			"Start": []event{
				event{" ",
					time.Duration(1200) * time.Millisecond,
					""},
				event{"Synchronization core status: paused\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tThe quota has not been received yet.\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
					time.Duration(250) * time.Millisecond,
					"Start simulation 1"},
				event{"Synchronization core status: index\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tThe quota has not been received yet.\n\n",
					time.Duration(600) * time.Millisecond,
					"Start simulation 2"},
				event{"Synchronization core status: busy\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tThe quota has not been received yet.\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
					time.Duration(100) * time.Millisecond,
					"Start simulation 3"},
				event{"Synchronization core status: index\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tThe quota has not been received yet.\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
					time.Duration(2200) * time.Millisecond,
					"Start simulation 4"},
			},
			"Synchronization": []event{
				event{"Synchronization core status: index\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tTotal: 43.50 GB\n\tUsed: 2.89 GB\n\tAvailable: 40.61 GB\n\tMax file size: 50 GB\n\tTrash size: 0 B\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
					time.Duration(900) * time.Millisecond,
					"Synchronization simulation started"},
				event{"Sync progress: 0 MB/ 139.38 MB (0 %)\nSynchronization core status: busy\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tTotal: 43.50 GB\n\tUsed: 2.89 GB\n\tAvailable: 40.61 GB\n\tMax file size: 50 GB\n\tTrash size: 0 B\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
					time.Duration(100) * time.Millisecond,
					"Synchronization simulation 1"},
				event{"Sync progress: 65.34 MB/ 139.38 MB (46 %)\nSynchronization core status: busy\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tTotal: 43.50 GB\n\tUsed: 2.89 GB\n\tAvailable: 40.61 GB\n\tMax file size: 50 GB\n\tTrash size: 0 B\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
					time.Duration(1500) * time.Millisecond,
					"Synchronization simulation 2"},
				event{"Sync progress: 139.38 MB/ 139.38 MB (100 %)\nSynchronization core status: index\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tTotal: 43.50 GB\n\tUsed: 2.89 GB\n\tAvailable: 40.61 GB\n\tMax file size: 50 GB\n\tTrash size: 0 B\n\nLast synchronized items:\n\tfile: 'NewFile'\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\n",
					time.Duration(500) * time.Millisecond,
					"Synchronization simulation 3"},
			},
			"Error": []event{
				event{"Synchronization core status: error\nError: access error\nPath: 'downloads/test1'\nPath to Yandex.Disk directory: '/home/stc/Yandex.Disk'\n\tTotal: 43.50 GB\n\tUsed: 2.88 GB\n\tAvailable: 40.62 GB\n\tMax file size: 50 GB\n\tTrash size: 654.48 MB\n\nLast synchronized items:\n\tfile: 'File.ods'\n\tfile: 'downloads/file.deb'\n\tfile: 'downloads/setup'\n\tfile: 'download'\n\tfile: 'down'\n\tfile: 'do'\n\tfile: 'd'\n\tfile: 'o'\n\tfile: 'w'\n\tfile: 'n'\n\n",
					time.Duration(500) * time.Millisecond,
					"Error simulation 1"},
			},
		},
	}
}

// setMsg is thread safe message update
func (s *Simulator) setMsg(m string) {
	s.msgLock.Lock()
	s.message = m
	s.msgLock.Unlock()
}

// Simulate starts the set of events simulation
// The set must be one of: "Start", "Synchronization" or "Error"
func (s *Simulator) Simulate(set string) {
	sequence, ok := s.simulations[set]
	if !ok {
		return
	}
	// run simulation in separate goroutine
	go func(seq []event, l io.Writer) {
		s.symLock.Lock()
		defer s.symLock.Unlock()
		for _, e := range seq {
			s.setMsg(e.msg)
			if e.logm != "" {
				if _, err := l.Write([]byte(e.logm + "\n")); err != nil {
					panic(err)
				}
				log.Println(e.logm)
			}
			time.Sleep(e.duration)
		}
		// at the end of simulation set the idle/synchronised status message
		s.setMsg(msgIdle)
		if _, err := l.Write([]byte(set + " simulation finished\n")); err != nil {
			panic(err)
		}
		log.Println(set + " simulation finished")
	}(sequence, s.logger)
}

// GetMessage returns the current status message
func (s *Simulator) GetMessage() string {
	s.msgLock.RLock()
	defer s.msgLock.RUnlock()
	return s.message
}
