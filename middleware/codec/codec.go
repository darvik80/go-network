package codec

type Codec int

const (
	CodecUnknown Codec = iota
	CodecSswDws
	CodecSswPlc
)

func GetCodec(codec string) Codec {
	switch codec {
	case "SSW_DWS":
		return CodecSswDws
	case "SSW_PLC":
		return CodecSswPlc
	}

	return CodecUnknown
}

