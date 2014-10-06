package main

import (
	"fmt"
	"log"
	"path"

	"encoding/json"
	"os"

	//"github.com/bitly/go-simplejson"
)

type request struct {
	StatusID    int    `json:"statusid"`
	Code        string `json:"code"`
	TimeLimit   int    `json:"timelimit"`
	MemoryLimit int    `json:"memorylimit"`
	Lang        string `json:"lang"`
	TestDataId  int    `json:"testdataid"`
	CaseCount   int    `json:"casecount"`
}

type result map[string]string

type detail struct {
	Result []result `json:"result"`
}

var (
	compiledProgram string
)

func dealMessage(message []byte, datapath, tmppath string) {
	r := parseRequest(message)

	fileName := saveCodeFile(r.Code, r.Lang, tmppath)

	compileMessage, err := compile(path.Join(tmppath, fileName), r.Lang)
	if err == nil && compileMessage != "" {
		// TODO(maemual): CE

	} else if err != nil {
		log.Fatalf("%s: %s", "Comple Error", err)
		panic(fmt.Sprintf("%s: %s", "Compile Error", err))
	}
	suffifx := ""
	if r.Lang == "java" {
		suffifx += ".class"
	}
	currentpath, _ := os.Getwd()
	compiledProgram = path.Join(currentpath, "Main"+suffifx)
	os.Chown(compiledProgram, uid, gid)
	judge(r)
	saveResult()
}

func parseRequest(message []byte) (r request) {
	json.Unmarshal(message, &r)
	return
}

func saveCodeFile(code, lang, tmppath string) (fileName string) {
	if lang != "java" {
		fileName = "code." + lang
	} else {
		fileName = "Main.java"
	}
	filePath := path.Join(tmppath, fileName)
	sourceCode, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("%s: %s", "Can not create file", err)
		panic(fmt.Sprintf("%s: %s", "Can not create file", err))
	}
	sourceCode.WriteString(code)
	sourceCode.Close()
	return
}

func addResult(caseNumber int, result string, time, memory int, extraMessage string) {
	fmt.Println(caseNumber, result, time, memory, extraMessage)
}

func saveResult() {

}
