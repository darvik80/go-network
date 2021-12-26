package codec

import "strings"

type Codec int

const (
	Unknown Codec = iota
	SswDws
	SswPlc
	SjfScada
)

func GetCodec(codec string) Codec {
	switch strings.ToUpper(codec) {
	case "SSW_DWS":
		return SswDws
	case "SSW_PLC":
		return SswPlc
	case "SJF_SCADA":
		return SjfScada
	}

	return Unknown
}
