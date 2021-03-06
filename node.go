package main

import (
	"fmt"
	"log"
	"path"

	"strconv"
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
	HighScore	int	   `json:"highest_score"`
}

var (
	compiledProgram string
	r               request
)

func dealMessage(message []byte) {
	r = parseRequest(message)

	fileName := saveCodeFile(r.Code, r.Lang)

	compileMessage, err := compile(path.Join(config.TempPath, fileName), r.Lang)
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

func saveCodeFile(code, lang string) (fileName string) {
	if lang != "java" {
		fileName = "code." + lang
	} else {
		fileName = "Main.java"
	}
	filePath := path.Join(config.TempPath, fileName)
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
		rconn := pool.Get()
		defer rconn.Close()
		hashtable_name := fmt.Sprintf("contestscore:%d:%d", r.ContestID, r.UserID)
		if r.HighScore == 1 {
			_score, _ := rconn.Do("HGET", hashtable_name, r.ProblemID)
			var score int
			if _score != nil {
				_score := _score.([]uint8)
				str := make([]byte, len(_score))
			    for i, v := range _score {
			        str[i] = byte(v)
			    }
			    score, _ = strconv.Atoi(string(str))
			}
			if _score == nil || score < total {
				rconn.Do("HSET", hashtable_name, r.ProblemID, total)
			}
		} else {
			rconn.Do("HSET", hashtable_name, r.ProblemID, total)
		}
	}
}
