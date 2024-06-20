package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/myczh-1/lazy-ctrl-agent/internal/executor"
)

var commandMap map[string]interface{}

func LoadCommands(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, &commandMap)
}

func GetCommandForPlatform(cmd interface{}) (string, bool) {
	switch v := cmd.(type) {
	case string:
		// 简单字符串命令，直接返回
		return v, true
	case map[string]interface{}:
		// 平台特定命令，根据当前平台选择
		if platformCmd, ok := v[runtime.GOOS]; ok {
			if cmdStr, ok := platformCmd.(string); ok {
				return cmdStr, true
			}
		}
	}
	return "", false
}

func HandleExecute(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	fmt.Printf("Received command ID: %s\n", id)
	fmt.Printf("Available commands: %+v\n", commandMap)
	
	cmd, ok := commandMap[id]
	if !ok {
		http.Error(w, "Command not found", http.StatusNotFound)
		return
	}
	
	path, ok := GetCommandForPlatform(cmd)
	if !ok {
		http.Error(w, fmt.Sprintf("Command not supported on platform: %s", runtime.GOOS), http.StatusNotFound)
		return
	}
	
	output, err := executor.RunCommand(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed: %s\n%s", err, output), 500)
		return
	}
	w.Write([]byte(output))
}

func GetCommand(id string) (interface{}, bool) {
	cmd, ok := commandMap[id]
	return cmd, ok
}

func GetAllCommands() map[string]interface{} {
	if commandMap == nil {
		return make(map[string]interface{})
	}
	return commandMap
}

