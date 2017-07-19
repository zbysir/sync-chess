package chess

type Player interface {
	// 获取id
	GetId() (string)
	// 设置手牌
	SetCards(cards Cards)

	// 获能进行的动作,应该根据手上的牌判断返回
	// isRounder是否该你出牌
	// card 别人打的牌,在自己摸牌后这个值0
	CanActions(isRounder bool, card Card) ActionTypes

	// 当玩家超时, 或者亮倒时, 执行自动打牌
	// 应当从actions中选取一个动作
	// card为要吃的牌, 比如胡哪张牌
	RequestActionAuto(actions ActionTypes, card Card) (playerAction *PlayerActionRequest)

	// 所有打牌动作,当用户请求时应该handle
	// playerDe 被操作者(点炮,点杠,碰等)
	DoAction(action *PlayerActionRequest, playerDe Player) (err error)

	// 序列化, 保存玩家状态(手上的牌,碰,杠,胡)
	Marshal() (bs []byte, err error)
	Unmarshal(bs []byte) (err error)
}

