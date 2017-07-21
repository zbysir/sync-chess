package rooms

import (
	"github.com/bysir-zl/sync-chess/chess"
	"github.com/bysir-zl/sync-chess/example/chess_i"
)

type Room struct {
	*chess.Manager
}

func (m *Room) Leave(uid string) (err error) {
	err = m.RemovePlayer(uid)
	return
}

func (m *Room) SendLastActions(uid string) (err error) {
	as, ok := m.LastPlayerNeedAction[uid]
	if !ok {
		return
	}
	m.MessageHandler.NotifyNeedAction(uid, as)
	return
}

func (m *Room) WriteAction(uid string, action *chess.PlayerActionRequest) (err error) {
	err = m.WritePlayerAction(uid, action)
	return
}

func (m *Room) SendRoom(uid string) (err error) {
	chess_i.NotifyRoom(m.Manager, uid)
	return
}

func (m *Room) JoinRoom(uid string) ( err error) {
	err = m.AddPlayer(uid)
	if err != nil {
		return
	}
	if len(m.Players) == 3 {
		m.Start()

		for _, uid := range m.Players.Ids() {
			chess_i.NotifyPlayerCards(m.Manager, uid)
		}
	}
	return
}
