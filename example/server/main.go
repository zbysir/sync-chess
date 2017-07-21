package main

import (
	"github.com/bysir-zl/hubs/core/net/listener"
	"github.com/bysir-zl/bygo/log"
	"time"
	"github.com/bysir-zl/bjson"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
	"github.com/bysir-zl/sync-chess/chess"
	"github.com/bysir-zl/sync-chess/example/chess_i"
	"github.com/bysir-zl/hubs/core/hubs"
	"github.com/bysir-zl/sync-chess/example/server/rooms"
)

func main() {
	l := listener.NewWs()

	log.InfoT("server", "running")
	s := hubs.New("127.0.0.1:10010", l, handler)
	chess_i.Hub = s
	s.Run()
}

func handler(s *hubs.Server, con conn_wrap.Interface) {
	log.InfoT("conn")

	// 玩家id
	var uid = ""
	// 玩家加入的房间
	var rom *rooms.Room

	go func() {
		time.Sleep(5 * time.Second)
		if rom == nil {
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

		if cmd != 0 && rom == nil {
			break
		}

		switch cmd {
		case chess_i.CMD_JoinRoom:
			// 登录
			uid = bj.Pos("Uid").String()
			if uid == "" {
				con.Close()
				break
			}

			s.Subscribe(con, "uid"+uid)
			roomId := bj.Pos("RoomId").String()

			r, err := rooms.FindOrCreateRoom(roomId)
			if err != nil {
				return
			}
			rom = r
			err = rom.JoinRoom(uid)
			if err != nil {
				// 重连
				if err == chess.ERR_JoinRoomPlayerExist {
					rom.SendRoom(uid)
					rom.SendLastActions(uid)
				} else {
					log.Info("joinRoom Err", err)
					break
				}
			}

			log.InfoT("login", uid)
			log.InfoT("joinRoom", roomId)

		case chess_i.CMD_Action:
			// 打牌动作
			action := chess.PlayerActionRequest{}
			err := bj.Pos("Action").Object(&action)
			log.InfoT("action", action, err)
			rom.WriteAction(uid, &action)
		}

		log.InfoT("read", string(bs))

		con.Write([]byte("SB"))
	}

	log.InfoT("close", uid)

	if uid != "" {
		err := rom.Leave(uid)
		s.UnSubscribe(con, "uid"+uid)
		log.Info("room Leave", err)
	}
}
