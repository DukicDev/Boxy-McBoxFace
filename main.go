package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/opencontainers/runtime-spec/specs-go"
)

var destDir = "./boxy-mcboxface/alpine"

func main() {
	cmd := os.Args[1]

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
	os.RemoveAll(destDir)
}

func child() {
	image := os.Args[2]

	if image != "alpine" {
		fmt.Println("Sorry, i only know alpine")
		return
	}
	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		panic(err)
	}
	imageCmd, err := ExtractImage("alpine")
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(imageCmd[0])
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cg()

	err = syscall.Sethostname([]byte("Boxy-McBoxFace"))
	if err != nil {
		panic(err)
	}

	err = syscall.Mount("proc", destDir+"/proc", "proc", 0, "")
	if err != nil {
		panic(err)
	}

	defer syscall.Unmount("proc", 0)

	err = syscall.Chroot(destDir)
	if err != nil {
		panic(err)
	}

	err = os.Chdir("/home")
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
