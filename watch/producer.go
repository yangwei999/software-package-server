package main

import (
	kfklib "github.com/opensourceways/kafka-lib/agent"

	"github.com/opensourceways/software-package-server/softwarepkg/domain/message"
)

type producer struct {
	softwarePkgInitialized string
}

func (p *producer) SendPkgInitializedEvent(e message.EventMessage) error {
	body, err := e.Message()
	if err != nil {
		return err
	}

	return kfklib.Publish(p.softwarePkgInitialized, nil, body)
}
