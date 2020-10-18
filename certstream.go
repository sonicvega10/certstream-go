package certstream

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/jsonq"
	"github.com/pkg/errors"
)

type Configuration struct {
	Timeout        time.Duration
	Server         string
	ShowHeartbeats bool
}

func (conf *Configuration) setDefaults() {
	// Construct default values if none provided
	if conf.Timeout == 0 {
		conf.Timeout = time.Duration(15) * time.Second
	}

	if conf.Server == "" {
		conf.Server = "wss://certstream.calidog.io"
	}
}

func CertStreamEventStream(conf *Configuration) (chan jsonq.JsonQuery, chan error) {
	conf.setDefaults()
	outputStream := make(chan jsonq.JsonQuery)
	errStream := make(chan error)

	go func() {
		for {
			c, _, err := websocket.DefaultDialer.Dial(conf.Server, nil)

			if err != nil {
				errStream <- errors.Wrap(err, "Error connecting to certstream! Sleeping a few seconds and reconnecting... ")
				time.Sleep(5 * time.Second)
				continue
			}

			defer c.Close()
			defer close(outputStream)

			done := make(chan struct{})

			go func() {
				ticker := time.NewTicker(conf.Timeout)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						c.WriteMessage(websocket.PingMessage, nil)
					case <-done:
						return
					}
				}
			}()

			for {
				var v interface{}
				c.SetReadDeadline(time.Now().Add(15 * time.Second))
				err = c.ReadJSON(&v)
				if err != nil {
					errStream <- errors.Wrap(err, "Error decoding json frame!")
					c.Close()
					break
				}

				jq := jsonq.NewQuery(v)

				res, err := jq.String("message_type")
				if err != nil {
					errStream <- errors.Wrap(err, "Could not create jq object. Malformed json input recieved. Skipping.")
					continue
				}

				if !conf.ShowHeartbeats && res == "heartbeat" {
					continue
				}

				outputStream <- *jq
			}
			close(done)
		}
	}()

	return outputStream, errStream
}
