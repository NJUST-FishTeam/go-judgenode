package main

import (
	"fmt"
	"log"
	"path"

	"os"

	"github.com/bitly/go-simplejson"
)

func dealMessage(message []byte, datapath, tmppath string) {
	js, _ := simplejson.NewJson(message)
	statusId := js.Get("statusid").MustInt()
	testdataId := js.Get("testdataid").MustArray()
	code := js.Get("code").MustString()
	timeLimit := js.Get("timelimit").MustInt()
	memoryLimit := js.Get("memorylimit").MustInt()
	lang := js.Get("lang").MustString()

	fileName := saveCodeFile(code, lang, tmppath)

	compileMessage, err := compile(path.Join(tmppath, fileName), lang)
	if err == nil && compileMessage == "" {

	} else if err == nil && compileMessage != "" {
		// TODO(maemual): CE

	} else {
		log.Fatalf("%s: %s", "Comple Error", err)
		panic(fmt.Sprintf("%s: %s", "Compile Error", err))
	}

	log.Println(statusId)
	log.Println(testdataId)
	log.Println(code)
	log.Println(timeLimit)
	log.Println(memoryLimit)
	log.Println(lang)
}

func saveCodeFile(code, lang, tmppath string) string {
	fileName := ""
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
	return fileName
}
