package main

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/pion/webrtc/v3"
)

func ChannelCmd(dc *webrtc.DataChannel, pc *webrtc.PeerConnection) {
	outr, outw, err := os.Pipe()
	if err != nil {
		internalErrorRTC(dc, "stdout:", err)
		pc.Close()
		return
	}

	inr, inw, err := os.Pipe()
	if err != nil {
		internalErrorRTC(dc, "stdin:", err)
		pc.Close()
		return
	}

	proc, err := os.StartProcess("/bin/sh", []string{}, &os.ProcAttr{
		Files: []*os.File{inr, outw, outw},
	})
	if err != nil {
		internalErrorRTC(dc, "start:", err)
		outr.Close()
		outw.Close()
		inr.Close()
		inw.Close()
		pc.Close()
		return
	}

	inr.Close()
	outw.Close()
	shutdownChan := make(chan struct{})
	dc.OnOpen(func() {
		err := dc.SendText("OPEN_RTC_CHANNEL")
		if err != nil {
			log.Println("write data error:", err)
		}
		_, err = io.Copy(&Wrap{dc}, outr)
		if err != nil {
			log.Println("io copy error:", err)
		}
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		cmd := string(msg.Data)
		cmd = cmd + "\n"
		if strings.Contains(cmd, "adb") {
			cmd = strings.Replace(cmd, "adb", "adb -s " +*deviceID, 1)
		} else if strings.Contains(cmd, "tidevice") {
			cmd = strings.Replace(cmd, "tidevice", "tidevice -u " +*deviceID, 1)
		} 
		
		_, err := inw.Write([]byte(cmd))
		log.Println("write cmd to shell:", string(cmd))
		if err != nil {
			log.Println("write cmd to shell error:", err)
		}
	})
	dc.OnClose(func() {
		log.Printf("Close underlying proc")
		close(shutdownChan)
	})

	go waitForProc(proc, shutdownChan)
}

func waitForProc(proc *os.Process, shutdownChan chan struct{}) {

	select {
	case <-shutdownChan:
		// A bigger bonk on the head.
		if err := proc.Signal(os.Kill); err != nil {
			log.Println("term:", err)
		}
	}

	if _, err := proc.Wait(); err != nil {
		log.Println("wait:", err)
	}
}

func internalErrorRTC(dc *webrtc.DataChannel, msg string, err error) {
	log.Println(msg, err)
	dc.SendText("Internal server error.")
}
