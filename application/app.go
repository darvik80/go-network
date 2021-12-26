package application

import (
	"context"
	"darvik80/go-network/config"
	"darvik80/go-network/database"
	"darvik80/go-network/eventbus"
	"darvik80/go-network/exchange"
	"darvik80/go-network/logging"
	"database/sql"
	"github.com/apache/rocketmq-client-go/v2"
	log "github.com/sirupsen/logrus"
)

type Application interface {
	Db() *sql.DB
	Bus() rocketmq.Producer
	Config() *config.Config
	Exchange() exchange.Exchange

	Setup(s ...func(app Application) (func(), error)) error
	Shutdown()
}

type app struct {
	db       *sql.DB
	eventbus rocketmq.Producer
	cfg      *config.Config
	exchange exchange.Exchange

	destroy []func()

	ctx context.Context
}

func NewApp() (Application, error) {
	logging.Setup()

	var a app
	var err error
	if a.cfg, err = config.ReadConfig(); err != nil && a.cfg == nil {
		log.Error("failed read config ", err)
		return nil, err
	}

	if a.db, err = database.NewDb(a.cfg.DataSource); err != nil {
		log.Error("failed open database ", err)
		return nil, err
	}

	if a.eventbus, err = eventbus.NewProducer(a.cfg.Eventbus); err != nil {
		log.Error("failed open rmq ", err)
		return nil, err
	}

	a.exchange = exchange.NewChanExchange(1024, 8)

	return &a, nil
}

func (a *app) Db() *sql.DB {
	return a.db
}

func (a *app) Bus() rocketmq.Producer {
	return a.eventbus
}

func (a *app) Config() *config.Config {
	return a.cfg
}

func (a *app) Exchange() exchange.Exchange {
	return a.exchange
}

func (a *app) Setup(s ...func(app Application) (func(), error)) error {
	for _, fn := range s {
		if d, err := fn(a); err != nil {
			return err
		} else {
			a.destroy = append(a.destroy, d)
		}

	}

	return nil
}

func (a *app) Shutdown() {
	for _, d := range a.destroy {
		d()
	}
	if a.eventbus != nil {
		a.eventbus.Shutdown()
	}
	if a.db != nil {
		a.db.Close()
	}

	if a.exchange != nil {
		a.exchange.Shutdown()
	}
}
