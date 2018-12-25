# yandex-disk-simulator
yandex-disk-simulator is a yandex-disk utility simulator for integration tests (Linux).

Usage:
	yandex-disk-similator <cmd>
Commands:
	start	starts the daemon and begin starting events simulation
	stop	stops the daemon
	status	get the daemon status
	sync	begin the synchronisation events simulation 
  help  get this message
Simulator commands:
	daemon	start as a daemon (don't use it)
Environment variables:
	DEBUG_SyncDir	can be used to set synchronized directory path (default: ~/Yandex.Disk)
	DEBUG_ConfDir	can be used to set configuration directory path (default: ~/.config/yandex-disk)

