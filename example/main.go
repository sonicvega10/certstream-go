package main

import (
	logging "github.com/op/go-logging"
	cs "github.com/sysgoblin/certstream-go"
)

var log = logging.MustGetLogger("example")

func main() {
	// Example
	// config := &certstream.Configuration{
	// 	Timeout: 30,
	// 	Server:  "wss://your.certstream.server:1234",
	//	ShowHeartbeats: true,
	// }

	// Declare empty Configuration for default options
	config := &cs.Configuration{}

	stream, errStream := cs.CertStreamEventStream(config)
	for {
		select {
		case jq := <-stream:
			messageType, err := jq.String("message_type")

			if err != nil {
				log.Fatal("Error decoding jq string")
			}

			log.Info("Message type -> ", messageType)
			log.Info("recv: ", jq)

		case err := <-errStream:
			log.Error(err)
		}
	}
}
