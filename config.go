package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pion/webrtc/v3"
)

// https://gist.github.com/zziuni/3741933
// stun.l.google.com:19302
// stun1.l.google.com:19302
// stun2.l.google.com:19302
// stun3.l.google.com:19302
// stun4.l.google.com:19302
// stun01.sipphone.com
// stun.ekiga.net
// stun.fwdnet.net
// stun.ideasip.com
// stun.iptel.org
// stun.rixtelecom.se
// stun.schlund.de
// stunserver.org
// stun.softjoys.com
// stun.voiparound.com
// stun.voipbuster.com
// stun.voipstunt.com
// stun.voxgratia.org
// stun.xten.com
// https://www.cnblogs.com/dch0/p/12176159.html
// stun:stun.ideasip.com
// stun:stun.schlund.de
// stun:stun.voiparound.com
// stun:stun.voipbuster.com
// stun:stun.voipstunt.com
// stun:stun.xten.com
// http://www.freeswitch.org.cn/2012/01/11/gong-yong-stun-fu-wu-qi-lie-biao.html
// provserver.televolution.net
// sip1.lakedestiny.cordiaip.com
// stun1.voiceeclipse.net
// stun01.sipphone.com
// stun.callwithus.com
// stun.counterpath.net
// stun.ekiga.net (alias for stun01.sipphone.com)
// stun.ideasip.com (no XOR_MAPPED_ADDRESS support)
// stun.internetcalls.com
// stun.ipns.com
// stun.noc.ams-ix.net
// stun.phonepower.com
// stun.phoneserve.com
// stun.rnktel.com
// stun.softjoys.com (no DNS SRV record) (no XOR_MAPPED_ADDRESS support)
// stunserver.org see their usage policy
// stun.sipgate.net
// stun.sipgate.net:10000
// stun.voip.aebc.com
// stun.voipbuster.com (no DNS SRV record) (no XOR_MAPPED_ADDRESS support)
// stun.voxalot.com
// stun.voxgratia.org (no DNS SRV record) (no XOR_MAPPED_ADDRESS support)
// stun.xten.com
// numb.viagenie.ca (http://numb.viagenie.ca) (XOR_MAPPED_ADDRESS only with rfc3489bis magic number in transaction ID)
// stun.ipshka.com inside UA-IX zone russsian explanation at http://www.ipshka.com/main/help/hlp_stun.php
// https://bbs.csdn.net/topics/390819725
// s1.taraba.net 203.183.172.196:3478
// s2.taraba.net 203.183.172.196:3478
// s1.voipstation.jp 113.32.111.126:3478
// s2.voipstation.jp 113.32.111.127:3478

var (
	configRTC = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
			{
				URLs: []string{"stun:global.stun.twilio.com:3478"},
			},
		},
	}
	defaultHost = "127.0.0.1"
	defaultPort = 22
)

func BuildConfigRTC() {
	timeout := time.Duration(1) * time.Second
	httpClient := http.Client{
		Timeout: timeout,
	}
	resp, err := httpClient.Get("https://food-ginkgo.stg-myteksi.com/data/turnservers.json")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != int(200) {
		return
	}
	config := webrtc.Configuration{}
	err = json.NewDecoder(resp.Body).Decode(&config)
	if err != nil {
		return
	}
	configRTC = config

}
