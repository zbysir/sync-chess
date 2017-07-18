package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
	"encoding/json"
	"github.com/bysir-zl/bygo/log"
	"fmt"
)

func NotifyActionResponse(playerId string, response *chess.PlayerActionResponse) () {
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

func NotifyFromOtherPlayerAction(playerId string, notice *chess.PlayerActionNotice) () {
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

func GetTopicUidInt(uid int) string {
	return fmt.Sprintf("uid%d", uid)
}

func GetTopicUidString(uid string) string {
	return fmt.Sprintf("uid%s", uid)
}
