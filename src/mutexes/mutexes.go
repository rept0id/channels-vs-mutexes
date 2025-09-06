package mutexes

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type typeSafeWsConn struct {
	conn   *websocket.Conn
	mu     sync.Mutex
	closed bool
}

func newSafeWsConn(conn *websocket.Conn) *typeSafeWsConn {
	return &typeSafeWsConn{
		conn:   conn,
		closed: false,
	}
}

func (safeWsConn *typeSafeWsConn) Close() error {
	safeWsConn.mu.Lock()
	defer safeWsConn.mu.Unlock()
	if safeWsConn.closed {
		return nil
	}
	safeWsConn.closed = true
	return safeWsConn.conn.Close()
}

func (safeWsConn *typeSafeWsConn) safeWrite(messageType int, data []byte) error {
	safeWsConn.mu.Lock()
	defer safeWsConn.mu.Unlock()

	if safeWsConn.closed {
		return nil
	}

	return safeWsConn.conn.WriteMessage(messageType, data)
}

func (safeWsConn *typeSafeWsConn) safeWriteText(data []byte) error {
	return safeWsConn.safeWrite(websocket.TextMessage, data)
}

func (safeWsConn *typeSafeWsConn) safeWriteTextBlueLoop() {
	for {
		var err error

		if safeWsConn.closed {
			break
		}

		err = safeWsConn.safeWriteText([]byte("blue"))
		if err != nil {
			log.Printf("Write error: %v", err)

			safeWsConn.Close()

			break
		}

		time.Sleep(250 * time.Millisecond)
	}
}

func (safeWsConn *typeSafeWsConn) safeWriteTextRedLoop() {
	for {
		var err error

		if safeWsConn.closed {
			break
		}

		err = safeWsConn.safeWriteText([]byte("red"))
		if err != nil {
			log.Printf("Write error: %v", err)
			safeWsConn.Close()
			return
		}

		time.Sleep(250 * time.Millisecond)
	}
}

func (safeWsConn *typeSafeWsConn) safeWritePingLoop() {
	for {
		var err error

		if safeWsConn.closed {
			break
		}

		err = (func() error {
			var err error

			safeWsConn.mu.Lock()
			defer safeWsConn.mu.Unlock()

			err = safeWsConn.conn.WriteMessage(websocket.PingMessage, []byte{})

			return err
		})()
		if err != nil {
			log.Printf("Ping error: %v", err)

			safeWsConn.Close()

			return
		}

		time.Sleep(1000 * time.Millisecond)
	}
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var wsConn *websocket.Conn
	var safeWsConn *typeSafeWsConn

	// wsConn
	wsConn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)

		return
	}
	// safeWsConn
	safeWsConn = newSafeWsConn(wsConn)
	defer safeWsConn.Close()

	// write
	// write : blue
	go safeWsConn.safeWriteTextBlueLoop()
	// write : red
	go safeWsConn.safeWriteTextRedLoop()

	// ping
	safeWsConn.safeWritePingLoop()

	/*** * * ***/

	for {
		if safeWsConn.closed {
			break
		}
	}
}
