package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

func judge(r request) {
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
}

func prepareFiles(r request, num int) {
	if _, err := os.Stat(runPath); err != nil && !os.IsExist(err) {
		os.MkdirAll(runPath, os.ModePerm)
		os.Chown(runPath, uid, gid)
	}
	if r.Lang != "java" {
		cmd := exec.Command("cp", compiledProgram,
			path.Join(runPath, "a.out"))
		cmd.Run()
		os.Chown(path.Join(runPath, "a.out"), uid, gid)
	} else {
		cmd := exec.Command("cp", compiledProgram, runPath)
		cmd.Run()
		os.Chown(path.Join(runPath, compiledProgram), uid, gid)
	}
	cmd := exec.Command("cp",
		path.Join(testdataPath, strconv.Itoa(r.TestDataId), strconv.Itoa(num)+".in"),
		path.Join(runPath, "in.in"))
	cmd.Run()
	os.Chown(path.Join(runPath, "in.in"), uid, gid)
	cmd = exec.Command("cp",
		path.Join(testdataPath, strconv.Itoa(r.TestDataId), strconv.Itoa(num)+".out"),
		path.Join(runPath, "out.out"))
	cmd.Run()
	os.Chown(path.Join(runPath, "out.out"), uid, gid)
	cmd = exec.Command("cp", "Core", runPath)
	cmd.Run()
	os.Chown(path.Join(runPath, "Core"), uid, gid)
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
		runPath,
		"-l",
		strconv.Itoa(lang),
	)
	return cmd.Run()
}

func getResult() (result string, time int, memory int) {
	resultFile := path.Join(runPath, "result.txt")
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
	os.RemoveAll(runPath)
}
