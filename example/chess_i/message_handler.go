package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
)

type MessageHandler struct{}

func (p *MessageHandler) NotifyNeedAction(playerId string, actions chess.ActionTypes) {
	NotifyNeedAction(playerId, actions)
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}
