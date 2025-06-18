package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/myczh-1/lazy-ctrl-agent/internal/handler"
)

func main() {
	err := handler.LoadCommands("config/commands.json")
	if err != nil {
		log.Fatal("加载 commands.json 失败：", err)
	}

	http.HandleFunc("/execute", handler.HandleExecute)

	fmt.Println("Agent 正在运行：localhost:7070")
	log.Fatal(http.ListenAndServe(":7070", nil))
}

