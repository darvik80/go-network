package ssw

import (
	"darvik80/go-network/exchange"
	"darvik80/go-network/network"
	"strconv"
	"strings"
	"time"
)

func DwsDecoder(ctx network.InboundContext, msg network.Message) {
	report := string(msg.([]byte))
	parts := strings.Split(report, ";;")
	if len(parts) == 2 {
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			return
		}
		ctx.HandleRead(exchange.DwsReport{Id: id, Width: 10.0, Weight: 20.0, Height: 30.0, Length: 40.0})
	}
}

func DwsPipeline(p network.Pipeline) network.Pipeline {
	p.AddLast(
		network.NewLogger(),
		network.ReadIdleHandler(time.Second*60),
		network.NewLineBase(),
		DwsDecoder,
	)

	return p
}
