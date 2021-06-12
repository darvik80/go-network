package codec

type Codec int

const (
	Unknown Codec = iota
	SswDws
	SswPlc
)

func GetCodec(codec string) Codec {
	switch codec {
	case "SSW_DWS":
		return SswDws
	case "SSW_PLC":
		return SswPlc
	}

	return Unknown
}

