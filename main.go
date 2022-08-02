package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

//docekr run <cmd> <arguments>
//docker child <cmd> <arguments>

func main() {

	if len(os.Args) <= 1 {
		panic("no enough arguments")
	}

	if len(os.Args) <= 2 {
		panic("please specify the command to run")
	}

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help")
	}
}

func run() {

	fmt.Printf("parent pid: %d\n", os.Getpid())
	//fmt.Printf("Running %v \n", os.Args[2:])

	args := append([]string{"child"}, os.Args[2:]...)

	//First, docker run will execute the /proc/self/exe command to fork the child process
	//The child process will do the real job
	cmd := exec.Command("/proc/self/exe", args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("child pid: %d\n", os.Getpid())
	//fmt.Printf("Running %v \n", os.Args[2:])

	//Set cgroup values
	cg()

	must(syscall.Sethostname([]byte("container")))

	//must(syscall.Mount("ubuntu-fs", "ubuntu-fs", "", syscall.MS_BIND, ""))
	//must(os.MkdirAll("ubuntu-fs/old-ubuntu-fs", 0700))
	//must(syscall.PivotRoot("ubuntu-fs", "ubuntu-fs/old-ubuntu-fs"))
	//must(os.Chdir("/"))

	//set the child process's root file.   use pivot_root
	//You have to download the os rootfs
	must(syscall.Chroot("/root/Youngpig1998/containers-from-scratch/ubuntu-fs"))
	must(os.Chdir("/"))

	//when we execute ps command,it will read /proc directory. However, the ubuntu rootfs
	//dont have /proc, so we have to mount the /proc in host machine.
	must(syscall.Mount("proc", "ubuntu-fs/proc", "proc", 0, ""))
	must(syscall.Mount("thing", "mytemp", "tmpfs", 0, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())

	//must(syscall.Unmount("ubuntu-fs", 0))
	must(syscall.Unmount("proc", 0))
	must(syscall.Unmount("thing", 0))
}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	//  /sys/fs/cgroup/pids/container
	must(os.Mkdir(filepath.Join(pids, "container"), 0755))
	must(ioutil.WriteFile(filepath.Join(pids, "container/pids.max"), []byte("10"), 0700))
	// Removes the new cgroup in place after the container exits
	must(ioutil.WriteFile(filepath.Join(pids, "container/notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "container/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
