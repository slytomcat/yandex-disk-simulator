# yandex-disk-simulator
[![CircleCI](https://circleci.com/gh/slytomcat/yandex-disk-simulator.svg?style=svg)](https://circleci.com/gh/slytomcat/yandex-disk-simulator)

[![DeepSource](https://static.deepsource.io/deepsource-badge-light.svg)](https://deepsource.io/gh/slytomcat/yandex-disk-simulator/?ref=repository-badge)

**yandex-disk-simulator** is a *yandex-disk* utility simulator for integration tests (Linux).

You can get compiled binaries (ELF) for linux/amd64 and linux/386 platforms from [last release](https://github.com/slytomcat/yandex-disk-simulator/releases/latest) or bild it yourself.

There is no additional libraries requirements to run the simulator.

**Buiding requerements:** 
 - go v.1.13 and higher

**Buiding:**

    go get -d github.com/slytomcat/yandex-disk-simulator
    go build yandex-disk-simulator

**Usage**

Help message:

    Usage:
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
    	Sim_ConfDir	can be used to set configuration directory path (default: ~/.config/yandex-disk)

**NOTE**

This simulator doesn't handle any additional command or option of original yandex-disk utility except the commands listed above.

In order to setup simulator and it's environment do the folloving steps:
1. set *Sim_SyncDir* (synchronized directory path) and *Sim_ConfDir* (configuration directory path) environment variables.
2. run

    yandex-disk-similator setup

**IMPORTANT**

If *Sim_SyncDir* and *Sim_ConfDir* are not set then *"$HOME/Yandex.Disk"* is used as syncronizition folder and *"$HOME/.config/yandex-disk"* is used as configuration folder. Those are same paths as original *yandex-disk* uses. And this can broke the original *yandex-disk* configuration.

**GOOD IDEA**

To use it as yandex-disk simulator consider renaming the *yandex-disk-similator* to *yandex-disk* and put it in the PATH before the original yandex-disk (if it is installed).
