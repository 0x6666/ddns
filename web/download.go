package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/config"
)

// getDownloads -> [GET] :/downloads?r=true&offset=10&limit=0
//
// Ret Code:[200]
//
// [request=json] 获取指定区间的记录数
// [request=html] 获取下载列表页面
//
// 成功返回值
//	{
//		"code": "OK"
//		"tasks" : [
//			"id": id,
//			"name": "name"
//			"src": "src",
//			"dest": "dest",
//			"size":	size,
//			"Transferred": Transferred
//		]
//	}
//
// 失败返回值
//		code: xxx
//
func (h *handler) getDownloads(c *gin.Context) {
	if t := requestType(c.Request); t != MIMEJSON {
		parameters := h.getTemplateParameter(c)
		parameters["BreadcrumbSecs"] = SectionItems{
			&SectionItem{"Download", pDownloads},
		}

		parameters["View"] = "downloads_view"
		h.HTML(c, http.StatusOK, parameters)
		return
	}

	//userid, _ := sessions.GetUserID(c.Request)
	tasksArray := []JsonMap{}
	tasks := h.ws.d.Tasks()
	if len(tasks) != 0 {
		for _, t := range tasks {
			tasksArray = append(tasksArray, JsonMap{
				"id":          t.Id,
				"name":        filepath.Base(t.Dest),
				"src":         t.Src,
				"dest":        strings.Replace(t.Dest, config.Data.Download.Dest, pFiles, 1),
				"size":        t.Size,
				"transferred": t.BytesTransferred,
			})
		}
	}

	filepath.Walk(config.Data.Download.Dest,
		func(path string, f os.FileInfo, err error) error {
			if f == nil {
				log.Error("walk file error: %v", err)
				return err
			}
			if f.IsDir() {
				return nil
			}
			tasksArray = append(tasksArray, JsonMap{
				"id":          0,
				"name":        filepath.Base(path),
				"src":         "",
				"dest":        strings.Replace(path, config.Data.Download.Dest, pFiles, 1),
				"size":        f.Size(),
				"transferred": f.Size(),
			})
			return nil
		})

	c.JSON(http.StatusOK, JsonMap{"code": CodeOK, "tasks": tasksArray})
}

// startDownloads -> [POST] :/downloads?r=true&offset=10&limit=0
//
// Ret Code:[200]
//
// [request=json] 开始一个下载任务
//
// 成功返回值
//	{
//		"code": "OK",
//		"id":	id
//	}
//
// 失败返回值
//		code: xxx
//
func (h *handler) startDownloads(c *gin.Context) {

	//userid, _ := sessions.GetUserID(c.Request)
	url := strings.ToLower(c.PostForm("url"))
	if url == "" {
		log.Error("url is empty")
		c.JSON(http.StatusOK, JsonMap{"code": CodeInvalidURL, "msg": "url is empty"})
		return
	}

	id, err := h.ws.d.Download(url)
	if err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, JsonMap{"code": CodeUnknowError, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, JsonMap{"code": CodeOK, "id": id})
}
