package packer

import (
	"log"
	"time"
)

/*
https://yourbasic.org/golang/measure-execution-time/

Usage:

func foo() {
	defer duration(track("foo"))
	// Code to measure
}
*/

func track(msg string) (string, time.Time) {
	return msg, time.Now()
}

func duration(msg string, start time.Time) {
	log.Printf("%v: %v\n", msg, time.Since(start))
}
