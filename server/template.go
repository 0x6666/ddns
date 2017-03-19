package server

import (
	"bytes"
	temp "html/template"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/yangsongfwd/backup/log"
)

//实现HTMLRender，主要用于调试时，可以实现动态加载有可以导入funcmap
type HtmlDebug struct {
	Glob    string
	FuncMap temp.FuncMap
}

func (r HtmlDebug) Instance(name string, data interface{}) render.Render {
	return render.HTML{
		Template: r.loadTemplate(),
		Name:     name,
		Data:     data,
	}
}

func (r HtmlDebug) loadTemplate() *temp.Template {
	if len(r.Glob) > 0 {
		return temp.Must(temp.New("main").Funcs(r.FuncMap).ParseGlob(r.Glob))
	}
	panic("the HTML debug render was created without files or glob pattern")
}

//TmplHelper 用动态加载模板的辅助参数
type TmplHelper struct {
	E *gin.Engine

	//调试环境
	viewPattern string
	viewFuncMap temp.FuncMap

	//生产环境用的
	views *temp.Template
}

func (th *TmplHelper) renderView(view string, pipelines map[string]interface{}) (temp.HTML, error) {

	log.Debug("render view [%v]", view)

	var t *temp.Template
	if gin.IsDebugging() && len(th.viewPattern) > 0 {
		t = temp.Must(temp.New("views").Funcs(th.viewFuncMap).ParseGlob(th.viewPattern))
	} else {
		t = th.views
	}

	var v *temp.Template
	if v = t.Lookup(view); v == nil {
		err := fmt.Errorf("View [%v] not exist...", view)
		log.Error(err.Error())
		return "load view failed...", err
	}

	var buf bytes.Buffer
	if err := v.Execute(&buf, pipelines); err != nil {
		log.Error("ExecuteTemplate view [%v] failed: [%v]", view, err.Error())
		return "load view failed...", err
	}

	return temp.HTML(buf.String()), nil
}

func (th *TmplHelper) LoadMainTmpl(pattern string) {
	funcMap := temp.FuncMap{}
	funcMap["RenderView"] = th.renderView

	if gin.IsDebugging() {
		th.E.HTMLRender = HtmlDebug{Glob: pattern, FuncMap: funcMap}
	} else {
		templ := temp.Must(temp.New("main").Funcs(funcMap).ParseGlob(pattern))
		th.E.SetHTMLTemplate(templ)
	}
}

func (th *TmplHelper) LoadView(pattern string) {
	funcMap := temp.FuncMap{}
	funcMap["RenderView"] = th.renderView

	if gin.IsDebugging() {
		th.viewPattern = pattern
		th.viewFuncMap = funcMap
	} else {
		th.views = temp.Must(temp.New("views").Funcs(funcMap).ParseGlob(pattern))
	}
}
