package core

type Player interface {
	// 获能进行的动作,应该根据手上的牌判断返回
	// isRounder是否该你出牌
	CanActions(isRounder bool) ActionTypes

	// 阻塞获取玩家操作
	// actions 为需要玩家的动作
	// playerAction 为玩家响应
	RequestAction(types ActionTypes) (playerAction PlayerActionRequest)

	// 所有打牌动作,当用户请求时应该handle
	// playerDe 被操作者
	DoAction(action *PlayerActionRequest, playerDe Player) (response *PlayerActionResponse)

	// 当玩家超时, 或者亮倒时, 执行自动打牌
	// playerDe 被操作者
	DoActionAuto(action *PlayerActionRequest, playerDe Player) (response *PlayerActionResponse)
}

// 玩家动作请求
type PlayerActionRequest struct {
	Types ActionType
	Cards Cards // 动作哪几张牌, 比如亮倒隐藏刻子时有用
	Card  Card  // 动作哪张牌
}

// 玩家动作相应
type PlayerActionResponse struct {
	Err  error
	Card Card // 动作哪张牌
}

func NewActionResponse() *PlayerActionResponse {
	return &PlayerActionResponse{
	}
}
