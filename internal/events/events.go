package events

const (
	CreateUserEvent = "faceit.usersvc.v1.users.create"
	UpdateUserEvent = "faceit.usersvc.v1.users.update"
	DeleteUserEvent = "faceit.usersvc.v1.users.delete"
)

// Here come future dependencies.
type Client struct {
}

func New() *Client {
	return &Client{}
}

func (c *Client) Publish(eventName string, data interface{}) {
	// TODO: do some stuff here.
	//
	// Possible solutions:
	// 1. Publish an event into some pubsub system, Cloud PubSub for instance.
	// 2. Use "webhooks" mechanism: allow services have changes pushed to them via http,
	//    developer puts array of urls in configs/config.yaml, then they are written to db,
	//    and on it event Publish iterates over them and send them data.
	// 3. Use gRPC server streaming, but not with this package.
	//
}
