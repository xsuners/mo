package xtcp

import (
	"fmt"
	"net"
)

// Config .
type Config struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// Wrapper .
type Wrapper struct {
	server *Server
	conf   *Config
}

// New .
func New(c *Config, opt ...ServerOption) *Wrapper {
	w := new(Wrapper)
	w.conf = c

	s := NewServer(opt...)
	w.server = s

	return w
}

// Server .
func (w *Wrapper) Server() *Server {
	return w.server
}

// Start .
func (w *Wrapper) Start() (err error) {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", w.conf.IP, w.conf.Port))
	if err != nil {
		return
	}

	w.server.Start(l)

	return
}

// Stop .
func (w *Wrapper) Stop() (err error) {
	w.server.Stop()
	return
}
