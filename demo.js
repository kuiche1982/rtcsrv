/* remote control communicate with minitouch */
function coords(boundingW, boundingH, relX, relY, rotation) {
  /**
* Rotation affects the screen as follows:
*
*             0deg
*           |------|
*           | MENU |
*           |------|
*      -->  |      |  --|
*      |    |      |    v
*           |      |
*           |      |
*           |------|
*        |----|-|          |-|----|
*        |    |M|          | |    |
*        |    |E|          | |    |
*  90deg |    |N|          |U|    | 270deg
*        |    |U|          |N|    |
*        |    | |          |E|    |
*        |    | |          |M|    |
*        |----|-|          |-|----|
*           |------|
*      ^    |      |    |
*      |--  |      |  <--
*           |      |
*           |      |
*           |------|
*           | UNEM |
*           |------|
*            180deg
*
* Which leads to the following mapping:
*
* |--------------|------|---------|---------|---------|
* |        | 0deg |  90deg  |  180deg |  270deg |
* |--------------|------|---------|---------|---------|
* | CSS rotate() | 0deg | -90deg  | -180deg |  90deg  |
* | bounding w   |  w   |    h    |    w    |    h    |
* | bounding h   |  h   |    w    |    h    |    w    |
* | pos x        |  x   |   h-y   |   w-x   |    y    |
* | pos y        |  y   |    x    |   h-y   |   h-x   |
* |--------------|------|---------|---------|---------|
*/
  var w, h, x, y;

  switch (rotation) {
    case 0:
      w = boundingW
      h = boundingH
      x = relX
      y = relY
      break
    case 90:
      w = boundingH
      h = boundingW
      x = boundingH - relY
      y = relX
      break
    case 180:
      w = boundingW
      h = boundingH
      x = boundingW - relX
      y = boundingH - relY
      break
    case 270:
      w = boundingH
      h = boundingW
      x = relY
      y = boundingW - relX
      break
  }

  return {
    xP: x / w,
    yP: y / h,
  }
}

function AddRemoteControl(videoElement) {
  Object.assign(videoElement, {
    rotation: 0,
    autoplay: true, 
    controls: false,
    // onaudioprocess(e){
    //   e.preventDefault() 
    // },
    // onloadeddata(e) {
    //   console.log("loadeddata")
    // },
    // onplay(e){
    //   let videoElement = this
    //   this.syncTouchpad(videoElement)
    //   console.log("play")
    // },
    // onchange(e){
    //   console.log("change", e)
    // },
    // onresize(e){
    //   this.getRotation()
    //   console.log("resize", e)
    // },
    onpause(e){
      this.play()
    },
    oncontextmenu(e) {
      return false
    },
    getRotation(){
      if (!this.channel || this.channel.readyState != "open") {
        this.rotation = 0
        return
      }
      this.channel.send("NMT:/info/rotation")
    },
    syncTouchpad(element){
      let channel = element.channel
      if (!channel || channel.readyState != "open") {
        channel = pc.createDataChannel("minitouch")
      }
      element.channel = channel

      let touchSync = function(operation, event){
        var e = event;
        if (e.originalEvent) {
            e = e.originalEvent
        }
        e.preventDefault()
        
        let x = e.offsetX, y = e.offsetY
        let w = e.target.clientWidth, h = e.target.clientHeight
        let scaled = coords(w, h, x, y, e.target.rotation);
        console.log("touchSync", e, e.offsetX, e.offsetY, e.target.clientWidth, e.target.clientHeight, e.target.rotation, element.rotation, scaled)
        channel.send(JSON.stringify({
            operation: operation, // u, d, c, w
            index: 0,
            pressure: 0.5,
            xP: scaled.xP,
            yP: scaled.yP,
        }))
        channel.send(JSON.stringify({ operation: 'c' }))
      }
      let mouseMoveListener = function(event) {
        touchSync('m', event)
      }
      let mouseUpListener = function(event) {
        touchSync('u', event)
        element.removeEventListener('mousemove', mouseMoveListener);
        document.removeEventListener('mouseup', mouseUpListener);
      }
      let mouseDownListener = function(event) {
        touchSync('d', event)
        element.addEventListener('mousemove', mouseMoveListener);
        document.addEventListener("mouseup", mouseUpListener)
      }

      channel.onopen  = function(e){
        console.log("minitouch channel open", e)
        element.readonly = false
        console.log("minitouch connected")
        element.getRotation()
        channel.send(JSON.stringify({ // touch reset, fix when device is outof control
          operation: "r",
        }))
        element.addEventListener("mousedown", mouseDownListener)
      }
      channel.onclose = function(e){
        element.readonly = true
        console.log("minitouch closed")
        element.removeEventListener("mousedown", mouseDownListener)
      }
      channel.onmessage = function(e){
        console.log("minitouch msg", e,  element)
        try {
          json = JSON.parse(e.data)
          if (json) {
            console.log("resized:", json)
            element.rotation = json.rotation
          } 
        }catch(e)  {
          console.log("resized exception:", 0)
          element.rotation = 0
        }
      }
    }
  })
  videoElement.addEventListener("loadeddata", e => {
    videoElement.syncTouchpad(videoElement)
  })
  videoElement.addEventListener("resize", videoElement.getRotation)
  return videoElement
}

