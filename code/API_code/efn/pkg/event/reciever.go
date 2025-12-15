/*
Copyright (C) 2022-2025 Contributors | TIM S.p.A. to CAMARA a Series of LF Projects, LLC

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
package event

import (
	"context"
	"net"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/reciever"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/config"
)

// Receiver starts an HTTP server that delivers incoming CloudEvents to the given handler.
// It's designed to be reused across services by configuring port/path via options.
type Receiver interface {
	Start(fn reciever.Handler) error
}

type receiver struct {
	client cloudevents.Client
}

// NewReceiver creates a CloudEvents HTTP server bound to conf.Address.
// It returns a Receiver that dispatches incoming events to a handler.
func NewReceiver(conf config.API) (Receiver, error) {
	ln, err := net.Listen("tcp", conf.Address)
	if err != nil {
		return nil, err
	}
	protocol, err := cloudevents.NewHTTP(cloudevents.WithListener(ln))
	if err != nil {
		return nil, err
	}
	client, err := cloudevents.NewClient(protocol)
	if err != nil {
		return nil, err
	}

	return &receiver{client: client}, nil
}

// Start runs the server and delivers events to fn until ctx is done.
func (r *receiver) Start(handler reciever.Handler) error {
	return r.client.StartReceiver(context.TODO(), handler.Handle)
}
