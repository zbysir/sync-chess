package chess

import (
	"github.com/bysir-zl/sync-chess/core"
	"github.com/bysir-zl/bygo/log"
)

type MessageHandler struct {

}

func (p *MessageHandler) NotifyNeedAction(playerId string, actions core.ActionTypes) {
	log.Info("NeedAction", "%s %+v", playerId, actions)

}

func (p *MessageHandler) NotifyActionResponse(playerId string, response *core.PlayerActionResponse) () {
	log.Info("ResponseAction", "=>", playerId, response)
}

func (p *MessageHandler) NotifyFromOtherPlayerAction(playerId string, notice *core.PlayerActionNotice) () {
	log.Info("NotifyFromOtherPlayerAction", "=>", playerId, notice)
}

func NewMessageHandler()*MessageHandler{
	return &MessageHandler{

	}
}