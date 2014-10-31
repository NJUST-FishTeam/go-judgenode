package main

import (
	"fmt"
	"log"
	"path"

	"github.com/garyburd/redigo/redis"

	"encoding/json"
	"os"
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
	ContestID   int    `json:"contestid"`
	ProblemID   string `json:"problemid"`
	UserID      int    `json:"userid"`
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
		saveResult(true, []byte(compileMessage), 0, r)
		return
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
	saveResult(false, response, total, r)
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

func saveResult(ce bool, data []byte, total int, r request) {
	db.Ping()
	var status string
	if ce {
		status = "Compile Error"
		ceMessage := string(data)
		stmt, _ := db.Prepare("update fishteam_cat.submit_status set status = ?, compilerOutput = ? , score = ? where id = ?;")
		defer stmt.Close()
		stmt.Exec(status, ceMessage, -1, r.StatusID)
	} else {
		stmt, _ := db.Prepare("update fishteam_cat.submit_status set status = ?, detail = ? , score = ? where id = ?")
		defer stmt.Close()
		totalscore := 0
		for _, val := range r.CaseScore {
			totalscore += val
		}
		if totalscore == total {
			status = "Accepted"
		} else {
			status = "Wrong Answer"
		}
		stmt.Exec(status, string(data), total, r.StatusID)
		if r.ContestID != 0 {
			rconn, _ := redis.Dial("tcp", ":6379")
			defer rconn.Close()
			hashtable_name := fmt.Sprintf("contestscore:%d:%d", r.ContestID, r.UserID)
			rconn.Do("HSET", hashtable_name, r.ProblemID, total)
		}
	}
}
