package server

import (
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

type IRouter interface {
	Public(url, path string)
	Group(relpath string, mid ...interface{}) IGroup
}

type IGroup interface {
	Get(path string, action interface{})
	Post(path string, action interface{})
	Patch(path string, action interface{})
	Delete(path string, action interface{})

	Group(path string, mid ...interface{}) IGroup
}

const (
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEPROTOBUF          = "application/x-protobuf"
)

type CtrlBase struct {
	C     *gin.Context
	abort bool
}

func (c CtrlBase) requestType() string {
	accept := c.C.Request.Header["Accept"]
	if len(accept) == 0 {
		return MIMEHTML
	}

	accept = strings.Split(accept[0], ",")

	return accept[0]
}

func (c CtrlBase) IsJsonReqest() bool {
	return c.requestType() == MIMEJSON
}

func (c *CtrlBase) Abort() {
	c.abort = true
}

var ctrlFuncReg = regexp.MustCompile(`.*\.([A-Za-z_]+\w*)\.([A-Za-z_]+\w*)[^\.]*$`)

func getCtrlAndAction(a interface{}) (ctrl string, action string) {
	name := runtime.FuncForPC(reflect.ValueOf(a).Pointer()).Name()
	names := ctrlFuncReg.FindStringSubmatch(name)
	if len(names) != 3 {
		panic("invalid mid name: " + name)
	}
	ctrl = names[1]
	action = names[2]
	return
}

type Router struct {
	e *gin.Engine
}

type Group struct {
	r      *Router
	path   string
	ctrl   string
	action []string
}

func (g *Group) Group(path string, mids ...interface{}) IGroup {

	_g := Group{r: g.r, path: joinPaths(g.path, path), action: append([]string{}, g.action...)}

	for _, mid := range mids {
		_g.addAction(mid)
	}
	return &_g
}

func (g *Group) Get(path string, action interface{}) {
	g.r.e.GET(joinPaths(g.path, path), _h(g, action))
}

func (g *Group) Post(path string, action interface{}) {
	g.r.e.POST(joinPaths(g.path, path), _h(g, action))
}

func (g *Group) Patch(path string, action interface{}) {
	g.r.e.PATCH(joinPaths(g.path, path), _h(g, action))
}

func (g *Group) Delete(path string, action interface{}) {
	g.r.e.DELETE(joinPaths(g.path, path), _h(g, action))
}

func _h(_g *Group, a interface{}) func(c *gin.Context) {
	i := internalhandler{}
	i.action = append([]string{}, _g.action...)

	ctrl, action := getCtrlAndAction(a)
	_g.testAction(ctrl, action)
	i.ctrl = ctrl
	i.action = append(i.action, action)
	return i.handler
}

func (g *Group) testAction(ctrl, action string) {
	if g.ctrl == "" {
		if _, b := controllers[ctrl]; !b {
			panic("ctrl not found: " + ctrl)
		}
	} else if g.ctrl != ctrl {
		panic("all actions should be the same ctrl: " + g.ctrl + " != " + ctrl)
	}

	//new controler
	ctrlType := controllers[ctrl]
	method, b := ctrlType.Type.MethodByName(action)
	if !b {
		panic("method not found: " + ctrl + "." + action)
	}

	if method.Type.NumIn() > 1 {
		panic("there should be no parameters: " + ctrl + "." + action)
	}
}

func (g *Group) addAction(a interface{}) {
	ctrl, action := getCtrlAndAction(a)
	g.testAction(ctrl, action)
	g.ctrl = ctrl
	g.action = append(g.action, action)
}

type internalhandler struct {
	ctrl   string
	action []string
}

func (g *internalhandler) handler(c *gin.Context) {

	//new controler
	ctrlType := controllers[g.ctrl]
	ctrl := initNewController(ctrlType, c)

	//actions
	for _, mname := range g.action {
		ctrlv := reflect.ValueOf(ctrl)
		methodValue := ctrlv.MethodByName(mname)
		methodValue.Call([]reflect.Value{})
		if ctrlv.Elem().FieldByName("abort").Bool() {
			break
		}
	}
}

func lastChar(str string) uint8 {
	size := len(str)
	if size == 0 {
		panic("The length of the string can't be 0")
	}
	return str[size-1]
}

func joinPaths(path1, path2 string) string {
	if len(path1) == 0 {
		return path2
	}

	if len(path2) == 0 {
		return path1
	}

	finalPath := path.Join(path1, path2)
	if lastChar(path1) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

func (r Router) Public(url, path string) {
	r.e.Static(url, path)
}

func (r Router) Group(relpath string, mids ...interface{}) IGroup {
	g := Group{r: &r, path: relpath, action: []string{}}

	for _, mid := range mids {
		g.addAction(mid)
	}
	return &g
}
