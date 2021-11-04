// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"

	"github.com/rs/cors"
	// "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	addr        = flag.String("addr", ":8081", "http service address")
	minicapAddr = flag.String("minicap", "", "localhost minicap port")
	deviceID    = flag.String("deviceid", "", "the device udid")
	deviceType  = flag.String("devicetype", "android", "the device type 'android' or 'ios'")
	cmdPath     string
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	BuildConfigRTC()
	flag.Parse()
	if *deviceID == "" || *minicapAddr == "" {
		flag.PrintDefaults()
		return
	}

	var err error
	cmdPath, err = exec.LookPath("/bin/sh")
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/signal", serveSignal)
	router.Handle("/{.*}", http.FileServer(http.Dir(".")))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "HEAD", "POST", "PUT", "OPTIONS"},
		AllowedHeaders: []string{},
	})

	handler := c.Handler(router)

	// http.Handle("/", http.FileServer(http.Dir(".")))
	// http.HandleFunc("/cmd", serveWs)
	// http.HandleFunc("/signal", serveSignal)
	// log.Fatal(http.ListenAndServe(*addr, handler))
	server := &http.Server{Addr: *addr, Handler: handler}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
	termChan := make(chan os.Signal)
	signal.Notify(termChan, os.Interrupt, os.Kill)
	<-termChan
	server.Close()
}
