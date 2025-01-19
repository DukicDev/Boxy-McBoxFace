package main

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/opencontainers/runtime-spec/specs-go"
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

	cg()

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

func cg() {
	shares := uint64(50)
	memLimit := int64(100 * 1024 * 1024)

	cg, err := cgroup1.New(cgroup1.StaticPath("/boxy-mcboxface"), &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Shares: &shares,
		},
		Memory: &specs.LinuxMemory{
			Limit: &memLimit,
		},
		Pids: &specs.LinuxPids{
			Limit: int64(20),
		},
	})
	if err != nil {
		panic(err)
	}

	pid := os.Getpid()

	if err := cg.Add(cgroup1.Process{Pid: pid}); err != nil {
		panic(err)
	}

}
