package ios

import (
	"bytes"
	"encoding/json"
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

// (sessionID, data) result.  data could be  empty
type opFun func([]byte) string

var opMap = map[string]opFun{
	"NMT:/info/rotation": getRotation,
	"NMT:/tap":           Tap,
	"NMT:/perform":       perform,
	"NMT:/homescreen":    pressHome,
	"noop":               noop,
}

func noop(_ []byte) string {
	return "noop"
}

type Status struct {
	SessionID string `json:"sessionId"`
}

type jsonMsg struct {
	Op   string          `json:"op"`
	Data json.RawMessage `json:"data"`
}

func getRotation(_ []byte) string {
	return defaultRotation
}

var (
	client          = &http.Client{}
	minicapAddr     = ""
	minicapAddrLock = sync.Mutex{}
	sessionID       = ""
	sessionIDLock   = sync.Mutex{}
)

func getSessionID() string {
	sessionIDLock.Lock()
	defer sessionIDLock.Unlock()
	return sessionID
}

func setSessionID(id string) {
	sessionIDLock.Lock()
	defer sessionIDLock.Unlock()
	sessionID = id
}

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

func pressHome(data []byte) string {
	// 	curl 'http://10.10.69.87:4001/10.6.24.149/20002/wda/homescreen' \
	//   -X 'POST' \
	//   -H 'Connection: keep-alive' \
	//   -H 'Content-Length: 0' \
	//   -H 'Pragma: no-cache' \
	//   -H 'Cache-Control: no-cache' \
	//   -H 'Accept: */*' \
	//   -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36' \
	//   -H 'Origin: http://localhost:4000' \
	//   -H 'Referer: http://localhost:4000/' \
	//   -H 'Accept-Language: en-US,en;q=0.9' \
	//   --compressed \
	//   --insecure
	port := getMinicapAddr()
	url := fmt.Sprintf("http://localhost:%s/wda/homescreen", port)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Print("IOS:pressHome:http.NewRequest", err)
		return err.Error()
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("IOS:pressHome:client.Do", err)
		return err.Error()
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Print("IOS:pressHome:resp.Status", resp.Status)
		sid := refreshSessionID()
		setSessionID(sid)
		return "error: resp code not 200"
	}
	if err != nil {
		log.Print("IOS:pressHome:read resp.Body", err)
		return err.Error()
	}
	return "done"
}

func Tap(data []byte) string {
	// 	curl 'http://10.10.69.87:4001/10.28.146.231/20007/session/4A31E598-F8A1-4697-80F7-FB318B591645/wda/tap/0' \
	//   -H 'Connection: keep-alive' \
	//   -H 'Pragma: no-cache' \
	//   -H 'Cache-Control: no-cache' \
	//   -H 'Accept: */*' \
	//   -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36' \
	//   -H 'Content-Type: application/x-www-form-urlencoded; charset=UTF-8' \
	//   -H 'Origin: http://10.10.69.87:4000' \
	//   -H 'Referer: http://10.10.69.87:4000/' \
	//   -H 'Accept-Language: en-US,en;q=0.9' \
	//   --data-raw '{"x":174,"y":406}' \
	//   --compressed \
	//   --insecure

	// Status Code: 200 OK

	// {
	// 	"value" : null,
	// 	"sessionId" : "4A31E598-F8A1-4697-80F7-FB318B591645"
	//   }
	port := getMinicapAddr()
	session := getSessionID()
	url := fmt.Sprintf("http://localhost:%s/session/%s/wda/tap/0", port, session)
	bodyReader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		log.Print("IOS:Tap:http.NewRequest", err)
		return err.Error()
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("IOS:Tap:client.Do", err)
		return err.Error()
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Print("IOS:Tap:resp.Status", resp.Status)
		sid := refreshSessionID()
		setSessionID(sid)
		return "error: resp code not 200"
	}
	if err != nil {
		log.Print("IOS:Tap:read resp.Body", err)
		return err.Error()
	}
	return "done"
}

// func keepAlive(stopChan <-chan struct{}) {
// 	for {
// 		select {
// 		case <-stopChan:
// 			return
// 		case time.After(3 * time.Second):
// 			// do the keep alive call
// 		}
// 	}

// 	// 	curl 'http://10.10.69.87:4000/api/v1/user/devices/ff8e7107eac04b83c6ec1ce36324fb74b88e51cb/active' \
// 	//   -H 'Connection: keep-alive' \
// 	//   -H 'Pragma: no-cache' \
// 	//   -H 'Cache-Control: no-cache' \
// 	//   -H 'Accept: application/json, text/javascript, */*; q=0.01' \
// 	//   -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36' \
// 	//   -H 'X-Requested-With: XMLHttpRequest' \
// 	//   -H 'Referer: http://10.10.69.87:4000/devices/ff8e7107eac04b83c6ec1ce36324fb74b88e51cb/remotecontrolrtc' \
// 	//   -H 'Accept-Language: en-US,en;q=0.9' \
// 	//   -H 'Cookie: user_id=2|1:0|10:1629708042|7:user_id|28:a3VpLmNoZW5AZ3JhYnRheGkuY29t|33ac4c50ab09ccc01c8265084d26752810e07d4d7569d214b5a5bb3d1b67ef54' \
// 	//   --compressed \
// 	//   --insecure
// 	// Code: 200 OK

