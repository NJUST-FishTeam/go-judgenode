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
	CaseScore   []int  `json:"casescore"`
}

var (
	compiledProgram string
	r               request
)

func dealMessage(message []byte, datapath, tmppath string) {
	r = parseRequest(message)

	fileName := saveCodeFile(r.Code, r.Lang, tmppath)

	compileMessage, err := compile(path.Join(tmppath, fileName), r.Lang)
	if err == nil && compileMessage != "" {
		// Compile Error
		saveResult(true, []byte(compileMessage), 0, r.StatusID)
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
	response, total := judge(r)
	saveResult(false, response, total, r.StatusID)
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

func saveResult(ce bool, data []byte, total int, statusid int) {
	db.Ping()
	if ce {
		ceMessage := string(data)
		stmt, _ := db.Prepare("update fishteam_cat.submit_status set compilerOutput = ? , score = ? where id = ?;")
		defer stmt.Close()
		stmt.Exec(ceMessage, -1, statusid)
	} else {
		stmt, _ := db.Prepare("update fishteam_cat.submit_status set detail = ? , score = ? where id = ?")
		defer stmt.Close()
		stmt.Exec(string(data), total, statusid)
	}
}
