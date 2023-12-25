package db

import "time"

type Client struct {
	ClientId   int
	ProtocolId []byte
	Created    time.Time
}

type ClientRepository interface {
	Create(client *Client) error
	GetLast() (*Client, error)
}
