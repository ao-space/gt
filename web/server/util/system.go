package util

import (
	"errors"
	"github.com/isrc-cas/gt/web/server/model/request"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

// SendSignal will create a new process to restart the services
func SendSignal(signal string) (err error) {
	execPath, err := os.Executable()
	if err != nil {
		return
	}
	var cmd *exec.Cmd
	switch signal {
	case "reload":
		cmd = exec.Command(execPath, "-s", "reload")
	case "restart":
		cmd = exec.Command(execPath, "-s", "restart")
	case "stop":
		cmd = exec.Command(execPath, "-s", "stop")
	case "kill":
		cmd = exec.Command(execPath, "-s", "kill")
	default:
		err = errors.New("unknown signal")
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		return
	}
	err = cmd.Process.Release()
	return
}

func GetServerInfo() (server *request.Server, err error) {
	var s request.Server
	s.Os = initOS()
	if s.Cpu, err = initCPU(); err != nil {
		return nil, err
	}
	if s.Ram, err = initRAM(); err != nil {
		return nil, err
	}
	if s.Disk, err = initDisk(); err != nil {
		return nil, err
	}
	return &s, nil
}

func initOS() (o request.Os) {
	o.GOOS = runtime.GOOS
	o.NumCPU = runtime.NumCPU()
	o.Compiler = runtime.Compiler
	o.GoVersion = runtime.Version()
	o.NumGoroutine = runtime.NumGoroutine()
	return o
}

func initCPU() (c request.Cpu, err error) {
	if cores, err := cpu.Counts(false); err != nil {
		return c, err
	} else {
		c.Cores = cores
	}
	if cpus, err := cpu.Percent(time.Duration(200)*time.Millisecond, true); err != nil {
		return c, err
	} else {
		c.Cpus = cpus
	}
	return c, nil
}

func initRAM() (r request.Ram, err error) {
	if u, err := mem.VirtualMemory(); err != nil {
		return r, err
	} else {
		r.UsedMB = int(u.Used) / MB
		r.TotalMB = int(u.Total) / MB
		r.UsedPercent = int(u.UsedPercent)
	}
	return r, nil
}

func initDisk() (d request.Disk, err error) {
	if u, err := disk.Usage("/"); err != nil {
		return d, err
	} else {
		d.UsedMB = int(u.Used) / MB
		d.UsedGB = int(u.Used) / GB
		d.TotalMB = int(u.Total) / MB
		d.TotalGB = int(u.Total) / GB
		d.UsedPercent = int(u.UsedPercent)
	}
	return d, nil
}
