package main

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:9999/agent/chat", nil)
			if err != nil {
				fmt.Printf("[%d] connect failed: %v\n", id, err)
				return
			}
			defer conn.Close()

			conn.WriteMessage(websocket.TextMessage, []byte("hi"))

			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("[%d] read failed: %v\n", id, err)
				return
			}
			fmt.Printf("[%d] received: %s\n", id, string(msg))
		}(i)
	}
	wg.Wait()
}
