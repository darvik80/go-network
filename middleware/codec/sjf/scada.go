package sjf

import (
	"darvik80/go-network/exchange"
	"darvik80/go-network/network"
	"darvik80/go-network/network/handler"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ScadaDecoder
// DWSReport: id|time|1|weight|width|height|length|barcodes
// SortReport: id|time|2|barcodes|destination
func ScadaDecoder(ctx network.InboundContext, msg network.Message) {
	report := string(msg.([]byte))
	parts := strings.Split(report, "|")
	if len(parts) < 3 {
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return
	}

	switch parts[2] {
	case "W1":
		ctx.HandleRead(exchange.StdKeepAliveRequest{
			StdMessage: exchange.StdMessage{
				Id: int64(id),
			},
		})
	case "1":
		if len(parts) == 9 {
			var report = exchange.StdDwsReport{
				StdBarcodeMessage: exchange.StdBarcodeMessage{
					StdMessage: exchange.StdMessage{
						Id: int64(id),
					},
					Barcodes: strings.Split(parts[7], ","),
				},
			}

			if report.Weight, err = strconv.ParseFloat(parts[3], 64); err != nil {
				return
			}
			if report.Width, err = strconv.ParseFloat(parts[4], 64); err != nil {
				return
			}
			if report.Height, err = strconv.ParseFloat(parts[5], 64); err != nil {
				return
			}
			if report.Length, err = strconv.ParseFloat(parts[6], 64); err != nil {
				return
			}

			ctx.HandleRead(report)
		}

	case "2":
		if len(parts) == 5 {
			var report = exchange.StdSortReport{
				StdBarcodeMessage: exchange.StdBarcodeMessage{
					StdMessage: exchange.StdMessage{
						Id: int64(id),
					},
					Barcodes: strings.Split(parts[3], ","),
				},
			}

			if report.Destination, err = strconv.Atoi(parts[3]); err != nil {
				return
			}

			ctx.HandleRead(report)
		}
	}
}

func ScadaEncoder(ctx network.OutboundContext, msg network.Message) {
	var ts = time.Now().Format("20060102150405")
	switch m := msg.(type) {
	case exchange.StdKeepAliveResponse:
		str := fmt.Sprintf("%d|%s|W2\r\n", m.Id, ts)

		ctx.Write([]byte(str))
	case exchange.StdSortRequest:
		var dst []string
		for _, d := range m.Destinations {
			dst = append(dst, strconv.Itoa(d))
		}
		str := fmt.Sprintf("%d|%s|2|%s|%s\r\n", m.Id, ts, strings.Join(m.Barcodes, ","), strings.Join(dst, ","))

		ctx.Write([]byte(str))
	case exchange.StdDwsSortReport:
		str := fmt.Sprintf("%s;;%d\r\n", m.Device.Name(), m.Destination)
		ctx.Write([]byte(str))
	default:
		break
	}
}

func ScadaPipeline(p network.Pipeline) network.Pipeline {
	p.AddLast(
		handler.NewLogger(),
		network.ReadIdleHandler(time.Second*60),
		handler.NewLineBase(),
		ScadaDecoder,
		ScadaEncoder,
	)

	return p
}
