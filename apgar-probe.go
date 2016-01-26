//
// The MIT License (MIT)
//
// Copyright (c) 2016 Daqri, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//
// Author: Joe Block <joe.block@daqri.com>

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
)

// Setup globals
var debug int
var documentRoot string
var healthCheckName string
var healthCheckTree string
var healthy bool
var testsGroup sync.WaitGroup

func isDir(path string) bool {
	s, err := os.Stat(path)
	if err == nil {
		return s.IsDir()
	}
	return false
}

func isExecutable(path string) bool {
	s, err := os.Stat(path)
	if err == nil {
		return (s.Mode().Perm() & 0111) != 0
	}
	return false
}

func printError(err error) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
}

func errorCheck(e error) {
	if e != nil {
		panic(e)
	}
}

func runHealthCheck(wg *sync.WaitGroup, path string) {
	fmt.Printf("* Running %s...\n", path)
	cmd := exec.Command(path)
	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		printError(err)
		// Did the command fail because of an unsuccessful exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			fmt.Printf("%v: %d\n", path, waitStatus.ExitStatus())
			healthy = false
			write_health_status(documentRoot, false)
		}
	} else {
		// Command was successful
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		fmt.Printf("%v: OK\n", path)
	}
	wg.Done()
}

func visit(path string, f os.FileInfo, err error) error {
	if debug > 0 {
		fmt.Printf("\nchecking %s\n", path)
	}
	if filepath.Base(path) == healthCheckName {
		testsGroup.Add(1)
		go runHealthCheck(&testsGroup, path)
	}
	return nil
}

func write_health_status(path string, healthy bool) error {
	var statusString string

	healthFilePath := fmt.Sprintf("%v/status.txt", path)
	if debug >= 0 {
		fmt.Printf("status path: %s\n", healthFilePath)
		fmt.Printf("Health: %v\n", healthy)
		fmt.Println("Writing status\n")
	}
	healthFile, err := os.Create(healthFilePath)
	errorCheck(err)
	if healthy {
		statusString = "OK\n"
	} else {
		statusString = "UNHEALTHY\n"
	}
	_, err = healthFile.WriteString(statusString)
	errorCheck(err)
	healthFile.Sync()
	defer healthFile.Close()
	return err
}

func debugDump() {
	fmt.Println("Apgar Settings:")
	fmt.Println("***************")
	fmt.Printf("documentRoot: %v\n", documentRoot)
	fmt.Printf("healthCheckName: %v\n", healthCheckName)
	fmt.Printf("healthCheckTree: %v\n", healthCheckTree)
	fmt.Println()
}

func main() {
	flag.IntVar(&debug, "debug", 0, "Debug level")
	flag.StringVar(&documentRoot, "document-root", "/var/lib/apgar", "Document root")
	flag.StringVar(&healthCheckName, "healthcheck-name", "healthCheck", "health check script name")
	flag.StringVar(&healthCheckTree, "healthcheck-tree", "/var/lib/apgar", "Directory tree to search for health checks")

	flag.Parse()

	if debug > 0 {
		debugDump()
	}
	healthy = true
	err := filepath.Walk(healthCheckTree, visit)
	testsGroup.Wait()
	if debug > 40 {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
	fmt.Printf("Health Status: %v\n", healthy)
	write_health_status(documentRoot, healthy)
	if healthy == true {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
