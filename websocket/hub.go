package websocket

// Hub maintains the set of active clients and fans incoming events out to all
// of them. All mutations to the clients map happen inside the Run goroutine so
// no mutex is needed.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

var GlobalHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run must be started in its own goroutine. It processes register, unregister,
// and broadcast events sequentially so client map access is always single-threaded.
func (h *Hub) Run() {
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
					// Client's send buffer is full — assume it is dead.
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

// Publish serialises an Event and queues it for broadcast to all clients.
func (h *Hub) Publish(event Event) {
	data, err := event.Marshal()
	if err != nil {
		return
	}
	h.broadcast <- data
}
