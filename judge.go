package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type result struct {
	Status    string `json:"status"`
	Runtime   int    `json:"runtime"`
	Runmemory int    `json:"runmemory"`
	Message   string `json:"message"`
}

type detail struct {
	Result []result `json:"result"`
}

func judge(r request) ([]byte, int) {
	var d detail
	var totalScore int
	addResult := func(num int, status string, time, memory int, message string) {
		d.Result = append(d.Result, result{Status: status, Runtime: time, Runmemory: memory, Message: message})
		if status == "Accepted" {
			totalScore += r.CaseScore[num]
		}
	}
	for i := 0; i < r.CaseCount; i++ {
		prepareFiles(r, i)
		err := runProgram(r)
		if err != nil {
			fmt.Println(err.Error())
			addResult(i, "Runtime Error", 0, 0, "")
			cleanFiles()
			continue
		}
		result, time, memory := getResult()
		addResult(i, result, time, memory, "")
		cleanFiles()
	}
	bytes, _ := json.Marshal(d)
	return bytes, totalScore
}

func prepareFiles(r request, num int) {
	if _, err := os.Stat(config.RunPath); err != nil && !os.IsExist(err) {
		os.MkdirAll(config.RunPath, os.ModePerm)
		os.Chown(config.RunPath, uid, gid)
	}
	if r.Lang != "java" {
		cmd := exec.Command("cp", compiledProgram,
			path.Join(config.RunPath, "a.out"))
		cmd.Run()
		os.Chown(path.Join(config.RunPath, "a.out"), uid, gid)
	} else {
		cmd := exec.Command("cp", compiledProgram, config.RunPath)
		cmd.Run()
		os.Chown(path.Join(config.RunPath, compiledProgram), uid, gid)
	}
	cmd := exec.Command("cp",
		path.Join(config.TestDataPath, strconv.Itoa(r.TestDataId), strconv.Itoa(num)+".in"),
		path.Join(config.RunPath, "in.in"))
	cmd.Run()
	os.Chown(path.Join(config.RunPath, "in.in"), uid, gid)
	cmd = exec.Command("cp",
		path.Join(config.TestDataPath, strconv.Itoa(r.TestDataId), strconv.Itoa(num)+".out"),
		path.Join(config.RunPath, "out.out"))
	cmd.Run()
	os.Chown(path.Join(config.RunPath, "out.out"), uid, gid)
	cmd = exec.Command("cp", "Core", config.RunPath)
	cmd.Run()
	os.Chown(path.Join(config.RunPath, "Core"), uid, gid)
}

func runProgram(r request) error {
	var lang int
	switch r.Lang {
	case "cpp":
		lang = 2
	case "c":
		lang = 1
	case "java":
		lang = 3
	default:
		lang = 0
	}
	cmd := exec.Command(
		"sudo",
		"./Core",
		"-t",
		strconv.Itoa(r.TimeLimit),
		"-m",
		strconv.Itoa(r.MemoryLimit),
		"-d",
		config.RunPath,
		"-l",
		strconv.Itoa(lang),
	)
	return cmd.Run()
}

func getResult() (result string, time int, memory int) {
	resultFile := path.Join(config.RunPath, "result.txt")
	f, _ := os.Open(resultFile)
	defer f.Close()
	buff := bufio.NewReader(f)
	line, _ := buff.ReadString('\n')
	result = strings.Trim(line, "\n")
	line, _ = buff.ReadString('\n')
	time, _ = strconv.Atoi(strings.Trim(line, "\n"))
	line, _ = buff.ReadString('\n')
	memory, _ = strconv.Atoi(strings.Trim(line, "\n"))
	return
}

func cleanFiles() {
	os.RemoveAll(config.RunPath)
}
