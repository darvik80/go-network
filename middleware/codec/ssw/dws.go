package ssw

import (
	"darvik80/go-network/exchange"
	"darvik80/go-network/network"
	"darvik80/go-network/network/handler"
	"strconv"
	"strings"
	"time"
)

/**
 *  id;;weight;;width;;height;;length;;barcodes;;tray
 */
func DwsDecoder(ctx network.InboundContext, msg network.Message) {
	report := string(msg.([]byte))
	parts := strings.Split(report, ";;")
	if len(parts) == 7 {
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			return
		}

		var report = exchange.StdDwsReport{}
		report.Id = int64(id)
		report.Barcodes = strings.Split(parts[5], ";")
		report.Barcodes = append(report.Barcodes, parts[6])
		if report.Weight, err = strconv.ParseFloat(parts[1], 64); err != nil {
			return
		}
		if report.Width, err = strconv.ParseFloat(parts[2], 64); err != nil {
			return
		}
		if report.Height, err = strconv.ParseFloat(parts[3], 64); err != nil {
			return
		}
		if report.Length, err = strconv.ParseFloat(parts[4], 64); err != nil {
			return
		}

		ctx.HandleRead(report)
	}
}

func DwsPipeline(p network.Pipeline) network.Pipeline {
	p.AddLast(
		handler.NewLogger(),
		network.ReadIdleHandler(time.Second*60),
		handler.NewLineBase(),
		DwsDecoder,
	)

	return p
}
