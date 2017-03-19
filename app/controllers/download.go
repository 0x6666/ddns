package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yangsongfwd/backup/log"
	"github.com/yangsongfwd/ddns/config"
	"github.com/yangsongfwd/ddns/download"
	"github.com/yangsongfwd/ddns/server"
)

type Downloader struct {
	ContrllerBase
}

func (d Downloader) d() *download.DownloadMgr {
	idload := server.Server.GetGlobalData("downloadMgr")
	if idload == nil {
		log.Error("get downloadMgr error")
		return nil
	}
	return idload.(*download.DownloadMgr)
}

// GetDownloads -> [GET] :/downloads?r=true&offset=10&limit=0
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
//			"Transferred": Transferred,
//			"bytesPerSecond": bytesPerSecond
//		]
//	}
//
// 失败返回值
//		code: xxx
//
func (d Downloader) GetDownloads() {
	if !d.IsJsonReqest() {
		parameters := d.getTemplateParameter()
		parameters["BreadcrumbSecs"] = SectionItems{
			&SectionItem{"Download", PDownloads},
		}

		parameters["View"] = "downloads_view"
		d.HTML(http.StatusOK, parameters)
		return
	}

	files := map[string]bool{}

	//userid, _ := sessions.GetUserID(c.Request)
	tasksArray := []JsonMap{}
	tasks := d.d().Tasks()
	if len(tasks) != 0 {
		for _, t := range tasks {
			files[t.Dest] = true
			tasksArray = append(tasksArray, JsonMap{
				"id":             t.Id,
				"name":           filepath.Base(t.Dest),
				"src":            t.Src,
				"dest":           strings.Replace(t.Dest, config.Data.Download.Dest, PFiles, 1),
				"size":           t.Size,
				"transferred":    t.BytesTransferred,
				"bytesPerSecond": t.AverageBytesPerSecond,
			})
		}
	}

	filepath.Walk(config.Data.Download.Dest,
		func(path string, f os.FileInfo, err error) error {
			if b := files[path]; b {
				return nil
			}
			if f == nil {
				log.Error("walk file error: %v", err)
				return err
			}
			if f.IsDir() {
				return nil
			}
			tasksArray = append(tasksArray, JsonMap{
				"id":             0,
				"name":           filepath.Base(path),
				"src":            "",
				"dest":           strings.Replace(path, config.Data.Download.Dest, PFiles, 1),
				"size":           f.Size(),
				"transferred":    f.Size(),
				"bytesPerSecond": 0,
			})
			return nil
		})

	d.JSON(http.StatusOK, JsonMap{"code": CodeOK, "tasks": tasksArray})
}

// StartDownloads -> [POST] :/downloads?r=true&offset=10&limit=0
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
func (d Downloader) StartDownloads() {

	//userid, _ := sessions.GetUserID(c.Request)
	url := strings.ToLower(d.C.PostForm("url"))
	if url == "" {
		log.Error("url is empty")
		d.JSON(http.StatusOK, JsonMap{"code": CodeInvalidURL, "msg": "url is empty"})
		return
	}

	id, err := d.d().Download(url)
	if err != nil {
		log.Error(err.Error())
		d.JSON(http.StatusOK, JsonMap{"code": CodeUnknowError, "msg": err.Error()})
		return
	}

	d.JSON(http.StatusOK, JsonMap{"code": CodeOK, "id": id})
}
