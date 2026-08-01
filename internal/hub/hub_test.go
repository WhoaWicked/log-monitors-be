package hub

import (
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

func TestHubConcurrentAccess(t *testing.T) {
	hub := NewHub()
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hub.Broadcast([]byte("test message"))
		}()
	}
	wg.Wait()
}

func TestHubConcurrentRegisterUnregisterBroadcast(t *testing.T) {
	hub := NewHub()
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn := &websocket.Conn{}
			hub.Register(conn)
			hub.Unregister(conn)
		}()
	}
	wg.Wait()
}
