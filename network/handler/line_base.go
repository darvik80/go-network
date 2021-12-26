package handler

import (
	"bufio"
	"bytes"
	"darvik80/go-network/network"
	"io"
)

type line struct {
	line []byte
}

func NewLineBase() network.InboundHandler {
	return &line{}
}

func (c *line) HandleRead(ctx network.InboundContext, msg network.Message) {
	switch data := msg.(type) {
	case []byte:
		c.line = append(c.line, data...)
		for {
			r := bufio.NewReader(bytes.NewReader(c.line))
			l, err := r.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					ctx.Close(err)
				}
				break
			} else {
				if l[len(l)-1] != '\n' {
					break
				}

				eol := 1
				if l[len(l)-2] == '\r' {
					eol++
				}
				ctx.HandleRead(l[:len(l)-2])
				c.line = c.line[len(l):]
			}
		}
	default:
		ctx.HandleRead(msg)
	}
}
