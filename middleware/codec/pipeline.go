package codec

import (
	"darvik80/go-network/middleware"
	"darvik80/go-network/network"
	"time"
)

func NewPipeline(h network.Handler, codec middleware.Codec) func(p network.Pipeline) network.Pipeline {
	switch codec {
	case middleware.CodecSswDws:
		return func(p network.Pipeline) network.Pipeline {
			p.AddLast(
				network.NewLogger(),
				network.ReadIdleHandler(time.Second*60),
			)

			return p
		}
	case middleware.CodecSswPlc:
		return nil
	}

	return nil
}
