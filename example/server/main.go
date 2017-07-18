package main

import (
	"github.com/bysir-zl/hubs/core/net/listener"
	"context"
	"github.com/bysir-zl/bygo/log"
	"github.com/bysir-zl/hubs/core/server"
	"time"
	"github.com/bysir-zl/bjson"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
	"github.com/bysir-zl/sync-chess/example/server/room"
	"github.com/bysir-zl/sync-chess/chess"
)

func main() {
	Run()
}

func Run() {
	l := listener.NewWs()
	ctx := context.Background()

	log.Info("server", "running")
	server.Run(ctx, "127.0.0.1:10010", l, handler)
}

func handler(con conn_wrap.Interface) {
	log.Info("conn")

	var uid = ""
	var room_id = ""

	go func() {
		time.Sleep(5 * time.Second)
		if uid == "" {
			con.Close()
		}
	}()

	for {
		bs, err := con.Read()
		if err != nil {
			break
		}
		bj, err := bjson.New(bs)
		if err != nil {
			continue
		}
		cmd := bj.Pos("Cmd").Int()
		switch cmd {

		case 0:
			// 登录
			uid = bj.Pos("uid").String()
			room_id = bj.Pos("room_id").String()
			con.Subscribe("uid" + uid)

			err := room.JoinRoom(room_id, uid)
			if err != nil {
				// 重连
				if err == chess.ERR_JoinRoomPlayerExist {
					// todo 发送整个房间
					room.SendLastActions(room_id, uid)
				}
				log.Info("joinRoom Err", err)
				break
			}

			log.Info("login", uid)
			log.Info("joinRoom", room_id)
		case 1:
			// 加入房间

		}
		if cmd == 0 {
			con.SetValue("uid", bj.Pos("Body").Int())
		}

		log.Info("read", string(bs))

		con.Write([]byte("SB"))
	}

	log.Info("close", uid)

	if uid != "" {
		err := room.Leave(room_id, uid)
		log.Info("room Leave", err)
	}
}
