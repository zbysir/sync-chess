package chess

type MessageHandler interface {
	// 通知玩家需要进行操作
	// actions 为需要玩家的动作
	NotifyNeedAction(playerId string, actions ActionTypes)
}
