package main

import (
	"log"

	"github.com/bitly/go-simplejson"
)

func dealMessage(message []byte, datapath string) {
	js, _ := simplejson.NewJson(message)
	statusId := js.Get("statusid").MustInt()
	testdataId := js.Get("testdataid").MustArray()
	code := js.Get("code").MustString()
	timeLimit := js.Get("timelimit").MustInt()
	memoryLimit := js.Get("memorylimit").MustInt()
	lang := js.Get("lang").MustString()

	log.Println(statusId)
	log.Println(testdataId)
	log.Println(code)
	log.Println(timeLimit)
	log.Println(memoryLimit)
	log.Println(lang)
}
