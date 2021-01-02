package service

import (
	"fmt"
	"log"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/nas"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func initServer(gnb *context.GNBContext) error {

	// initiated GNB server with unix sockets.
	log.Println("Starting Unix server")
	ln, err := net.Listen("unix", "/tmp/gnb.sock")
	if err != nil {
		fmt.Errorf("Listen error: ", err)
	}

	gnb.SetListener(ln)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(ln net.Listener, c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		ln.Close()
		os.Exit(0)
	}(ln, sigc)

	go gnbListen(gnb)

	return nil
}

func gnbListen(gnb *context.GNBContext) {

	ln := gnb.GetListener()

	for {

		fd, err := ln.Accept()
		if err != nil {
			log.Fatal("Accept error: ", err)
		}

		// TODO this region of the code may induces race condition.

		// new instance GNB UE context
		// store UE in UE Pool.
		// store UE connection.
		// select AMF and get sctp association.
		ue := gnb.NewGnBUe(fd)

		// accept and handle connection.
		go processingConn(ue, gnb)
	}

}

func processingConn(ue *context.GNBUe, gnb *context.GNBContext) {

	buf := make([]byte, 65535)
	conn := ue.GetUnixSocket()

	for {

		n, err := conn.Read(buf[:])
		if err != nil {
			return
		}

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// send to dispatch.
		go nas.Dispatch(ue, forwardData, gnb)
	}
}
