package httpsyet

import (
	"errors"
	"io"
	"log"
	"time"
)

type Crawler struct {
	Sites    []string
	Out      io.Writer
	Log      *log.Logger
	Depth    int
	Parallel int
	Delay    time.Duration
}

func (c Crawler) Run() error {
	if err := validate(c); err != nil {
		return err
	}

	return nil
}

func validate(c Crawler) error {
	if len(c.Sites) == 0 {
		return errors.New("no sites given")
	}
	if c.Out == nil {
		return errors.New("no output writer given")
	}
	if c.Log == nil {
		return errors.New("no error logger given")
	}
	if c.Depth < 0 {
		return errors.New("depth cannot be negative")
	}
	if c.Parallel < 0 {
		return errors.New("parallel cannot be negative")
	}
	return nil
}
