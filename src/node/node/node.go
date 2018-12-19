package node

import (
	"sync"
	"time"

	router "router/client"
	"storage"
)

// Config stores configuration for a Node service.
//
// Config -- содержит конфигурацию Node.
type Config struct {
	// Addr is an address to listen at.
	// Addr -- слушающий адрес Node.
	Addr storage.ServiceAddr
	// Router is an address of Router service.
	// Router -- адрес Router service.
	Router storage.ServiceAddr
	// Heartbeat is a time interval between heartbeats.
	// Heartbeat -- интервал между двумя heartbeats.
	Heartbeat time.Duration

	// Client specifies client for Router.
	// Client -- клиент для Router.
	Client router.Client `yaml:"-"`
}

// Node is a Node service.
type Node struct {
	lock sync.Mutex
	cfg  Config
	test chan struct{}
	data map[storage.RecordID][]byte
}

// New creates a new Node with a given cfg.
//
// New создает новый Node с данным cfg.
func New(cfg Config) *Node {
	return &Node{lock: sync.Mutex{}, cfg: cfg, test: make(chan struct{}), data: make(map[storage.RecordID][]byte)}
}

// Heartbeats runs heartbeats from node to a router
// each time interval set by cfg.Heartbeat.
//
// Heartbeats запускает отправку heartbeats от node к router
// через каждый интервал времени, заданный в cfg.Heartbeat.
func (node *Node) Heartbeats() {
	go startHb(node)
}

func startHb(node *Node) {
	for {
		select {
		case <-node.test:
			return
		default:
			node.cfg.Client.Heartbeat(node.cfg.Router, node.cfg.Addr)
			time.Sleep(node.cfg.Heartbeat)
		}
	}
}

// Stop stops heartbeats
//
// Stop останавливает отправку heartbeats.
func (node *Node) Stop() {
	node.test <- struct{}{}
}

// Put an item to the node if an item for the given key doesn't exist.
// Returns the storage.ErrRecordExists error otherwise.
//
// Put -- добавить запись в node, если запись для данного ключа
// не существует. Иначе вернуть ошибку storage.ErrRecordExists.
func (node *Node) Put(k storage.RecordID, d []byte) error {
	node.lock.Lock()
	defer node.lock.Unlock()
	if _, ok := node.data[k]; !ok {
		node.data[k] = d
		return nil
	}
	return storage.ErrRecordExists
}

// Del an item from the node if an item exists for the given key.
// Returns the storage.ErrRecordNotFound error otherwise.
//
// Del -- удалить запись из node, если запись для данного ключа
// существует. Иначе вернуть ошибку storage.ErrRecordNotFound.
func (node *Node) Del(k storage.RecordID) error {
	node.lock.Lock()
	defer node.lock.Unlock()
	if _, ok := node.data[k]; ok {
		delete(node.data, k)
		return nil
	}
	return storage.ErrRecordNotFound
}

// Get an item from the node if an item exists for the given key.
// Returns the storage.ErrRecordNotFound error otherwise.
//
// Get -- получить запись из node, если запись для данного ключа
// существует. Иначе вернуть ошибку storage.ErrRecordNotFound.
func (node *Node) Get(k storage.RecordID) ([]byte, error) {
	node.lock.Lock()
	defer node.lock.Unlock()
	if d, ok := node.data[k]; ok {
		return d, nil
	}
	return nil, storage.ErrRecordNotFound
}
