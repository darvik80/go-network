package ssw

import (
	"darvik80/go-network/exchange"
	"darvik80/go-network/network"
	"fmt"
	"time"
)

func PlcEncoder(ctx network.OutboundContext, msg network.Message) {
	switch m := msg.(type) {
	case exchange.SortReport:
		str := fmt.Sprintf("%d;;%d\r\n", m.Id, m.ChuteId)
		ctx.Write([]byte(str))
	default:
		break
	}
}

func PlcPipeline(p network.Pipeline) network.Pipeline {
	p.AddLast(
		network.NewLogger(),
		network.ReadIdleHandler(time.Second*60),
		network.NewLineBase(),
		PlcEncoder,
	)

	return p
}
