package network

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

var byte2hex [256]string
var hexpadding [16]string
var byte2char [256]string
var bytepadding [16]string

func init() {
	// Generate the lookup table for byte-to-hex-dump conversion

	for i := 0; i < len(byte2hex); i++ {
		byte2hex[i] = fmt.Sprintf(" %02x", i)
	}

	for i := 0; i < len(hexpadding); i++ {
		str := ""
		for j := 0; j < len(hexpadding)-i; j++ {
			str += "   "
		}
		hexpadding[i] = str
	}

	for i := 0; i < len(byte2char); i++ {
		if i <= 0x1f || i >= 0x7f {
			byte2char[i] = "."
		} else {
			byte2char[i] = string([]byte{byte(i)})
		}
	}

	for i := 0; i < len(bytepadding); i++ {
		str := ""
		for j := 0; j < len(bytepadding)-i; j++ {
			str += " "
		}
		bytepadding[i] = str
	}

}

type logger struct {
	log *log.Entry
}

func NewLogger() *logger {
	return &logger{log.WithFields(log.Fields{"module": "logger"})}
}

func (l *logger) HandleActive(ctx ActiveContext) {
	l.log = log.WithFields(log.Fields{
		"module": "logger",
		"addr":   ctx.Channel().RemoteAddr().String(),
	},
	)
	l.log.Info("active")
	ctx.HandleActive()
}

func (l *logger) HandleInactive(ctx InactiveContext, err error) {
	l.log.Info("inactive")
	ctx.HandleInactive(err)
}

func dump(data []byte) string {
	length := len(data)
	if length == 0 {
		return ""
	}

	b := strings.Builder{}
	b.WriteString("         +-------------------------------------------------+\r\n")
	b.WriteString("         |  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f |\r\n")
	b.WriteString("+--------+-------------------------------------------------+----------------+")

	startIndex := 0
	endIndex := length

	i := 0
	for i = 0; i < endIndex; i++ {
		relIdx := i - startIndex
		relIdxMod16 := relIdx & 15
		if relIdxMod16 == 0 {
			b.WriteString(fmt.Sprintf("\r\n|%08X|", relIdx))
		}

		b.WriteString(byte2hex[data[i]])
		if relIdxMod16 == 15 {
			b.WriteString(" |")
			for j := i - 15; j <= i; j++ {
				b.WriteString(byte2char[data[j]])
			}
			b.WriteString("|")
		}
	}

	if ((i - startIndex) & 15) != 0 {
		remainder := length & 15
		b.WriteString(hexpadding[remainder])
		b.WriteString(" |")
		for j := i - remainder; j < i; j++ {
			b.WriteString(byte2char[data[j]])
		}
		b.WriteString(bytepadding[remainder])
		b.WriteString("|")
	}
	b.WriteString("\r\n+--------+-------------------------------------------------+----------------+")
	b.WriteString("\r\n                                                                             ")

	return b.String()
}

func (l *logger) HandleWrite(ctx OutboundContext, msg Message) {
	switch data := msg.(type) {
	case []byte:
		l.log.Info(fmt.Sprintf("send: (%d) %s\r\n%s\r\n", len(data), data, dump(data)))
	}

	ctx.HandleWrite(msg)
}

func (l *logger) HandleRead(ctx InboundContext, msg Message) {
	switch data := msg.(type) {
	case []byte:
		l.log.Infof("read: (%d) %s\r\n%s\r\n", len(data), data, dump(data))
	}

	ctx.HandleRead(msg)
}
