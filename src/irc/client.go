package irc

import (
	"fmt"
	"net"
)

type Client struct {
	// communication
	conn net.Conn
	send chan<- Reply
	recv <-chan string
	// basic info
	username   string
	realname   string
	hostname   string
	nick       string
	serverPass bool
	// modes
	away          bool
	registered    bool
	invisible     bool
	wallOps       bool
	restricted    bool
	operator      bool
	localOperator bool
	// relations
	server   *Server
	channels ChannelSet
}

type ClientSet map[*Client]bool

func NewClient(server *Server, conn net.Conn) *Client {
	client := &Client{
		channels: make(ChannelSet),
		conn:     conn,
		hostname: LookupHostname(conn.RemoteAddr()),
		recv:     StringReadChan(conn),
		server:   server,
	}
	client.SetReplyToStringChan()
	return client
}

func (c *Client) SetReplyToStringChan() {
	send := make(chan Reply)
	write := StringWriteChan(c.conn)
	go func() {
		for reply := range send {
			write <- reply.String(c)
		}
	}()
	c.send = send
}

// Adapt `chan string` to a `chan Message`.
func (c *Client) Communicate() {
	for str := range c.recv {
		m, err := ParseMessage(str)
		if err != nil {
			// TODO handle error
			return
		}
		c.server.recv <- &ClientMessage{c, m}
	}
}

func (c *Client) Nick() string {
	if c.nick != "" {
		return c.nick
	}
	return "*"
}

func (c *Client) UModeString() string {
	if c.invisible {
		return "+i"
	}
	return ""
}

func (c *Client) HasNick() bool {
	return c.nick != ""
}

func (c *Client) HasUser() bool {
	return c.username != ""
}

func (c *Client) UserHost() string {
	return fmt.Sprintf("%s!%s@%s", c.nick, c.username, c.hostname)
}

func (c *Client) Id() string {
	return c.UserHost()
}
