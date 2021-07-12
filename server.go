package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"studyWebsocket/impl"
	"time"
)

var (
	upgrade = websocket.Upgrader{
		//允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func wsHandle(w http.ResponseWriter, r *http.Request)  {
	var(
		wsConn *websocket.Conn
		err error
		data []byte
		conn *impl.Connection
	)
	if wsConn, err = upgrade.Upgrade(w, r, nil); err != nil {
		return
	}
	if conn, err = impl.InitConnection(wsConn); err != nil {
		goto ERR
	}
	go func() {
		var(
			err error
		)
		for {
			if err = conn.WriteMessage([]byte("健康检查")); err != nil {
				return
			}
			time.Sleep(time.Second * 1)
		}
	}()
	for {
		if data, err = conn.ReadMessage(); err != nil {
			goto ERR
		}
		if err = conn.WriteMessage(data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close(err)
}

func main() {
	/*handle2 := func (w http.ResponseWriter, r *http.Request)  {
		w.Write([]byte("hi"))
	}
	http.HandleFunc("/ws", handle2)*/
	http.HandleFunc("/ws", wsHandle)
	_ = http.ListenAndServe("0.0.0.0:7777", nil)
}