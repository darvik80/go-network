package ssw

import (
	"darvik80/go-network/exchange"
	"darvik80/go-network/network"
	"strconv"
	"strings"
	"time"
)

func DwsEncoder(ctx network.InboundContext, msg network.Message) {
	report := string(msg.([]byte))
	parts := strings.Split(report, ";;")
	if len(parts) == 2 {
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			return
		}
		chuteId, err := strconv.Atoi(parts[1])
		if err != nil {
			return
		}
		ctx.HandleRead(exchange.SortReport{Id: id, ChuteId: chuteId})
	}
}

func DwsPipeline(p network.Pipeline) network.Pipeline {
	p.AddLast(
		network.NewLogger(),
		network.ReadIdleHandler(time.Second*60),
		network.NewLineBase(),
		DwsEncoder,
	)

	return p
}