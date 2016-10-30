package web

import (
	"bytes"
	"html/template"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/inimei/backup/log"
)

//实现HTMLRender，主要用于调试时，可以实现动态加载有可以导入funcmap
type htmlDebug struct {
	Glob    string
	FuncMap template.FuncMap
}

func (r htmlDebug) Instance(name string, data interface{}) render.Render {
	return render.HTML{
		Template: r.loadTemplate(),
		Name:     name,
		Data:     data,
	}
}

func (r htmlDebug) loadTemplate() *template.Template {
	if len(r.Glob) > 0 {
		return template.Must(template.New("main").Funcs(r.FuncMap).ParseGlob(r.Glob))
	}
	panic("the HTML debug render was created without files or glob pattern")
}

//tmplHelper 用动态加载模板的辅助参数
type tmplHelper struct {
	e *gin.Engine

	//调试环境
	viewPattern string
	viewFuncMap template.FuncMap

	//生产环境用的
	views *template.Template
}

func (th *tmplHelper) renderView(view string, pipelines map[string]interface{}) (template.HTML, error) {

	log.Debug("render view [%v]", view)

	var t *template.Template
	if gin.IsDebugging() && len(th.viewPattern) > 0 {
		t = template.Must(template.New("views").Funcs(th.viewFuncMap).ParseGlob(th.viewPattern))
	} else {
		t = th.views
	}

	var v *template.Template
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

	return template.HTML(buf.String()), nil
}

func (th *tmplHelper) loadMainTmpl(pattern string) {
	funcMap := template.FuncMap{}
	funcMap["RenderView"] = th.renderView

	if gin.IsDebugging() {
		th.e.HTMLRender = htmlDebug{Glob: pattern, FuncMap: funcMap}
	} else {
		templ := template.Must(template.New("main").Funcs(funcMap).ParseGlob(pattern))
		th.e.SetHTMLTemplate(templ)
	}
}

func (th *tmplHelper) loadView(pattern string) {
	funcMap := template.FuncMap{}
	funcMap["RenderView"] = th.renderView

	if gin.IsDebugging() {
		th.viewPattern = pattern
		th.viewFuncMap = funcMap
	} else {
		th.views = template.Must(template.New("views").Funcs(funcMap).ParseGlob(pattern))
	}
}
