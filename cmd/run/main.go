package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/docker/docker/pkg/reexec"
)

func init() {
	reexec.Register("nsInitialisation", nsInitialisation)
	if reexec.Init() {
		os.Exit(0)
	}
}

func nsInitialisation() {
	newrootPath := os.Args[1]

	if err := mountProc(newrootPath); err != nil {
		fmt.Printf("Error mounting /proc - %s\n", err)
		os.Exit(1)
	}

	if err := pivotRoot(newrootPath); err != nil {
		fmt.Printf("Error running pivot_root - %s\n", err)
		os.Exit(1)
	}

	if err := syscall.Sethostname([]byte("ns-process")); err != nil {
		fmt.Printf("Error setting hostname - %s\n", err)
		os.Exit(1)
	}

	//if err := waitForNetwork(); err != nil {
	//	fmt.Printf("Error waiting for network - %s\n", err)
	//	os.Exit(1)
	//}

	nsRun()
}

func nsRun() {
	fmt.Println("=====================")
	fmt.Printf("running: %+v \n", os.Args[2:])
	fmt.Println("=====================")
	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// cmd.Env = []string{"PS1=-[ns-process]- # "}

	if err := cmd.Run(); err != nil {
		cmdStr := strings.Join(os.Args[2:], " ")
		fmt.Printf("Error running the %s command: %+v\n", cmdStr, err)
		os.Exit(1)
	}
}

func mountProc(newroot string) error {
	source := "proc"
	target := filepath.Join(newroot, "/proc")
	fstype := "proc"
	flags := 0
	data := ""

	os.MkdirAll(target, 0755)
	if err := syscall.Mount(source, target, fstype, uintptr(flags), data); err != nil {
		return err
	}

	return nil
}

func pivotRoot(newroot string) error {
	putold := filepath.Join(newroot, "/.pivot_root")

	// bind mount newroot to itself - this is a slight hack needed to satisfy the
	// pivot_root requirement that newroot and putold must not be on the same
	// filesystem as the current root
	if err := syscall.Mount(newroot, newroot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return err
	}

	// create putold directory
	if err := os.MkdirAll(putold, 0700); err != nil {
		return err
	}

	// call pivot_root
	if err := syscall.PivotRoot(newroot, putold); err != nil {
		return err
	}

	// ensure current working directory is set to new root
	if err := os.Chdir("/"); err != nil {
		return err
	}

	// umount putold, which now lives at /.pivot_root
	putold = "/.pivot_root"
	if err := syscall.Unmount(putold, syscall.MNT_DETACH); err != nil {
		return err
	}

	// remove putold
	if err := os.RemoveAll(putold); err != nil {
		return err
	}

	return nil
}

func exitIfRootfsNotFound(rootfsPath string) {
	if _, err := os.Stat(rootfsPath); os.IsNotExist(err) {
		usefulErrorMsg := fmt.Sprintf(`
"%s" does not exist.
Please create this directory and create a suitable root filesystem inside it.
ROOTFS=%s make rootfs
`, rootfsPath, rootfsPath)

		fmt.Println(usefulErrorMsg)
		os.Exit(1)
	}
}

func main() {
	var rootfsPath string
	// var rootfsPath, netsetgoPath string
	flag.StringVar(&rootfsPath, "rootfs", "/tmp/rootfs", "Path to the root filesystem to use")
	// flag.StringVar(&netsetgoPath, "netsetgo", "/usr/local/bin/netsetgo", "Path to the netsetgo binary")
	flag.Parse()

	exitIfRootfsNotFound(rootfsPath)
	// exitIfNetsetgoNotFound(netsetgoPath)

	args := []string{"nsInitialisation", rootfsPath}
	args = append(args, os.Args[1:]...)
	cmd := reexec.Command(args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting the reexec.Command - %s\n", err)
		os.Exit(1)
	}

	// run netsetgo using default args
	// note that netsetgo must be owned by root with the setuid bit set
	pid := fmt.Sprintf("%d", cmd.Process.Pid)
	fmt.Printf("pid: %+v\n", pid)
	//netsetgoCmd := exec.Command(netsetgoPath, "-pid", pid)
	//if err := netsetgoCmd.Run(); err != nil {
	//	fmt.Printf("Error running netsetgo - %s\n", err)
	//	os.Exit(1)
	//}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for the reexec.Command - %s\n", err)
		os.Exit(1)
	}
}
