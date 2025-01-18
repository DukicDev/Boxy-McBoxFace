package main

import (
	"os"
	"os/exec"
	"syscall"
)

func main() {
	cmd := os.Args[1]

	switch cmd {
	case "run":
		run()

	case "child":
		child()
	}
}

func run() {
	cmd := exec.Command("/proc/self/exe", "child")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func child() {
	cmd := exec.Command("/bin/bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := syscall.Sethostname([]byte("Boxy-McBoxFace"))
	if err != nil {
		panic(err)
	}

	err = os.Chdir("/")
	if err != nil {
		panic(err)
	}

	err = syscall.Mount("proc", "proc", "proc", 0, "")
	if err != nil {
		panic(err)
	}

	defer syscall.Unmount("proc", 0)

	err = cmd.Run()
	if err != nil {
		panic(err)
	}

}
