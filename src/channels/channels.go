package channels

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type typeWebsocketConnMsg struct {
	messageType int
	data        []byte
}
type typeSafeWsConn struct {
	conn                 *websocket.Conn
	websocketConnMsgChan chan typeWebsocketConnMsg
	doneChan             chan struct{}
}

/*** * * ***/

func newSafeWsConn(conn *websocket.Conn) *typeSafeWsConn {
	var safeWsConn *typeSafeWsConn

	safeWsConn = &typeSafeWsConn{
		conn:                 conn,
		websocketConnMsgChan: make(chan typeWebsocketConnMsg, 100),
		doneChan:             make(chan struct{}),
	}

	go safeWsConn.connWriteChanLoop()

	return safeWsConn
}

func (safeWsConn *typeSafeWsConn) Close() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Close() called on already-closed channel: %v", r)
		}
	}()
	close(safeWsConn.doneChan) // This may panic if already closed

	return safeWsConn.conn.Close()
}

func (safeWsConn *typeSafeWsConn) connWriteChanLoop() {
loop:
	for {
		var websocketConnMsg typeWebsocketConnMsg
		select {
		case <-safeWsConn.doneChan:
			break loop
		case websocketConnMsg = <-safeWsConn.websocketConnMsgChan:
			err := safeWsConn.conn.WriteMessage(websocketConnMsg.messageType, websocketConnMsg.data)
			if err != nil {
				log.Printf("Write error: %v", err)

				safeWsConn.Close()

				break loop
			}
		}
	}
}

// todo : make it safe and rename to safeReadLoop()
//
// func (safeWsConn *typeSafeWsConn) readLoop() {
// loop:
//
//		for {
//			select {
//			case <-safeWsConn.doneChan:
//				break loop
//			default:
//				_, msg, err := safeWsConn.conn.ReadMessage() // blocking
//				if err != nil {
//					log.Printf("Read error: %v", err)
//					safeWsConn.Close()
//					break loop
//				}
//				// echo
//				safeWsConn.safeWriteText(msg)
//			}
//		}
//	}
//

func (safeWsConn *typeSafeWsConn) safeWrite(websocketConnMsg typeWebsocketConnMsg) {
	safeWsConn.websocketConnMsgChan <- websocketConnMsg
}

func (safeWsConn *typeSafeWsConn) safeWriteText(data []byte) {
	var websocketConnMsg typeWebsocketConnMsg

	websocketConnMsg.messageType = websocket.TextMessage
	websocketConnMsg.data = data

	safeWsConn.safeWrite(websocketConnMsg)
}

func (safeWsConn *typeSafeWsConn) safeWriteTextBlueLoop() {
loop:
	for {
		select {
		case <-safeWsConn.doneChan:
			break loop
		default:
			safeWsConn.safeWriteText([]byte("blue"))

			time.Sleep(250 * time.Millisecond)
		}
	}
}

func (safeWsConn *typeSafeWsConn) safeWriteTextRedLoop() {
loop:
	for {
		select {
		case <-safeWsConn.doneChan:
			break loop
		default:
			safeWsConn.safeWriteText([]byte("red"))

			time.Sleep(250 * time.Millisecond)
		}
	}
}

func (safeWsConn *typeSafeWsConn) safeWritePingLoop() {
loop:
	for {
		select {
		case <-safeWsConn.doneChan:
			break loop
		default:
			var websocketConnMsg typeWebsocketConnMsg

			websocketConnMsg.messageType = websocket.PingMessage
			websocketConnMsg.data = []byte{}

			safeWsConn.safeWrite(websocketConnMsg)

			time.Sleep(1000 * time.Millisecond)
		}
	}
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var wsConn *websocket.Conn
	var safeWsConn *typeSafeWsConn

	wsConn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	safeWsConn = newSafeWsConn(wsConn)
	defer safeWsConn.Close()

	/*** * * ***/

	// read
	// go safeWsConn.readLoop()

	// write
	// write : red
	go safeWsConn.safeWriteTextBlueLoop()
	// write : blue
	go safeWsConn.safeWriteTextRedLoop()
	// ping
	safeWsConn.safeWritePingLoop()

	/*** * * ***/

loop:
	for {
		select {
		case <-safeWsConn.doneChan:
			break loop
		default:
			continue
		}
	}
}
