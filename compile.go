package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os/exec"
	"time"
)

func compile(codepath, lang string) (compileOutput string, err error) {
	ch := make(chan string)
	compile := func(cmd *exec.Cmd) (err error) {
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Printf("Error: %s\n", err)
			return
		}
		if err = cmd.Start(); err != nil {
			log.Printf("Error: %s\n", err)
			return
		}
		bytes, _ := ioutil.ReadAll(stderr)
		if err := cmd.Wait(); err != nil {
			ch <- (err.Error() + "\n" + string(bytes))
		}
		ch <- string(bytes)
		return
	}

	switch lang {
	case "cpp":
		cmd := exec.Command("g++", codepath)
		go compile(cmd)
	case "c":
		cmd := exec.Command("gcc", codepath)
		go compile(cmd)
	case "java":
		cmd := exec.Command("javac", codepath)
		go compile(cmd)
	}

	select {
	case res := <-ch:
		return res, nil
	case <-time.After(time.Second * 15):
		return "", errors.New("Compile Time Out")
	}
	return "", nil
}
