package server

import (
	"github.com/yangsongfwd/backup/log"
)

var Server *_Server

type Hook func()

type _Server struct {
	beforStartHook []Hook
	beforStopHook  []Hook
	globalData     map[string]interface{}
}

func (s *_Server) Run() {

	for _, h := range s.beforStartHook {
		h()
	}

}

func (s *_Server) Stop() {
	for _, h := range s.beforStopHook {
		h()
	}

	log.Close()
}

func (s *_Server) RegBeforStart(h Hook) {
	s.beforStartHook = append(s.beforStartHook, h)
}

// TODO: order?
func (s *_Server) RegBeforStop(h Hook) {
	s.beforStopHook = append(s.beforStopHook, h)
}

func (s *_Server) AddGlobalData(key string, data interface{}) {
	s.globalData[key] = data
}

func (s *_Server) GetGlobalData(key string) interface{} {
	return s.globalData[key]
}

func init() {
	Server = &_Server{[]Hook{}, []Hook{}, map[string]interface{}{}}

	//reg log
	Server.RegBeforStart(func() {
		log.SetLevel(log.LevelInfo | log.LevelDebug | log.LevelWarn | log.LevelError)
	})

}
