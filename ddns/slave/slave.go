package slave

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"fmt"

	"sync"

	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/config"
	"github.com/inimei/ddns/data"
	"github.com/inimei/ddns/data/model"
	"github.com/inimei/ddns/errs"
	"github.com/inimei/ddns/web"
	"github.com/inimei/ddns/web/signature"
)

type SlaveServer struct {
	db data.IDatabase

	sync.RWMutex
	updating bool
}

type recode struct {
	Id      int64  `json:"id"`
	Type    int    `json:"type"`
	Name    string `json:"name"`
	Dynamic bool   `json:"dynamic"`
	Value   string `json:"value"`
	Ttl     int    `json:"ttl"`
	Key     string `json:"key"`
}

func (ss *SlaveServer) Init(db data.IDatabase) error {

	if db == nil {
		log.Error(errs.ErrInvalidParam.Error())
		return errs.ErrInvalidParam
	}

	if len(config.Data.Slave.MasterHost) == 0 {
		log.Error("master host can't be empty")
		return errors.New("master host can't be empty")
	}

	if len(config.Data.Slave.Accesskey) == 0 || len(config.Data.Slave.SecretKey) == 0 {
		log.Error("accesskey & secretKey  host can't be empty")
		return errors.New("accesskey & secretKey  host can't be empty")
	}

	ss.db = db
	ss.updating = false

	return nil
}

func (ss *SlaveServer) IsUpdating() bool {
	ss.RLock()
	defer ss.RUnlock()
	return ss.updating
}

func (ss *SlaveServer) setStatus(b bool) {
	ss.Lock()
	defer ss.Unlock()
	ss.updating = b
}

func (ss *SlaveServer) Start() {
	ticker := time.NewTicker(time.Second * time.Duration(config.Data.Slave.UpdateTime))

	go ss.checkUpdate(true)

	go func() {
		for {
			ss.checkUpdate(false)
			<-ticker.C
		}
	}()
}

func (ss *SlaveServer) checkUpdate(force bool) {

	if ss.IsUpdating() {
		return
	}

	ss.setStatus(true)
	defer ss.setStatus(false)

	v, err := ss.getVersion()
	if err != nil {
		log.Error(err.Error())
		return
	}

	if v.SchemaVersion != model.CurrentVersion {
		log.Error("schema version not match, need update...")
		return
	}

	if !force && v.DataVersion == ss.db.GetVersion() {
		log.Debug("the same data version....")
		return
	}

	recodes, err := ss.getRecodes()
	if err != nil {
		log.Error(err.Error())
		return
	}

	if len(recodes) == 0 {
		return
	}

	db, err := ss.db.BeginTransaction()
	if err != nil {
		log.Error(err.Error())
		return
	}

	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()

	err = db.ClearRecodes(true)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for _, r := range recodes {
		_r := model.Recode{}
		_r.Dynamic = r.Dynamic
		_r.RecodeValue = r.Value
		_r.RecordName = r.Name
		_r.RecordType = r.Type
		_r.TTL = r.Ttl
		_r.Synced = true
		if len(r.Key) > 0 {
			_r.UpdateKey.String = r.Key
			_r.UpdateKey.Valid = true
		}
		if _, err = db.CreateRecode(&_r); err != nil {
			log.Error(err.Error())
			return
		}
	}

	//jast ss.db
	ss.db.SetVersion(v.DataVersion)
}

func (ss *SlaveServer) getRecodes() ([]recode, error) {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
		},
	}

	url := "/api/recodes"
	req, err := http.NewRequest("GET", config.Data.Slave.MasterHost+url, strings.NewReader(""))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	signature.SignRequest(req, url, config.Data.Slave.Accesskey, config.Data.Slave.SecretKey, "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)

	var data = struct {
		Code    string   `json:"code"`
		Msg     string   `json:"msg"`
		Recodes []recode `json:"recodes"`
	}{}

	err = json.Unmarshal(bodyData, &data)
	if err != nil {
		log.Error("Unmarshal recodes failed: %v", err)
		return nil, err
	}

	if data.Code != web.CodeOK {
		err = fmt.Errorf("get recodes failed: %v, msg: %v", data.Code, data.Msg)
		log.Error(err.Error())
		return nil, err
	}

	return data.Recodes, nil
}

func (ss *SlaveServer) getVersion() (*model.Version, error) {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
		},
	}

	url := "/api/dataversion"
	req, err := http.NewRequest("GET", config.Data.Slave.MasterHost+url, strings.NewReader(""))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	signature.SignRequest(req, url, config.Data.Slave.Accesskey, config.Data.Slave.SecretKey, "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)

	var version = struct {
		Code    string        `json:"code"`
		Msg     string        `json:"msg"`
		Version model.Version `json:"version"`
	}{}

	err = json.Unmarshal(bodyData, &version)
	if err != nil {
		log.Error("Unmarshal version failed: %v", err.Error())
		return nil, err
	}

	if version.Code != web.CodeOK {
		err = fmt.Errorf("get version failed: %v, msg: %v", version.Code, version.Msg)
		log.Error(err.Error())
		return nil, err
	}

	return &version.Version, nil
}
