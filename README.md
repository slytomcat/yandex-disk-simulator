# yandex-disk-simulator
yandex-disk-simulator is a yandex-disk utility simulator for integration tests (Linux).

    Usage:
    	yandex-disk-similator <cmd>
    Commands:
    	start	starts the daemon and begin starting events simulation
    	stop	stops the daemon
    	status	get the daemon status
    	sync	begin the synchronisation events simulation 
    	help	get this message
    Simulator commands:
    	prepare prepare the simulation environment. It creates the cofig and token files in 
    		$Sim_ConfDir path and syncronized directory as $Sim_SyncDir.
    		Environment variables Sim_ConfDir and Sim_SyncDir should be set in advance.
    	daemon	start as a daemon (don't use it)
    Environment variables:
    	Sim_SyncDir	can be used to set synchronized directory path (default: ~/Yandex.Disk)
    	Sim_ConfDir	can be used to set configuration directory path (default: ~/.config/yandex-disk)

Note:

At the moment, this simulator doesn't handle any additional command or option of original yandex-disk utility except the commands listed above.

Use 
    yandex-disk-similator prepare
to initialize the simulation environment.

To use it as yandex-disk simulator consider renaming the **yandex-disk-similator** to **yandex-disk** and put it in the PATH before the original yandex-disk (if it is installed).
