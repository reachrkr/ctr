/*
	Based on

	talk from Liz Rice at UK Docker Conf 2016
	https://www.youtube.com/watch?v=HPuvDm8IC-4

	AND

	Article by Julian Friedman
	https://www.infoq.com/articles/build-a-container-golang

	Namespacing - what it sees
		UNIX Timesharing System
		Process IDs
		File System (mount points)
		Users
		IPC
		Networking

	Control Groups - what resources can use
		CPU
		Memory
		Disk I/O
		Network
		Device permissions (/dev)
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"net"
	"time"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("what?")
	}
}

func run() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET ,
	}

	fmt.Println("run()")
	if err := cmd.Start(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

fmt.Printf("%d",cmd.Process.Pid)
pid := fmt.Sprintf("%d", cmd.Process.Pid)
netsetgoCmd := exec.Command("/usr/local/bin/netsetgo", "-pid", pid)
if err := netsetgoCmd.Run(); err != nil {
        fmt.Printf("Error running netsetgo - %s\n", err)
        os.Exit(1)
}

if err := cmd.Wait(); err != nil {
        fmt.Printf("Error waiting for reexec.Command - %s\n", err)
        os.Exit(1)
}


}

func waitForNetwork() error {
	maxWait := time.Second * 3
	checkInterval := time.Second
	timeStarted := time.Now()

	for {
		interfaces, err := net.Interfaces()
		if err != nil {
			return err
		}
fmt.Printf("inside wait network")
		// pretty basic check ...
		// > 1 as a lo device will already exist
		if len(interfaces) > 1 {
			return nil
		}

		if time.Since(timeStarted) > maxWait {
			return fmt.Errorf("Timeout after %s waiting for network", maxWait)
		}

		time.Sleep(checkInterval)
	}
}
func child() {
	fmt.Printf("running %v as pid %d\n", os.Args[2:], os.Getpid())

	must(syscall.Chroot("ubuntu"))
	must(os.Chdir("/"))
//	fmt.Printf("changed root")
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
//	fmt.Printf("mounted root")
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// running command
	syscall.Sethostname([]byte("demo"))




waitForNetwork()

	fmt.Println("running command")
	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
