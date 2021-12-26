package ssw

import (
	"darvik80/go-network/exchange"
	"darvik80/go-network/network"
	"darvik80/go-network/network/handler"
	"fmt"
	"time"
)

func PlcEncoder(ctx network.OutboundContext, msg network.Message) {
	switch m := msg.(type) {
	case exchange.StdDwsSortReport:
		str := fmt.Sprintf("%d;;%d\r\n", m.DeviceId, m.Destination)
		ctx.Write([]byte(str))
	default:
		break
	}
}

func PlcPipeline(p network.Pipeline) network.Pipeline {
	p.AddLast(
		handler.NewLogger(),
		network.ReadIdleHandler(time.Second*60),
		handler.NewLineBase(),
		PlcEncoder,
	)

	return p
}
