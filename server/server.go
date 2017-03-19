package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yangsongfwd/backup/log"
	"github.com/yangsongfwd/ddns/config"
)

var Server *_Server

type Hook func()
type RouteFunc func(IRouter)

type _Server struct {
	beforStartHook []Hook
	beforStopHook  []Hook
	globalData     map[string]interface{}

	regRoute RouteFunc
	r        Router
	t        *TmplHelper
}

func (s *_Server) loadTemplate() {
	s.t = &TmplHelper{E: s.r.e}

	curPath := config.CurDir()
	curPath += "/ddns_static"
	s.t.LoadMainTmpl(curPath + "/tmpl/*.*")
	s.t.LoadView(curPath + "/tmpl/views/*.*")
}

func (s *_Server) Run(regRoute RouteFunc) {
	for _, h := range s.beforStartHook {
		h()
	}

	s.r.e = gin.Default()
	if !config.Data.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	s.loadTemplate()
	regRoute(s.r)
	go func() {
		s.r.e.Run(":" + fmt.Sprintf("%v", config.Data.Web.Port))
	}()
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
	Server = &_Server{beforStartHook: []Hook{}, beforStopHook: []Hook{}, globalData: map[string]interface{}{}}

	//reg log
	Server.RegBeforStart(func() {
		log.SetLevel(log.LevelInfo | log.LevelDebug | log.LevelWarn | log.LevelError)
	})
}
