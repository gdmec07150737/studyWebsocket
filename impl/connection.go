package impl

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type Connection struct {
	wsConn *websocket.Conn
	inChan chan []byte
	outChan chan []byte
	closeChan chan byte
	mutex sync.Mutex
	isClosed bool
}

//初始化操作
func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:  wsConn,
		inChan:  make(chan []byte, 1000),
		outChan: make(chan []byte, 1000),
		closeChan: make(chan byte, 1),
	}
	//启动读协程
	go conn.readLoop()
	//启动写协程
	go conn.writeLoop()
	return
}

// 提供对外的读操作
func (conn *Connection) ReadMessage() (data []byte, err error) {
	select {
		case data = <- conn.inChan:
		case <- conn.closeChan:
			err = errors.New("ReadMessage: websocket 已被关闭")
	}
	return
}

//提供对外的写操作
func (conn *Connection) WriteMessage(data []byte) (err error) {
	select {
		case conn.outChan <- data:
		case <- conn.closeChan:
			err = errors.New("SendMessage: websocket 已被关闭")
	}
	return
}

//提供关闭操作
func (conn *Connection) Close(err error) {
	//线程安全的Close,可重入
	_ = conn.wsConn.Close()

	//只执行一次
	conn.mutex.Lock()
	if conn.isClosed != true {
		close(conn.closeChan)
		conn.isClosed = true
	}
	conn.mutex.Unlock()
	fmt.Println(err)
}

//websocket的读操作
func (conn *Connection) readLoop() {
	var(
		data []byte
		err error
	)
	for {
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			goto ERR
		}
		fmt.Println("1:", string(data))
		select {
			case conn.inChan <- data:
			case <- conn.closeChan:
				err = errors.New("readLoop: websocket 已被关闭")
				goto ERR
		}
	}
ERR:
	conn.Close(err)
}

//websocket的写操作
func (conn *Connection) writeLoop() {
	var(
		data []byte
		err error
	)
	for {
		select {
			case data = <- conn.outChan:
			case <- conn.closeChan:
				err = errors.New("writeLoop: websocket 已被关闭")
				goto ERR
		}
		fmt.Println("2:", string(data))
		if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close(err)
}