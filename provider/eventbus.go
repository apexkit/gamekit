package provider

import (
	"fmt"

	gkconf "github.com/apexkit/gamekit/conf"
	"github.com/apexkit/gamekit/eventbus"
	"github.com/google/wire"
)

// EventbusProviderSet is deprecated: use gkbootstrap.ProviderSet with WithNats in main.
var EventbusProviderSet = wire.NewSet(NewEventBus)

// NewEventBus opens the configured eventbus and verifies connectivity when applicable.
func NewEventBus(data *gkconf.Data) (eventbus.EventBus, func(), error) {
	if data == nil || data.Eventbus == nil {
		return nil, nil, fmt.Errorf("eventbus: data.eventbus is nil")
	}
	eb := data.Eventbus
	switch eb.GetType() {
	case "nats":
		bus, err := eventbus.NewConnectedNatsBus(eb.GetUrl())
		if err != nil {
			return nil, nil, err
		}
		return bus, func() { _ = bus.Close() }, nil
	case "kafka":
		bus := eventbus.NewKafkaBus(eb.GetUrl())
		return bus, func() { _ = bus.Close() }, nil
	default:
		return nil, nil, fmt.Errorf("eventbus: unsupported type %q", eb.GetType())
	}
}
