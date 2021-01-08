// Copyright (c) 2020 Red Hat, Inc.

// +build testrunmain

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestRunMain(t *testing.T) {
	go main()
	fmt.Print("Waiting Signal")
	// hacks for handling signals
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			fmt.Printf("Signal Interupt: %s", sig.String())
			return
		case syscall.SIGTERM:
			//handle SIGTERM
			fmt.Printf("Signal SIGTERM: %s", sig.String())
			fmt.Print("Sleep 30 sec")
			time.Sleep(30 * time.Second)
			return
		}
	}()
}
