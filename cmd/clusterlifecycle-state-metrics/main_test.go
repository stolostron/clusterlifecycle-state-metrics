// Copyright (c) 2020 Red Hat, Inc.

// +build testrunmain

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/open-cluster-management/clusterlifecycle-state-metrics/pkg/options"
	"github.com/open-cluster-management/clusterlifecycle-state-metrics/pkg/version"
	"k8s.io/klog"
)

func TestMain(m *testing.M) {
	klog.Info("Create options")
	opts = options.NewOptions()
	opts.AddFlags()
	argsTest := make([]string, 0)
	argsMain := make([]string, 0)
	argsMain = append(argsMain, os.Args[0])
	before := true
	for _, a := range os.Args {
		if a == "--" {
			before = false
			continue
		}
		if before {
			argsTest = append(argsTest, a)
		} else {
			argsMain = append(argsMain, a)
		}
	}
	os.Args = argsMain
	klog.Infof("argsMain=%v\n", argsMain)
	err := opts.Parse()
	if err != nil {
		klog.Fatalf("Error: %s", err)
	}

	klog.Infof("Opts=%v\n", opts)

	if opts.Version {
		fmt.Printf("%#v\n", version.GetVersion())
		os.Exit(0)
	}

	if opts.Help {
		opts.Usage()
		os.Exit(0)
	}
	os.Args = argsTest
	klog.Infof("argsTest=%v\n", argsTest)
	os.Exit(m.Run())
}

func TestStart(t *testing.T) {
	fmt.Print("Start tests")
	go start(opts)
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
			return
		}
	}()
}
