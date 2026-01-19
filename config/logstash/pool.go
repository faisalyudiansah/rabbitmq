package logstash

import (
	"net"
	"sync"
	"time"

	logrustash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/sirupsen/logrus"
)

type LogstashConnection struct {
	conn      net.Conn
	address   string
	timeout   time.Duration
	mu        sync.Mutex
	hook      logrus.Hook
	hookAdded bool
}

func (lc *LogstashConnection) connect() error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	conn, err := net.DialTimeout("tcp", lc.address, lc.timeout)
	if err != nil {
		return err
	}

	lc.conn = conn

	lc.hook = logrustash.New(conn, logrustash.DefaultFormatter(logrus.Fields{
		"service": ServiceName,
	}))
	Logger.Hooks.Add(lc.hook)
	lc.hookAdded = true

	log.Infof("Connected to Logstash at %s", lc.address)
	return nil
}

func startHealthCheck(lc *LogstashConnection, interval time.Duration) {
	ticker := time.NewTicker(interval)
	lastHealthy := true

	go func() {
		for range ticker.C {
			testConn, err := net.DialTimeout("tcp", lc.address, 5*time.Second)
			isHealthy := err == nil

			if isHealthy {
				testConn.Close()
				if !lastHealthy {
					log.Infof("Logstash at %s back online, reconnecting...", lc.address)
					if err := lc.reconnect(); err != nil {
						log.Errorf("Reconnect Logstash failed: %v", err)
						isHealthy = false
					} else {
						log.Infof("Reconnected Logstash successfully: %s", lc.address)
					}
				} else {
					log.Infof("Health check Logstash OK: %s", lc.address)
				}
			} else {
				log.Debugf("Health check Logstash failed: %v, reconnecting...", err)
				if err := lc.reconnect(); err != nil {
					log.Errorf("Reconnect Logstash failed: %v", err)
				}
			}

			lastHealthy = isHealthy
		}
	}()
}

func (lc *LogstashConnection) reconnect() error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if lc.conn != nil {
		lc.conn.Close()
	}

	if lc.hookAdded {
		for level := range Logger.Hooks {
			Logger.Hooks[level] = []logrus.Hook{}
		}
		lc.hookAdded = false
	}

	conn, err := net.DialTimeout("tcp", lc.address, lc.timeout)
	if err != nil {
		return err
	}

	lc.conn = conn

	lc.hook = logrustash.New(conn, logrustash.DefaultFormatter(logrus.Fields{
		"service": ServiceName,
	}))
	Logger.Hooks.Add(lc.hook)
	lc.hookAdded = true

	return nil
}
