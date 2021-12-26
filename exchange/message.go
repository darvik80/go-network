package exchange

type StdMessage struct {
	DeviceId *int
	Device   Source
	Id       int64
}

type StdMsg interface {
	Header() *StdMessage
}

func (s *StdMessage) Header() *StdMessage {
	return s
}

type StdBarcodeMessage struct {
	StdMessage
	Barcodes []string
}

type StdDwsReport struct {
	StdBarcodeMessage
	Width  float64
	Weight float64
	Height float64
	Length float64
}

type StdSortRequest struct {
	StdBarcodeMessage
	Destinations []int
}

type StdSortReport struct {
	StdBarcodeMessage
	Destination int
}

type StdDwsSortReport struct {
	StdDwsReport
	Destination int
}

type StdKeepAliveRequest struct {
	StdMessage
}

type StdKeepAliveResponse struct {
	StdMessage
}

type StdHeartbeat struct {
	StdMessage
}
