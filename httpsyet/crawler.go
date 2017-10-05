package httpsyet

import (
	"io"
	"log"
)

type Crawler struct {
	Sites    []string
	Out      io.Writer
	Logger   *log.Logger
	Depth    int
	External bool
}

func Run(c Crawler) {
}
