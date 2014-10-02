package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os/exec"
	"time"
)

func compile(codepath, lang string) (string, error) {

	cmd := new(exec.Cmd)

	switch lang {
	case "cpp":
		cmd = exec.Command(
			"g++",
			codepath,
			"-o", "Main",
			"-static", "-w",
			"-lm", "-O2",
			"-DONLINE_JUDGE",
		)
	case "c":
		cmd = exec.Command(
			"gcc",
			codepath,
			"-o", "Main",
			"-static", "-w",
			"-lm", "-std=c99",
			"-O2", "-DONLINE_JUDGE",
		)
	case "java":
		cmd = exec.Command(
			"javac",
			codepath,
			"-d", ".",
		)
	}

	ch := make(chan string)
	e := make(chan bool)

	go func(cmd *exec.Cmd) {
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Printf("Error: %s\n", err)
			e <- true
			return
		}
		if err = cmd.Start(); err != nil {
			log.Printf("Error: %s\n", err)
			e <- true
			return
		}
		bytes, _ := ioutil.ReadAll(stderr)
		if err := cmd.Wait(); err != nil {
			log.Printf("Error: %s\n", err)
			e <- true
			return
		}
		ch <- string(bytes)
		return
	}(cmd)

	select {
	case res := <-ch:
		return res, nil
	case <-e:
		return "", errors.New("System Error")
	case <-time.After(time.Second * 15):
		return "", errors.New("Compile Time Out")
	}
	return "", errors.New("System Error")
}
