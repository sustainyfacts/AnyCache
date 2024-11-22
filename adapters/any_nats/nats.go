/*
Copyright © 2023 The Authors (See AUTHORS file)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package any_nats

import (
	"io"
	"log"

	"github.com/nats-io/nats.go"
	"sustainyfacts.dev/anycache/cache"
)

type NatsBroker struct {
	conn  *nats.Conn
	topic string
}

func NewAdapter(urls, topic string, options ...nats.Option) (cache.MessageBroker, error) {
	// Connect to NATS
	nc, err := nats.Connect(urls, options...)
	if err != nil {
		return nil, err
	}
	version := nc.ConnectedServerVersion()
	log.Printf("Connected to NATS, version %s", version)
	return &NatsBroker{conn: nc, topic: topic}, nil
}

// Implement Cache.MessageBroker
func (b *NatsBroker) Send(msg []byte) error {
	err := b.conn.Publish(b.topic, msg)
	if err != nil {
		return err
	}
	b.conn.Flush() // FIXME Maybe not ideal
	return nil
}

// Implement Cache.MessageBroker
func (b *NatsBroker) Subscribe(messageHandler func(message []byte)) (io.Closer, error) {
	sub, err := b.conn.Subscribe(b.topic, func(msg *nats.Msg) {
		messageHandler(msg.Data)
	})
	if err != nil {
		return nil, err
	}
	return &natsSubscription{sub: sub}, nil
}

type natsSubscription struct {
	sub *nats.Subscription
}

// Implement Closer interface
func (ns *natsSubscription) Close() error {
	return ns.sub.Drain()
}

// To be able to return an anonymous function in Subscribe()
type closerFunc func() error

func (f closerFunc) Close() error {
	return f()
}
