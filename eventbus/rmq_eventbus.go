package eventbus

import (
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	log "github.com/sirupsen/logrus"
)

type RmqConfig struct {
	NameServer string `yaml:"name-server"`
	Topic      string `yaml:"topic"`
}

func NewProducer(cfg RmqConfig) (rocketmq.Producer, error) {
	var p, err = rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{cfg.NameServer})),
		producer.WithRetry(2),
	)

	if err != nil {
		log.WithFields(log.Fields{"module": "bus"}).Errorf("configure failed: %e", err)
		return nil, err
	}

	log.WithFields(log.Fields{"module": "bus"}).Info("created")
	return p, p.Start()
}
