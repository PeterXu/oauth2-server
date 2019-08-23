package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	pie "github.com/PeterXu/oauth2-server/proto"
	proto "github.com/golang/protobuf/proto"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 16
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewService(service serviceInfo) {
	hub := gg.Hub
	addr := fmt.Sprintf("%s:%d", service.Host, service.Port)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(hub, w, r)
	})

	log.Println("[core] Service is running at:", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		return
	}
}

func wsHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("wsHandler")
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, maxMessageSize)}
	client.hub.register <- client

	//go client.writePump()
	go client.readPump()
}

/// Client is a middleman between the websocket connection and the hub.

type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Status
	status *pie.ServerStatus
}

func (c *Client) readPump() {
	const TAG string = "readPump"

	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		log.Println(TAG, "reading")
		mt, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Println(TAG, "error: ", err)
			} else {
				log.Println(TAG, "error: ", err)
			}
			break
		}
		//log.Println(TAG, "message=", mt, len(message), string(message))
		if mt == websocket.BinaryMessage {
			req := &pie.ServiceRequest{}
			if err := proto.Unmarshal(message[:], req); err != nil {
				log.Printf("unmarshaling error: %v", err)
				continue
			}
			if req.GetType() == pie.ServiceType_SERVICE_SERVER_STATUS {
				server := req.GetServer()
				if server != nil {
					if server.GetType() == pie.ServerType_SERVER_CAPACITY {
						log.Println(TAG, "server:", server)
						c.status = server
					} else {
						log.Println(TAG, "server type:", server.GetType())
					}
				} else {
					log.Println(TAG, "server is empty")
				}
			} else {
				log.Println(TAG, "unkown service type=", req.GetType())
			}
		}
		//c.hub.broadcast <- message
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

/// Hub for all clients

type ClientStatus struct {
	status *pie.ServerStatus
}

type ClientRequest struct {
	reply chan *ClientInfo
}

type ClientInfo struct {
	ip          string
	domain      string
	tcp_port    uint32
	udp_port    uint32
	connections uint32
}

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Select one available client
	choose chan *ClientRequest
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		choose:     make(chan *ClientRequest),
	}
}

func (h *Hub) getServer() {
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case req := <-h.choose:
			req.reply <- h.getClient()
		}
	}
}

func (h *Hub) getClient() *ClientInfo {
	selected := make(map[int][]*ClientInfo)

	// check connection
	checkConnection := func(status *pie.ServerStatus, idx int, threshold uint32) bool {
		if *status.ConnectionCount < threshold {
			info := &ClientInfo{
				*status.Ip,
				*status.Domain,
				*status.TcpPort,
				*status.UdpPort,
				*status.ConnectionCount,
			}

			var ls []*ClientInfo
			if old, ok := selected[idx]; ok {
				ls = append(old, info)
			} else {
				ls = append(ls, info)
			}
			selected[idx] = ls
			return true
		}
		return false
	}

	thresholds := []uint32{20, 50, 100, 150, 200, 250, 300, 1000}
	for client := range h.clients {
		if client.status != nil {
			for idx := range thresholds {
				if checkConnection(client.status, idx, thresholds[idx]) {
					break
				}
			}
		}
	}

	for idx := range thresholds {
		if ls, ok := selected[idx]; ok {
			var pos int = -1
			var limit uint32 = 1024
			for k := range ls {
				if ls[k].connections < limit {
					limit = ls[k].connections
					pos = k
				}
			}
			if pos >= 0 {
				return ls[pos]
			}
		}
	}

	return nil
}
