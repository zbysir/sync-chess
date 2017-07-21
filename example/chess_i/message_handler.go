package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
)

type MessageHandler struct {
	m *chess.Manager
}

func (p *MessageHandler) Mount(m *chess.Manager) {
	p.m = m
}

func (p *MessageHandler) NotifyNeedAction(playerId string, actions chess.ActionTypes) {
	NotifyNeedAction(playerId, actions)
}

func (p *MessageHandler) OnGameStart() {

}

func (p *MessageHandler) OnGameEnd() {

}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}
