package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/s2ar/swagat/config"
)

type Service struct {
	Hub *Hub
}

type Messages struct {
	Action string `json:"action"`
	Data   []SymbolData
}

type SymbolData struct {
	Symbol    string  `json:"symbol"`
	LastPrice float32 `json:"lastPrice"`
	Timestamp string  `json:"timestamp"`
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	cfg *config.Configuration
	// Registered clients.
	clients map[*Client]bool

	// who subscribes to what symbol
	symbolSubscription    map[string]map[*Client]bool
	symbolSubscriptionAll map[*Client]bool

	// subscription manager channel(messages from the clients)
	management chan *ClientCommand

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	unregister    chan *Client
	bitmexSymbols chan *SymbolData
}

func newHub(cfg *config.Configuration) *Hub {
	return &Hub{
		cfg:                   cfg,
		management:            make(chan *ClientCommand),
		Register:              make(chan *Client),
		unregister:            make(chan *Client),
		clients:               make(map[*Client]bool),
		bitmexSymbols:         make(chan *SymbolData),
		symbolSubscription:    make(map[string]map[*Client]bool),
		symbolSubscriptionAll: make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	go h.connectBitmex(h.bitmexSymbols)

	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				unsubscribe(h, client)
				close(client.Send)
			}
		case cc := <-h.management:
			if cc.Action == "subscribe" {
				if len(cc.Symbols) > 0 {
					for _, v := range cc.Symbols {
						if _, ok := h.symbolSubscription[v]; ok {
							if _, ok := h.symbolSubscription[v][cc.Client]; !ok {
								h.symbolSubscription[v][cc.Client] = true
							}
						} else {
							h.symbolSubscription[v] = make(map[*Client]bool)
							h.symbolSubscription[v][cc.Client] = true
						}
					}
				} else {
					// сперва удалим все одиночные подписки клиента
					unsubscribeSingleSymbol(h.symbolSubscription, cc.Client)

					// теперь подпишемся на все
					if _, ok := h.symbolSubscriptionAll[cc.Client]; !ok {
						h.symbolSubscriptionAll[cc.Client] = true
					}
				}

			} else if cc.Action == "unsubscribe" {
				unsubscribe(h, cc.Client)
			}

		case bitmexSymbol := <-h.bitmexSymbols:

			bs, err := json.Marshal(bitmexSymbol)
			if err != nil {
				log.Printf("error: %v", err)
				break
			}
			log.Printf("get Symbol %s", bitmexSymbol.Symbol)
			for client := range h.symbolSubscriptionAll {
				select {
				case client.Send <- []byte(bs):
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}

			if cl, ok := h.symbolSubscription[bitmexSymbol.Symbol]; ok {
				for client := range cl {
					select {
					case client.Send <- []byte(bs):
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}

		}
	}
}

func (h *Hub) bitmexSignature(apiSecret, verb, endpoint string, expires int) string {
	message := verb + endpoint + strconv.Itoa(expires)
	hm := hmac.New(sha256.New, []byte(apiSecret))
	// Write Data to it
	hm.Write([]byte(message))
	// Get result and encode as hexadecimal string
	return hex.EncodeToString(hm.Sum(nil))
}

func (h *Hub) connectBitmex(с chan *SymbolData) {
	bitmexCfg := h.cfg.BitmexService

	u := url.URL{
		Scheme: bitmexCfg.Scheme,
		Host:   bitmexCfg.Host,
		Path:   bitmexCfg.Endpoint,
	}
	log.Printf("connecting to %s", u.String())

	// собрать секретный ключ
	expires := int(time.Now().Unix()) + 500
	signature := h.bitmexSignature(bitmexCfg.APISecret, bitmexCfg.Verb, bitmexCfg.Endpoint, expires)

	// Send API Key with signed message.
	request := `{"op": "authKeyExpires", "args": ["` + bitmexCfg.APIKey + `",` + strconv.Itoa(expires) + `, "` + signature + `"]}`

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	err = c.WriteMessage(websocket.TextMessage, []byte(request))
	if err != nil {
		log.Println("write:", err)
		return
	}

	request = `{"op": "subscribe", "args": ["instrument"]}`
	err = c.WriteMessage(websocket.TextMessage, []byte(request))
	if err != nil {
		log.Println("write:", err)
		return
	}

	for {
		for {
			var m Messages
			err := c.ReadJSON(&m)
			if err != nil {
				log.Println("write:", err)
			}

			// allowing only with LastPrice and action update
			if m.Action != "update" || m.Data[0].LastPrice <= 0 {
				continue
			}

			log.Printf("Data %+v", m.Data[0])

			с <- &m.Data[0]
		}
	}
}

func unsubscribeSingleSymbol(ss map[string]map[*Client]bool, c *Client) {
	for s, mc := range ss {
		for cc := range mc {
			if c == cc {
				delete(ss[s], c)
			}
		}
	}
}

func unsubscribe(h *Hub, c *Client) {
	// сперва удалим все одиночные подписки клиента
	unsubscribeSingleSymbol(h.symbolSubscription, c)
	delete(h.symbolSubscriptionAll, c)
}

func New(cfg *config.Configuration) *Service {
	return &Service{
		Hub: newHub(cfg),
	}
}
