package signature

import (
	"crypto/md5"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yangsongfwd/backup/log"
)

//map[accessID]secretKey
var secretKeyMap map[string]string

var acceptableContentTypes = []string{"application/json", "application/x-www-form-urlencoded"}

const maxBodyLength int64 = 1024 //1k

func isAcceptableContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	for _, t := range acceptableContentTypes {
		if strings.Index(contentType, t) == 0 {
			return true
		}
	}
	return false
}

func generateKey(secretKey string) string {
	m := md5.New()
	m.Write([]byte(secretKey))
	return fmt.Sprintf("%x", m.Sum(nil))
}

func getMd5(raw []byte) string {
	h := md5.New()
	h.Write(raw)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func sign(cMd5, ct, date, secretKey string) (sign string) {
	h := sha1.New()

	io.WriteString(h, secretKey)
	io.WriteString(h, cMd5)
	io.WriteString(h, ct)
	io.WriteString(h, date)

	sign = fmt.Sprintf("%x", h.Sum(nil))
	return
}

func init() {
	secretKeyMap = make(map[string]string)

	// slave
	secretKeyMap["14024bd5e4cf786a"] = "ritHvADDLjOp19e787ytNg80nUsQ"
}

func GetSecretKey(accessID string) (string, error) {
	key, ok := secretKeyMap[accessID]
	if ok {
		return key, nil
	}
	return "", errors.New("no such accessID:" + accessID)
}

func VerifySignature(authType, accessID, sig string, r *http.Request, w http.ResponseWriter) error {
	secretKey, err := GetSecretKey(accessID)
	if err != nil {
		return err
	}

	if authType != "DDNS-1" {
		return errors.New("unsupport auth type:" + authType)
	}

	contentType := r.Header.Get("Content-Type")
	if !isAcceptableContentType(contentType) {
		return errors.New("unacceptable Content-Type:" + contentType)
	}

	var content []byte
	if r.ContentLength > 0 {
		if r.Body != nil {
			mr := http.MaxBytesReader(w, r.Body, maxBodyLength)
			content, err = ioutil.ReadAll(mr)
			if err != nil {
				log.Debug("http read body failed, %v", err)
				return err
			}

			if r.PostForm == nil {
				if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
					r.PostForm, err = url.ParseQuery(string(content))
				}
			}

		} else {
			log.Warn("r.Body == nil && r.ContentLenght > 0....????")
		}
	} else {
		content = []byte(r.RequestURI)
	}

	contentMD5 := r.Header.Get("Content-MD5")
	validMD5 := getMd5(content)
	if contentMD5 != validMD5 {
		return fmt.Errorf("request md5 not match, Content-MD5:%s, validMD5:%s, content:%s", contentMD5, validMD5, string(content))
	}

	date := r.Header.Get("Date")
	_, err = time.Parse("Mon, 2 Jan 2006 15:04:05 GMT", date)
	if err != nil {
		log.Error("request has not date info or date info format error, date:%s", date)
		return err
	}

	validSig := sign(contentMD5, contentType, date, secretKey)
	if sig != validSig {
		return fmt.Errorf("invalid sig, request sig:%s, valid sig:%s", sig, validSig)
	}

	return nil
}

func SignRequest(req *http.Request, data, accessID, secretKey, ctType string) {
	md5 := getMd5([]byte(data))
	req.Header.Set("Content-Md5", md5)
	date := time.Now().Format("Mon, 2 Jan 2006 15:04:05 GMT")
	req.Header.Set("Date", date)
	// ctType:= "application/json"
	req.Header.Set("Content-Type", ctType)

	sig := sign(md5, ctType, date, secretKey)
	req.Header.Set("Authorization", "DDNS-1:"+accessID+":"+sig)
}
