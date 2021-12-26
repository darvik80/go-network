package handler

import (
	"context"
	"darvik80/go-network/exchange"
	"encoding/json"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	log "github.com/sirupsen/logrus"
)

type busHandler struct {
	topic    string
	producer rocketmq.Producer
}

func NewEventBusHandler(topic string, producer rocketmq.Producer, ex exchange.Exchange) *busHandler {
	var h = &busHandler{
		topic:    topic,
		producer: producer,
	}

	ex.Subscribe(func(source exchange.Source, msg exchange.StdDwsReport) {
		h.OnDwsMessage(source, msg)
	})

	ex.Subscribe(func(source exchange.Source, msg exchange.StdSortReport) {
		h.OnSortMessage(source, msg)
	})

	ex.Subscribe(func(source exchange.Source, msg exchange.StdDwsSortReport) {
		h.OnDwsSortMessage(source, msg)
	})

	return h
}
func (h *busHandler) send(tag string, message exchange.Message) {
	var data, err = json.Marshal(message)
	if err == nil {
		msg := &primitive.Message{
			Topic: h.topic,
			Body:  data,
		}
		msg.WithTag(tag)

		go h.producer.SendAsync(
			context.Background(), func(ctx context.Context, result *primitive.SendResult, err error) {
				if err != nil {
					log.WithFields(log.Fields{"module": "bus"}).Infof("failed rmq-msg, %e", err)
				}
			},
			msg,
		)
	}
}

func (h *busHandler) OnDwsMessage(source exchange.Source, report exchange.StdDwsReport) {
	h.send("DWS_REPORT", report)
}

func (h *busHandler) OnSortMessage(source exchange.Source, report exchange.StdSortReport) {
	h.send("SORT_REPORT", report)
}

func (h *busHandler) OnDwsSortMessage(source exchange.Source, report exchange.StdDwsSortReport) {
	h.send("DWS_SORT_REPORT", report)
}
