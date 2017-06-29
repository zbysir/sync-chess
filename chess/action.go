package chess

type ActionType uint16
type ActionTypes []ActionType

const (
	AT_Get  ActionType = iota + 1 // 摸牌, 服务器不会下发这个命令 而是自动(比如杠牌后)下发通知告知玩家摸了那张牌
	AT_Play                       // 出牌
	AT_Peng                       // 碰
	AT_Gang                       // 杠 杠的牌是上家出的牌, 就是直杠; 是自摸的牌 并且是碰过的就是补杠; 是自摸的牌 并且手上有三张了 就是暗杠;
	AT_Hu                         // 胡的牌是上家出的 就是点炮; 是自摸的就是自摸;
	AT_Pass                       // 过, 可以过 杠,碰,胡
)

func (p ActionType) String() (s string) {
	switch p {
	case AT_Get:
		s = "Get"
	case AT_Play:
		s = "Play"
	case AT_Peng:
		s = "Pong"
	case AT_Gang:
		s = "Gang"
	case AT_Hu:
		s = "Hu"
	case AT_Pass:
		s = "Pass"
	}
	return
}

func (p *ActionTypes) Contain(a ActionType) bool {
	for _, at := range *p {
		if at == a {
			return true
		}
	}
	return false
}

// 玩家动作
type PlayerAction struct {
	ActionType ActionType
	Cards      []uint16 // 动作哪几张牌, 比如亮倒隐藏刻子有用
	Card       uint16   // 动作哪张牌
}
