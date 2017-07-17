package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
	"github.com/bysir-zl/bygo/log"
	"fmt"
	"encoding/json"
	"github.com/bysir-zl/hubs/core/server"
)

type MessageHandler struct{}

var connManager = server.GetConnManager()

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

func (p *MessageHandler) NotifyActionResponse(playerId string, response *chess.PlayerActionResponse) () {
	log.Info("ResponseAction", "=>", playerId, response)
	s := struct {
		Cmd      int `json:"cmd"`
		Response *chess.PlayerActionResponse `json:"response"`
	}{
		Cmd:      101,
		Response: response,
	}
	bs, _ := json.Marshal(s)

	cs := connManager.ConnByTopic(GetTopicUidString(playerId))
	if len(cs) == 0 {
		return
	}
	conn := cs[0]
	conn.Write(bs)
}

func (p *MessageHandler) NotifyFromOtherPlayerAction(playerId string, notice *chess.PlayerActionNotice) () {
	log.Info("NotifyFromOtherPlayerAction", "=>", playerId, notice)

	s := struct {
		Cmd    int `json:"cmd"`
		Notice *chess.PlayerActionNotice `json:"notice"`
	}{
		Cmd:    101,
		Notice: notice,
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
	return &MessageHandler{

	}
}

func GetTopicUidInt(uid int) string {
	return fmt.Sprintf("uid%d", uid)
}

func GetTopicUidString(uid string) string {
	return fmt.Sprintf("uid%s", uid)
}