/* RTC Connection */

let pc = new RTCPeerConnection({
  iceServers: [
    {
      urls: "stun:stun.l.google.com:19302"
    },
    {
      urls: "turn:10.6.18.69:34778",
      username: "user1",
      credential: "pass1"
    },
    {
      urls: "turn:10.10.69.87:3478",
      username: "user1",
      credential: "pass1"
    }
  ]
})
let log = msg => {
  console.log(msg)
  // document.getElementById('logs').innerHTML = msg + '<br>'+document.getElementById('logs').innerHTML
}

pc.ontrack = function (event) {
  var el = document.createElement(event.track.kind)
  el.srcObject = event.streams[0]
  el = AddRemoteControl(el)
  el.autoplay = true
  document.getElementById('remoteVideos').appendChild(el)
}

pc.addTransceiver('video', {'direction':"recvonly"})
// pc.addTransceiver('video')

let dec = new TextDecoder("ascii")
let cmdChannel = pc.createDataChannel('CMD')
cmdChannel.onclose = () => console.log('cmdChannel has closed')
cmdChannel.onopen = () => {
  console.log('cmdChannel has opened')
  for(i = 0; i< 10 ; i ++ ){
    cmdChannel.send("adb shell input keyevent HOME")
  }
  // setInterval(function() {
  //   let selectedPair = pc.sctp.transport.iceTransport.getSelectedCandidatePair()

  //   console.log("local", selectedPair.local.candidate)
  //   console.log("remote",selectedPair.remote.candidate)
  // }, 3000);
}
cmdChannel.onmessage = e => {
  console.log(e)
  let msg = e.data
  if (typeof(msg) !== "string"){
    log('decoding')
    msg = dec.decode(msg)
    msg = msg.split('\n').join("<br>")
  }

  log(`Message from DataChannel '${cmdChannel.label}' payload <br>'${msg}'`)
}

pc.oniceconnectionstatechange = e => {
  log("oniceconnectionstatechange")
  log(pc.iceConnectionState)
  console.log(e)
}
pc.onicecandidate = event => {
  // log("onicecandidate")
  // log(JSON.stringify(event))
  // console.log(event)
}

// pc.onnegotiationneeded = e =>{
//   log("onnegotiationneeded")
//   pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)
// }
  

cmdInput = (e) => {
  let element = document.getElementById('message')
  if(element && element==document.activeElement && e.keyCode == 13){
    let message = element.value

    if (message === '') {
      return alert('Message must not be empty')
    }
    log("sending:"+message)
    cmdChannel.send(message)
  }
  
}

sendMessage = (e) => {
  let element = document.getElementById('message')
  if(element){
    let message = element.value

    if (message === '') {
      return alert('Message must not be empty')
    }
    log("sending:"+message)
    cmdChannel.send(message)
  }
  
}

fixRotation = () => {
  let video = document.getElementById("remoteVideos").firstElementChild
  console.log
  video.syncTouchpad(video)
}

window.onload = () => {
  if (window["WebSocket"]) {
    url =  document.location.href.replace("http://", "ws://").replace("demo.html", "signal")
    conn = new WebSocket(url);
    conn.onclose = function (evt) {
      log("conn.onclose")
    };
    conn.onmessage = function (evt) {
      log("conn.onmessage")
      var answer = JSON.parse(evt.data)
      log(JSON.stringify(answer))
      remoteSDP = new RTCSessionDescription(answer)      
      pc.setRemoteDescription(remoteSDP).catch(log)
    };
    conn.onopen = function(evt){
      log("connectionopen")
      try {
        pc.createOffer().then(d => {
          pc.setLocalDescription(d)
          log("offer:"+JSON.stringify(d))
          conn.send(JSON.stringify(d))
        })
      } catch (e) {
        alert(e)
      }
    }
  } else {
      var item = document.createElement("div");
      item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
      appendLog(item);
  }
}

window.addEventListener("beforeunload", e => {
  pc.close()
})