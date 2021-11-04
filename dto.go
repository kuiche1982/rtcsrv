package main

type Config struct {
	Uuid string `ini:"uuid,identify"`
	Host string `ini:"host,ssh"`
	Port int    `ini:"port,ssh"`
}

type Session struct {
	Type  string `json:"type"`
	Sdp   string `json:"sdp"`
	Error string `json:"error"`
}
