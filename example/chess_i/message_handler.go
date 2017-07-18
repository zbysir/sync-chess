package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
	"github.com/bysir-zl/bygo/log"
	"encoding/json"
	"github.com/bysir-zl/hubs/core/server"
)

var connManager = server.GetConnManager()

type MessageHandler struct{}

func (p *MessageHandler) NotifyNeedAction(playerId string, actions chess.ActionTypes) {
	log.Info("NeedAction", "%s %+v", playerId, actions)
	s := struct {
		Cmd     int `json:"cmd"`
		Actions chess.ActionTypes `json:"actions"`
	}{
		Cmd:     100,
		Actions: actions,
	}
	bs, _ := json.Marshal(s)

	cs := connManager.ConnByTopic(GetTopicUidString(playerId))
	if len(cs) == 0 {
		return
	}
	conn := cs[0]
	conn.Write(bs)
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}
