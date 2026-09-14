package service

import (
	"log"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/external"
	"github.com/IceWhaleTech/CasaOS-UserService/codegen/message_bus"
	"github.com/IceWhaleTech/CasaOS-UserService/pkg/config"
	"gorm.io/gorm"
)

var MyService Repository

type Repository interface {
	Gateway() external.ManagementService
	User() UserService
	MessageBus() *message_bus.ClientWithResponses
	Event() EventService
}

func NewService(db *gorm.DB, RuntimePath string) Repository {

	// The gateway rewrites its management.url at boot, so on a fresh start the
	// file can briefly hold a stale address (a dead port from the previous run)
	// and the first ping returns a connection error. Retry within a bounded
	// window instead of panicking on the first attempt.
	const retryAttempts = 30
	const retryDelay = 500 * time.Millisecond

	gatewayManagement, err := external.NewManagementService(RuntimePath)
	for attempt := 1; err != nil && attempt <= retryAttempts; attempt++ {
		log.Printf("[service] gateway management service not ready (attempt %d/%d): %s", attempt, retryAttempts, err)
		time.Sleep(retryDelay)
		gatewayManagement, err = external.NewManagementService(RuntimePath)
	}
	if err != nil {
		panic(err)
	}

	return &store{
		gateway: gatewayManagement,
		user:    NewUserService(db),
		event:   NewEventService(db),
	}
}

type store struct {
	gateway external.ManagementService
	user    UserService
	event   EventService
}

func (c *store) Event() EventService {
	return c.event
}
func (c *store) Gateway() external.ManagementService {
	return c.gateway
}

func (c *store) User() UserService {
	return c.user
}
func (c *store) MessageBus() *message_bus.ClientWithResponses {
	client, _ := message_bus.NewClientWithResponses("", func(c *message_bus.Client) error {
		// error will never be returned, as we always want to return a client, even with wrong address,
		// in order to avoid panic.
		//
		// If we don't avoid panic, message bus becomes a hard dependency, which is not what we want.

		messageBusAddress, err := external.GetMessageBusAddress(config.CommonInfo.RuntimePath)
		if err != nil {
			c.Server = "message bus address not found"
			return nil
		}

		c.Server = messageBusAddress
		return nil
	})

	return client
}
