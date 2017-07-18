package main

import (
	"github.com/bysir-zl/hubs/core/net/listener"
	"context"
	"github.com/bysir-zl/bygo/log"
	"github.com/bysir-zl/hubs/core/server"
	"time"
	"github.com/bysir-zl/bjson"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
	"github.com/bysir-zl/sync-chess/example/server/rooms"
	"github.com/bysir-zl/sync-chess/chess"
	"github.com/bysir-zl/sync-chess/example/chess_i"
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

		if cmd != 0 && uid == "" {
			break
		}

		switch cmd {
		case chess_i.CMD_JoinRoom:
			// 登录
			uid = bj.Pos("Uid").String()
			room_id = bj.Pos("RoomId").String()
			con.Subscribe("uid" + uid)

			err := rooms.JoinRoom(room_id, uid)
			if err != nil {
				// 重连
				if err == chess.ERR_JoinRoomPlayerExist {
					rooms.SendRoom(room_id, uid)
					rooms.SendLastActions(room_id, uid)
				} else {
					log.Info("joinRoom Err", err)
					break
				}
			}

			log.Info("login", uid)
			log.Info("joinRoom", room_id)

		case chess_i.CMD_Action:
			// 打牌动作
			action := chess.PlayerActionRequest{}
			err := bj.Pos("Action").Object(&action)
			log.Info("action", action, err)
			rooms.WriteAction(room_id, uid, &action)
		}
		if cmd == 0 {
			con.SetValue("uid", bj.Pos("Body").Int())
		}

		log.Info("read", string(bs))

		con.Write([]byte("SB"))
	}

	log.Info("close", uid)

	if uid != "" {
		err := rooms.Leave(room_id, uid)
		log.Info("room Leave", err)
	}
}
