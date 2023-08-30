package service

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// StartClient executePath: the path of gt service(client/server) executable file
//
// configPath: the path of gt client config file
//
// StartClient starts the gt client
// TODO: autoFind path of gt client
func StartService(executablePath, configPath string) (pid int, err error) {
	cmd := exec.Command(executablePath, "--config", configPath)
	//make sure the gt client process is not the child process of the current process
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	outfile, err := os.Create("/home/seb/Desktop/web_log/gt_service.log")
	if err != nil {
		fmt.Println("Failed to create gt service log file", err)
		return 0, err
	}

	defer outfile.Close()
	cmd.Stdout = outfile
	cmd.Stderr = outfile

	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start gt service", err)
		return 0, err
	}

	fmt.Println("gt service started successfully: ", cmd.Process.Pid)

	go func() {
		state, err := cmd.Process.Wait()
		if err != nil {
			fmt.Println("gt service wait failed: ", err)
		}
		if ws, ok := state.Sys().(syscall.WaitStatus); ok {
			fmt.Println("gt service exit with code: ", ws.ExitStatus())
			fmt.Println("gt service exit with signal: ", ws.Signal())
			fmt.Println("gt service exit with core dump: ", ws.CoreDump())
		}
	}()
	return cmd.Process.Pid, err
}

func SendInterruptSignal(pid int) error {
	//Find the process by pid
	process, err := os.FindProcess(pid)
	if err != nil {
		err = fmt.Errorf("failed to find process by pid: %d,cause %s", pid, err.Error())
		return err
	}
	//Send interrupt signal to the process
	err = process.Signal(os.Interrupt)
	if err != nil {
		fmt.Println("Failed to send interrupt signal to :", pid, err)
		return err
	}
	fmt.Println("Sent interrupt signal to :", pid)
	return nil
}
