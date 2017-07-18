package room

import (
	"github.com/bysir-zl/sync-chess/chess"
	"github.com/bysir-zl/sync-chess/example/chess_i"
	"errors"
)

var Managers = map[string]*chess.Manager{}

func JoinRoom(roomId string, uid string) (err error) {
	m, ok := Managers[roomId]
	if !ok {
		err = errors.New("404")
		return
	}

	err = m.AddPlayer(uid)
	if err != nil {
		return
	}
	if len(m.Players) == 3 {
		m.Start()
	}
	return
}

func Leave(roomId, uid string) (err error) {
	m, ok := Managers[roomId]
	if !ok {
		err = errors.New("404")
		return
	}

	err = m.RemovePlayer(uid)
	return
}

func SendLastActions(roomId string, uid string) (err error) {
	m, ok := Managers[roomId]
	if !ok {
		err = errors.New("404")
		return
	}

	as, ok := m.LastPlayerNeedAction[uid]
	if !ok {
		return
	}
	m.MessageHandler.NotifyNeedAction(uid, as)
	return
}


func init() {
	cg := chess_i.NewCardGenerator()
	pl := chess_i.NewPlayerLeader()
	mh := chess_i.NewMessageHandler()
	m := chess.NewManager("1", cg, pl, mh)
	Managers["1"] = m
}
