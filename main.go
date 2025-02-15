package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"syscall"

	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/opencontainers/runtime-spec/specs-go"
)

var baseDir = "/var/lib/boxy-mcboxface/containers/"

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
	image := os.Args[2]

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
		os.RemoveAll(baseDir + image)
		panic(err)
	}

	os.RemoveAll(baseDir + image)
}

func child() {
	image := os.Args[2]
	destDir := baseDir + image

	imageConfig, err := pullImage(image, "latest")
	if err != nil {
		panic(err)
	}
	imageCmd := getImageCmd(imageConfig)
	cmd := exec.Command(imageCmd[0], imageCmd[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cg()

	err = syscall.Sethostname([]byte("Boxy-McBoxFace"))

	if err != nil {
		panic(err)
	}

	err = syscall.Mount("proc", destDir+"/proc", "proc", 0, "")

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		panic(err)
	}

	defer syscall.Unmount("proc", 0)

	err = syscall.Chroot(destDir)

	if err != nil {
		panic(err)
	}

	var workDir string
	if imageConfig.WorkingDir == "" {
		workDir = "/"
	} else {
		workDir = imageConfig.WorkingDir
	}

	err = os.Chdir(workDir)

	if err != nil {
		panic(err)
	}

	for key, env := range imageConfig.getEnvMap() {
		os.Setenv(key, env)
	}

	var cmdPath string
	if len(os.Args) >= 4 {
		imageCmd = os.Args[3:]
	}
	cmdPath, err = exec.LookPath(imageCmd[0])
	if err != nil {
		panic(err)
	}

	cmd.Path = cmdPath
	if len(imageCmd) >= 2 {
		cmd.Args = imageCmd
	}

	fmt.Printf("\nBoxy-McBoxFace running %s with cmd: %s\n\n", image, imageCmd)
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

func getImageCmd(imageConfig Config) []string {
	var cmd []string
	if len(imageConfig.Entrypoint) > 0 {
		cmd = append(imageConfig.Entrypoint, imageConfig.Cmd...)
	} else {
		cmd = imageConfig.Cmd
	}
	return cmd
}
