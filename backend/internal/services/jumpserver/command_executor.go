package jumpserver

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

// ExecuteSSHCommand 执行单条 SSH 命令
func ExecuteSSHCommand(ctx context.Context, host string, port int, username, password, command string) BatchExecHostResult {
	config := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return BatchExecHostResult{
			Success: false,
			Error:   fmt.Sprintf("SSH连接失败: %v", err),
		}
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return BatchExecHostResult{
			Success: false,
			Error:   fmt.Sprintf("创建会话失败: %v", err),
		}
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// 使用 context 控制超时
	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case <-ctx.Done():
		return BatchExecHostResult{
			Success: false,
			Error:   "命令执行超时",
			Output:  stdout.String(),
		}
	case err := <-done:
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				exitCode = exitErr.ExitStatus()
			} else {
				exitCode = -1
			}
		}
		output := stdout.String()
		if stderr.Len() > 0 {
			output += "\n[stderr]\n" + stderr.String()
		}
		return BatchExecHostResult{
			Success:  err == nil || exitCode == 0,
			Output:   output,
			ExitCode: exitCode,
		}
	}
}