// }

func perform(data []byte) string {
	// 	curl 'http://10.10.69.87:4001/10.28.146.231/20007/session/4A31E598-F8A1-4697-80F7-FB318B591645/wda/touch/perform' \
	//   -H 'Connection: keep-alive' \
	//   -H 'Pragma: no-cache' \
	//   -H 'Cache-Control: no-cache' \
	//   -H 'Accept: */*' \
	//   -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36' \
	//   -H 'Content-Type: application/x-www-form-urlencoded; charset=UTF-8' \
	//   -H 'Origin: http://10.10.69.87:4000' \
	//   -H 'Referer: http://10.10.69.87:4000/' \
	//   -H 'Accept-Language: en-US,en;q=0.9' \
	//   --data-raw '{"actions":[{"action":"press","options":{"x":245,"y":296}},{"action":"wait","options":{"ms":100}},{"action":"moveTo","options":{"x":236,"y":667}},{"action":"release","options":{}}]}' \
	//   --compressed \
	//   --insecure

	port := getMinicapAddr()
	session := getSessionID()
	url := fmt.Sprintf("http://localhost:%s/session/%s/wda/touch/perform", port, session)
	bodyReader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		log.Print("IOS:perform:http.NewRequest", err)
		return err.Error()
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("IOS:perform:client.Do", err)
		return err.Error()
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Print("IOS:perform:resp.Status", resp.Status)
		sid := refreshSessionID()
		setSessionID(sid)
		return "error: resp code not 200"
	}
	if err != nil {
		log.Print("IOS:perform:read resp.Body", err)
		return err.Error()
	}
	return "done"
}

func createSession() string {
	// POST 10.28.146.231/20007/session
	// Return 200 {sessionId:"thesessionid"}
	bodyReader := bytes.NewReader([]byte("{\"capabilities\": {}}"))
	req, err := http.NewRequest("POST", "http://localhost:"+minicapAddr+"/session", bodyReader)
	if err != nil {
		log.Print("createSession:http.NewRequest", err)
		return ""
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Print("createSession:client.Do", err)
		return ""
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Print("createSession:resp.Status", resp.Status)
		return ""
	}
	if err != nil {
		log.Print("createSession:read resp.Body", err)
		return ""
	}

	status := &Status{}
	err = json.Unmarshal(result, status)
	if err != nil {
		log.Print("createSession:json decode", err, string(result))
		return ""
	}

	return status.SessionID
}

func refreshSessionID() string {
	// GET 10.28.146.231/20007/status
	// Return 200 {sessionId:"thesessionid"}
	req, err := http.NewRequest("GET", "http://localhost:"+minicapAddr+"/status", nil)
	if err != nil {
		log.Print("http.NewRequest", err)
		return ""
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Print("client.Do", err)
		return ""
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Print("resp.Status", resp.Status)
		return ""
	}
	if err != nil {
		log.Print("read resp.Body", err)
		return ""
	}

	status := &Status{}
	err = json.Unmarshal(result, status)
	if err != nil {
		log.Print("json decode", err, string(result))
		return ""
	}
	if status.SessionID == "" {
		return createSession()
	}
	return status.SessionID
}

func getCmdAndData(msg []byte) (string, []byte) {
	// byte tojsonMsg
	jmsg := &jsonMsg{}
	err := json.Unmarshal(msg, jmsg)
	if err != nil {
		log.Print("ios:getCmdAndData:json decode", err, string(msg))
		return "noop", []byte{}
	}
	if jmsg == nil {
		return "noop", []byte{}
	}
	return jmsg.Op, []byte(jmsg.Data)
}

func ChannelWDATouch(opPort string, dc *webrtc.DataChannel, pc *webrtc.PeerConnection) {
	setMinicapAddr(opPort)
	sid := refreshSessionID()
	setSessionID(sid)

	ws, err := websocket.Dial("ws://localhost:"+opPort+"/screen", "", "http://localhost:"+opPort)
	if err != nil {
		dc.SendText(err.Error())
		dc.Close()
		log.Fatal(err)
	}
	// doneChan := make(chan struct{})

	dc.OnOpen(func() {
		err := dc.SendText("OPEN_MINITOUCH_CHANNEL")
		if err != nil {
			log.Println("IOS:ChannelWDATouch:OnOpen:write data error:", err)
		}
		// go keepAlive(doneChan)
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		opAction, data := getCmdAndData(msg.Data)
		log.Print("IOS:ChannelWDATouch:OnMessage:minitouch fwd op: ", opAction, ", data: ", string(data))
		if op, ok := opMap[opAction]; ok {
			result := op(data)
			dc.SendText(result)
			return
		}
		ws.Write(msg.Data)
	})
	dc.OnClose(func() {
		// ws.Close()
		log.Printf("Close underlying proc")
	})
}
