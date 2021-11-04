package main

import (
	"errors"
	"fmt"
	"log"

	"rtcsrv/touch/android"
	"rtcsrv/touch/ios"

	"github.com/gorilla/websocket"
	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v3"
)

type Wrap struct {
	*webrtc.DataChannel
}

var pc *webrtc.PeerConnection

func (rtc *Wrap) Write(data []byte) (int, error) {
	err := rtc.DataChannel.Send(data)
	return len(data), err
}

func hub(ws *websocket.Conn, doneChan chan struct{}) {
	var msg Session
	for {
		err := ws.ReadJSON(&msg)
		if err != nil {
			_, ok := err.(*websocket.CloseError)
			if !ok {
				log.Println("websocket", err)
			}
			close(doneChan)
			break
		}
		err = startRTC(ws, msg, doneChan)
		if err != nil {
			log.Println(err)
			close(doneChan)
			break
		}
		select {
		case <-doneChan:
			break
		default:
		}
	}
}

func startRTC(ws *websocket.Conn, data Session, doneChan chan struct{}) error {
	if data.Error != "" {
		return fmt.Errorf(data.Error)
	}

	switch data.Type {
	case "offer":
		var err error
		pc, err = webrtc.NewPeerConnection(configRTC)
		if err != nil {
			return err
		}

		pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
			log.Println("ICE Connection State has changed:", state.String())
		})

		pc.OnDataChannel(func(dc *webrtc.DataChannel) {
			switch dc.Label() {
			case "CMD":
				ChannelCmd(dc, pc)
			case "minitouch":
				if *deviceType == "ios" {
					ios.ChannelWDATouch(*minicapAddr, dc, pc)
					return
				}
				android.ChannelMinitouch(*minicapAddr, dc, pc)
			}
		})

		var wsurl, wsorigin = "ws://localhost:" + *minicapAddr + "/minicap", "http://localhost:" + *minicapAddr
		if *deviceType == "ios" {
			wsurl, wsorigin = "ws://localhost:"+*minicapAddr+"/screen", "http://localhost:"+*minicapAddr
		}
		phonesource, stopPhoneSource := NewPhoneVideoSource(*deviceID, wsurl, wsorigin)
		stopPC := func() {
			stopPhoneSource()
			pc.Close()
		}
		phonetrack := mediadevices.NewVideoTrack(phonesource, NewH264Selector())
		v2track, ok := phonetrack.(webrtc.TrackLocal)
		if !ok {
			log.Println("AddTrack error:", err)
			stopPhoneSource()
			pc.Close()
			return errors.New("video track is not webrtc.Track")
		}
		if _, err := pc.AddTrack(v2track); err != nil {
			log.Println("AddTrack error:", err)
			stopPhoneSource()
			pc.Close()
			return err
		}
		if err := pc.SetRemoteDescription(webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  data.Sdp,
		}); err != nil {
			stopPC()
			return err
		}
		gatherComplete := webrtc.GatheringCompletePromise(pc)
		answer, err := pc.CreateAnswer(nil)
		if err != nil {
			stopPC()
			return err
		}
		err = pc.SetLocalDescription(answer)
		if err != nil {
			stopPC()
			return err
		}
		log.Println(<-gatherComplete)

		if err = ws.WriteJSON(pc.LocalDescription()); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown signaling message %v", data.Type)
	}
	return nil
}
