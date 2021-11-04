package android

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/pion/webrtc/v3"
	"golang.org/x/net/websocket"
)

const (
	NMTPrefix       = "NMT:"
	defaultRotation = "{\"rotation\":0}"
)

type opFun func(string) string

var opMap = map[string]opFun{
	"NMT:/info/rotation": getRotation,
}

var (
	client          = &http.Client{}
	minicapAddr     = ""
	minicapAddrLock = sync.Mutex{}
)

func getMinicapAddr() string {
	minicapAddrLock.Lock()
	defer minicapAddrLock.Unlock()
	return minicapAddr
}

func setMinicapAddr(id string) {
	minicapAddrLock.Lock()
	defer minicapAddrLock.Unlock()
	minicapAddr = id
}

func getRotation(_ string) string {
	minicapAddr := getMinicapAddr()
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%s/info/rotation", minicapAddr), nil)
	if err != nil {
		log.Print("http.NewRequest", err)
		return defaultRotation
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("client.Do", err)
		return defaultRotation
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Print("resp.Status", resp.Status)
		return defaultRotation
	}
	if err != nil {
		log.Print("read resp.Body", err)
	}
	return string(result)

}

func ChannelMinitouch(opPort string, dc *webrtc.DataChannel, pc *webrtc.PeerConnection) {
	setMinicapAddr(opPort)
	ws, err := websocket.Dial(fmt.Sprintf("ws://localhost:%s/minitouch", opPort), "", fmt.Sprintf("http://localhost:%s", opPort))
	if err != nil {
		dc.SendText(err.Error())
		dc.Close()
		log.Fatal(err)
	}
	go func() {
		for {
			buf := []byte{}
			n, err := ws.Read(buf)
			if err != nil {
				log.Println("minitouch data error:", err)
				return
			}
			if n > 0 {
				log.Println("minitouch data:", string(buf))
			}

		}

	}()

	dc.OnOpen(func() {
		err := dc.SendText("OPEN_MINITOUCH_CHANNEL")
		if err != nil {
			log.Println("write data error:", err)
		}
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		data := string(msg.Data)
		log.Print("minitouch fwd data", data)
		if op, ok := opMap[data]; ok {
			result := op(data)
			dc.SendText(result)
			return
		}
		ws.Write(msg.Data)
	})
	dc.OnClose(func() {
		ws.Close()
		log.Printf("Close underlying proc")
	})
}
