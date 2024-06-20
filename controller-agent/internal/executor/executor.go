package executor

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
)

func RunCommand(path string) (string, error) {
	var cmd *exec.Cmd
	
	if runtime.GOOS == "windows" {
		// 如果是PowerShell命令，直接调用PowerShell
		if strings.HasPrefix(path, "powershell") {
			// 解析PowerShell命令参数
			parts := strings.SplitN(path, " ", 3)
			if len(parts) >= 3 && parts[1] == "-c" {
				// 去掉外层引号
				script := strings.Trim(parts[2], "\"")
				cmd = exec.Command("powershell", "-Command", script)
			} else {
				cmd = exec.Command("cmd", "/C", path)
			}
		} else {
			cmd = exec.Command("cmd", "/C", path)
		}
	} else {
		cmd = exec.Command("sh", "-c", path)
	}
	
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func RunCommandWithContext(ctx context.Context, path string) (string, error) {
	var cmd *exec.Cmd
	
	if runtime.GOOS == "windows" {
		// 如果是PowerShell命令，直接调用PowerShell
		if strings.HasPrefix(path, "powershell") {
			// 解析PowerShell命令参数
			parts := strings.SplitN(path, " ", 3)
			if len(parts) >= 3 && parts[1] == "-c" {
				// 去掉外层引号
				script := strings.Trim(parts[2], "\"")
				cmd = exec.CommandContext(ctx, "powershell", "-Command", script)
			} else {
				cmd = exec.CommandContext(ctx, "cmd", "/C", path)
			}
		} else {
			cmd = exec.CommandContext(ctx, "cmd", "/C", path)
		}
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", path)
	}
	
	output, err := cmd.CombinedOutput()
	return string(output), err
}

