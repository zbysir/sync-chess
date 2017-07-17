package chess

type MessageHandler interface {
	// 通知玩家需要进行操作
	// actions 为需要玩家的动作
	NotifyNeedAction(playerId string, actions ActionTypes)

	// 响应玩家操作
	NotifyActionResponse(playerId string, response *PlayerActionResponse) ()

	// 来自其他玩家的动作
	NotifyFromOtherPlayerAction(playerId string, notice *PlayerActionNotice) ()
}
