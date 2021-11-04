package main

import (
	"context"
	"image"
	_ "image/gif"
	_ "image/png"
	"log"
	"sync"
	"time"

	"github.com/pion/mediadevices" // This is required to use h264 video encoder
	"github.com/pion/mediadevices/pkg/codec/vpx"
	"golang.org/x/net/websocket"
)

type PhoneVideoSource struct {
	id          string
	wsAddr      string
	origin      string
	ws          *websocket.Conn
	lock        sync.Mutex
	logger      *log.Logger
	imgChan     chan image.Image
	dftImg      image.Image
	stopDFTImag bool
}

var _ mediadevices.VideoSource = &PhoneVideoSource{}

// call func NewVideoTrack(source VideoSource, selector *CodecSelector) Track { to creat the track

func NewH264Selector() *mediadevices.CodecSelector {
	vp8Params, err := vpx.NewVP8Params()
	if err != nil {
		panic(err)
	}
	vp8Params.BitRate = 64_000 // 64kbps

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&vp8Params),
	)
	return codecSelector
}

func NewPhoneVideoSource(id, wsAddr, wsOrigin string) (*PhoneVideoSource, context.CancelFunc) {
	pvs := &PhoneVideoSource{
		id:          id,
		wsAddr:      wsAddr,
		origin:      wsOrigin,
		ws:          nil,
		logger:      log.Default(),
		imgChan:     make(chan image.Image, 5),
		dftImg:      image.NewGray(image.Rect(0, 0, 450, 800)),
		stopDFTImag: false,
	}
	cancel := pvs.Start()
	return pvs, cancel
}

func (source *PhoneVideoSource) ID() string { return source.id }
func (source *PhoneVideoSource) Close() error {
	source.cleanup()
	return nil
}
func (source *PhoneVideoSource) Read() (image.Image, func(), error) {
	if source.stopDFTImag {
		// source.logger.Print("reading image")
		img, ok := <-source.imgChan
		if !ok {
			// source.logger.Printf("=====%s\n", img.Bounds())
			// return source.dftImg, source.release, nil
			panic("could not get image")
		}
		return img, source.release, nil
	}
	select {
	case img, ok := <-source.imgChan:
		if !ok {
			return source.dftImg, source.release, nil
		}
		source.stopDFTImag = true
		return img, source.release, nil
	case <-time.After(10 * time.Millisecond):
		return source.dftImg, source.release, nil
	}
}
func (source *PhoneVideoSource) release() {
}
func (source *PhoneVideoSource) cleanup() {
	if source.ws != nil {
		err := source.ws.Close()
		if err != nil {
			log.Print("error when closing websocket:", err)
		}
		source.ws = nil
	}
}
func (source *PhoneVideoSource) Start() context.CancelFunc {
	ctx := context.TODO()
	ctx2, cancel := context.WithCancel(ctx)
	go source.fetchImages(ctx2)
	return cancel
}

func (source *PhoneVideoSource) fetchImages(ctx context.Context) {
	// origin := "http://localhost/"
	// url := "ws://localhost:20001/minicap"
	// url := "ws://ITCN000021-MAC.local:20039/screen"
	// url := "ws://ITCN000021-MAC.local:20130/minicap"
	ws, err := websocket.Dial(source.wsAddr, "", source.origin)
	if err != nil {
		log.Fatal(err)
	}
	source.ws = ws
	for {
		select {
		case <-ctx.Done():
			source.cleanup()
			return
		default:
			// do nothing
		}
		// ignore headers,and read at least on image
		var msg = make([]byte, 100)
		var m int = 0
		if m, err = ws.Read(msg); err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("%s\n",msg[:m])
		// index out of range, len(msg) >=  2
		if m >= 2 && msg[m-1] == 0xD9 && msg[m-2] == 0xFF {
			break
		}
	}
	var n int = 0
READIMGData:
	for ; ; n++ {
		// img, imgFmt, err := image.Decode(ws)
		img, _, err := image.Decode(ws)
		select {
		case <-ctx.Done():
			source.cleanup()
			return
		default:
			// do nothing
		}
		if err != nil {
		READNonImageData:
			for {
				select {
				case <-ctx.Done():
					source.cleanup()
					return
				default:
					// do nothing
				}
				// ignore headers,and read at least on image
				var msg = make([]byte, 100)
				var m int = 0
				if m, err = ws.Read(msg); err != nil {
					log.Fatal(err)
				}
				// fmt.Printf("%s\n",msg[:m])
				if m >= 2 && msg[m-1] == 0xD9 && msg[m-2] == 0xFF {
					break READNonImageData
				}
			}
			continue READIMGData
		}
		// if n == 1000 {
		// 	break
		// }
		// fmt.Printf("=====%d:%s:%s\n", n, imgFmt, img.Bounds())
		select {
		case source.imgChan <- img:
			// source.logger.Println(imgFmt)
			// if imgFmt == "jpeg" {
			// 	w, err := os.Create(fmt.Sprintf("%d.jpeg", n))
			// 	if err != nil {
			// 		source.logger.Println("when create file", err)
			// 	}
			// 	defer w.Close()
			// 	err = jpeg.Encode(w, img, &jpeg.Options{
			// 		Quality: 80,
			// 	})
			// 	if err != nil {
			// 		source.logger.Println("when write file", err)
			// 	}
			// }
			// try to send the image to encoder
			// source.logger.Printf("%d succeeded", n)
		default:
			// do nothing
			// source.logger.Printf("xxxxxxxxxxx%d Failed", n)
		}

	}
}
