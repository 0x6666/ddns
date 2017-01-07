package container

import "errors"

type ErrorType int

const (
	KeyNotFound ErrorType = iota
)

type ContainerErr struct {
	param string
	t     ErrorType
}

func (e ContainerErr) Error() string {
	switch e.t {
	case KeyNotFound:
		return e.param + " not found"
	}

	return "container error"
}

type Container interface {
	Get(key string) (string, error)
	Set(key string, val string, expire int64) error //expire: secend
	Delete(key string) error
	IsExist(key string) bool
	Count() (int64, error)
}

const (
	CTRedis  = "Redis"
	CTMemery = "Memery"
)

func NewContainer(containerType, cfg string) (Container, error) {
	if containerType == CTRedis {
		c := newRedisContainer()
		if err := c.start(cfg); err != nil {
			return nil, err
		}
		return c, nil
	} else if containerType == CTMemery {
		return newMemContainer(), nil
	} else {
		return nil, errors.New("Not Implement")
	}
}
