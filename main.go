package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func main() {
	fmt.Println(os.Args)
	cmd := os.Args[1]
	image := os.Args[2]

	if image != "alpine" {
		fmt.Println("Sorry, i only know alpine")
		return
	}

	switch cmd {
	case "run":
		run()

	case "child":
		child()

	default:
		panic("No command -> No Boxy-McBoxFace!!1111!!!!")
	}
}

func run() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
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
	cmd := exec.Command(os.Args[3], os.Args[4:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cg()

	err := syscall.Sethostname([]byte("Boxy-McBoxFace"))
	if err != nil {
		panic(err)
	}

	err = syscall.Mount("proc", "./images/alpine/layer/proc", "proc", 0, "")
	if err != nil {
		panic(err)
	}

	defer syscall.Unmount("proc", 0)

	err = syscall.Chroot("./images/alpine/layer/")
	if err != nil {
		panic(err)
	}

	err = os.Chdir("./images/alpine/layer/home/")
	if err != nil {
		panic(err)
	}

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
