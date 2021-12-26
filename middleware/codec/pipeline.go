package codec

import (
	"darvik80/go-network/middleware/codec/sjf"
	"darvik80/go-network/middleware/codec/ssw"
	"darvik80/go-network/network"
)

func NewPipelineFactory(codec Codec) network.PipelineFactory {
	switch codec {
	case SswDws:
		return network.PipelineFactoryFunc(func(p network.Pipeline) network.Pipeline {
			return ssw.DwsPipeline(p)
		})
	case SswPlc:
		return network.PipelineFactoryFunc(func(p network.Pipeline) network.Pipeline {
			return ssw.PlcPipeline(p)
		})
	case SjfScada:
		return network.PipelineFactoryFunc(func(p network.Pipeline) network.Pipeline {
			return sjf.ScadaPipeline(p)
		})
	}

	return nil
}
