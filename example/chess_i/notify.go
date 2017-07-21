package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
	"encoding/json"
	"github.com/bysir-zl/bygo/log"
	"fmt"
	"github.com/bysir-zl/hubs/core/hubs"
)
var Hub *hubs.Server

func NotifyNeedAction(playerId string, actions chess.ActionTypes) {
	log.InfoT("NeedAction", "%s %+v", playerId, actions)
	s := struct {
		Cmd     int
		Actions chess.ActionTypes
	}{
		Cmd:     CMD_NeedAction,
		Actions: actions,
	}
	bs, _ := json.Marshal(s)

	Hub.SendToTopic(GetTopicUid(playerId), bs, nil)
}

func NotifyActionResponse(playerId string, action *chess.PlayerActionRequest) () {
	if action.ActionFrom == chess.AF_Storage {
		return
	}

	log.InfoT("ResponseAction", "=>", playerId, action)
	s := struct {
		Cmd    int
		Action *chess.PlayerActionRequest
	}{
		Cmd:    CMD_ActionRsp,
		Action: action,
	}
	bs, _ := json.Marshal(s)

	cs := Hub.ConnByTopic(GetTopicUid(playerId))
	if len(cs) == 0 {
		return
	}
	conn := cs[0]
	conn.Write(bs)
}

// 通知玩家来自其他玩家动作
type PlayerActionNotice struct {
	PlayerIdFrom string
	Types        chess.ActionType
	Card         chess.Card  // 动作哪张牌,有部分情况会为空 如摸牌
	Cards        chess.Cards // 动作哪些牌
}

func NotifyFromOtherPlayerAction(playerIdFrom string, playerIdTo []string, action *chess.PlayerActionRequest) () {
	if action.Types == chess.AT_Pass {
		return
	}

	log.InfoT("NotifyFromOtherPlayerAction", "from", playerIdFrom, "to", playerIdTo)

	notice := &PlayerActionNotice{
		Card:         action.Card,
		Cards:        action.Cards,
		Types:        action.Types,
		PlayerIdFrom: playerIdFrom,
	}
	if action.Types == chess.AT_Get {
		notice.Card = 0
	}

	s := struct {
		Cmd    int
		Notice *PlayerActionNotice
	}{
		Cmd:    CMD_ActionFormOther,
		Notice: notice,
	}
	bs, _ := json.Marshal(s)
	for _, playerId := range playerIdTo {
		if playerId == playerIdFrom {
			continue
		}
		Hub.SendToTopic(GetTopicUid(playerId), bs, nil)
	}
}

func NotifyRoom(m *chess.Manager, uid string) {
	playerInfo := make([]interface{}, len(m.Players))
	for i, player := range m.Players {
		p := player.(*Player)
		if p.Id == uid {
			playerInfo[i] = p.InfoSelf()
		} else {
			playerInfo[i] = p.InfoOther()
		}
	}

	s := struct {
		Cmd     int
		Players []interface{}
		CardLen int
	}{
		Cmd:     CMD_RoomInfo,
		Players: playerInfo,
		CardLen: len(m.CardGenerator.GetCardsSurplus()),
	}

	bs, _ := json.Marshal(s)
	Hub.SendToTopic(GetTopicUid(uid), bs, nil)
}

func NotifyPlayerCards(m *chess.Manager, uid string) {
	playerInfo := make([]interface{}, len(m.Players))
	for i, player := range m.Players {
		p := player.(*Player)
		if p.Id == uid {
			playerInfo[i] = p.InfoSelf()
		} else {
			playerInfo[i] = p.InfoOther()
		}
	}

	s := struct {
		Cmd     int
		Players []interface{}
		CardLen int
	}{
		Cmd:     CMD_PlayerCards,
		Players: playerInfo,
		CardLen: len(m.CardGenerator.GetCardsSurplus()),
	}

	bs, _ := json.Marshal(s)
	Hub.SendToTopic(GetTopicUid(uid), bs, nil)
}

func GetTopicUid(uid string) string {
	return fmt.Sprintf("uid%s", uid)
}
