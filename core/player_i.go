package core

type Player interface {
	Peng(player Player, card Card) (err error) // 只能碰别人p的牌card
	GangDian(p Player, card Card)              // 点杠
	GangBu(card Card)                          // 补杠
	GangZi(card Card)                          // 自杠
	HuZiMo(card Card)                          // 自摸
	HuDian(p Player, c Card)                   // 点炮
	HuQiangGang(p Player, c Card)              // 抢杠胡

	// 获能进行的动作,应该根据手上的牌判断返回, isRounder是否该你出牌
	CanActions(isRounder bool) ActionTypes

	// 阻塞获取玩家操作
	// actions 为需要玩家的动作
	// playerAction 为玩家响应
	RequestAction(actions ActionTypes) (playerAction PlayerAction)
}

// 玩家动作
type PlayerAction struct {
	ActionType ActionType
	Cards      []uint16 // 动作哪几张牌, 比如亮倒隐藏刻子时有用
	Card       uint16   // 动作哪张牌
}
