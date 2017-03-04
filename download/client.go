package download

import (
	"net/http"
	"time"

	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/config"
	"github.com/inimei/ddns/errs"
	"github.com/inimei/grab"
)

type DownloadClent struct {
	resp      *grab.Response
	isRunning bool
	Error     error
	Src       string
	id        int

	c *grab.Client
}

// GetAsync sends a file transfer request and returns a channel to receive the
// file transfer response context.
//
// The Response is sent via the returned channel and the channel closed as soon
// as the HTTP/1.1 GET request has been served; before the file transfer begins.
//
// The Response may then be used to monitor the progress of the file transfer
// while it is in process.
//
// Any error which occurs during the file transfer will be set in the returned
// Response.Error field as soon as the Response.IsComplete method returns true.
//
// GetAsync is a wrapper for DefaultClient.DoAsync.
func (d *DownloadClent) GetAsync(dst, src string) (<-chan *grab.Response, error) {
	// init client and request
	req, err := grab.NewRequest(src)
	if err != nil {
		return nil, err
	}

	req.Filename = dst
	//req.Tag = ""
	req.CreateMissing = true

	// execute async with default client
	return d.c.DoAsync(req), nil
}

func (d *DownloadClent) newClient() *grab.Client {
	c := &grab.Client{}
	c.UserAgent = "ysong-download"

	c.HTTPClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	return c
}

func (d *DownloadClent) start(url string) error {
	if len(url) == 0 {
		log.Error("download error is empty")
		return errs.ErrDownloadUrlError
	}
	d.Src = url
	d.c = d.newClient()

	respch, err := d.GetAsync(config.Data.Download.Dest, url)
	if err != nil {
		log.Error("Error downloading %s: %v", url, err)
		return err
	}

	log.Info("Initializing download...")

	d.resp = <-respch
	d.isRunning = true

	ticker := time.NewTicker(200 * time.Millisecond)
	go func() {
	downloadWhile:
		for {
			select {
			case <-ticker.C:
				if !d.resp.IsComplete() {
					log.DebugLine("Progress %d / %d bytes (%d%%)", d.resp.BytesTransferred(), d.resp.Size, int(100*d.resp.Progress()))
				} else if d.resp.Error != nil {
					log.Error("Error downloading %s: %v", url, d.resp.Error)
					d.Error = d.resp.Error
					break downloadWhile
				}
			}
		}
	}()

	return d.resp.Error
}

func (d *DownloadClent) Stop() {
	if d.c != nil {
		d.c.CancelRequest(d.resp.Request)
	}
}

func (d *DownloadClent) Size() uint64 {
	return d.resp.Size
}

func (d *DownloadClent) BytesTransferred() uint64 {
	return d.resp.BytesTransferred()
}

func (d *DownloadClent) FileName() string {
	return d.resp.Filename
}

func (d *DownloadClent) IsRunning() bool {
	return d.isRunning
}

func (d *DownloadClent) AverageBytesPerSecond() uint64 {
	return uint64(d.resp.AverageBytesPerSecond())
}
