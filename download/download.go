package download

import (
	"strings"
)

type Task struct {
	Size                  uint64
	BytesTransferred      uint64
	AverageBytesPerSecond uint64
	Src                   string
	Dest                  string
	Id                    int
}

var clientIDBase = 0

func NewDownloadMgr() *DownloadMgr {
	return &DownloadMgr{[]*DownloadClent{}}
}

type DownloadMgr struct {
	dcs []*DownloadClent
}

func (d *DownloadMgr) Start() {
}

func (d *DownloadMgr) Download(url string) (int, error) {
	url = strings.ToLower(url)

	/*if _, b := d.dcs[url]; b {
		return 0, errs.ErrTaskAlreadyExist
	}*/

	clientIDBase = clientIDBase + 1

	c := DownloadClent{}
	c.id = clientIDBase
	err := c.start(url)

	d.dcs = append(d.dcs, &c)

	return c.id, err
}

func (d *DownloadMgr) Tasks() []*Task {
	tasks := []*Task{}

	for _, t := range d.dcs {
		task := Task{}
		task.Size = t.Size()
		task.BytesTransferred = t.BytesTransferred()
		task.Src = t.Src
		task.Dest = t.FileName()
		task.Id = t.id
		task.AverageBytesPerSecond = t.AverageBytesPerSecond()
		tasks = append(tasks, &task)
	}

	return tasks
}
