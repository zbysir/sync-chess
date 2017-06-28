package core

import "time"

type ActionType uint16

const (
	AT_Get  ActionType = iota + 1 // 摸牌, 服务器不会下发这个命令 而是自动(比如杠牌后)下发通知告知玩家摸了那张牌
	AT_Play                       // 出牌
	AT_Pong                       // 碰
	AT_Gang                       // 杠 杠的牌是上家出的牌, 就是直杠; 是自摸的牌 并且是碰过的就是补杠; 是自摸的牌 并且手上有三张了 就是暗杠;
	AT_Hu                         // 胡的牌是上家出的 就是点炮; 是自摸的就是自摸;
	AT_Pass                       // 过, 可以过 杠,碰,胡
)

type Player struct {
	Name   string
	Reader chan *PlayerAction
}

// 玩家动作
type PlayerAction struct {
	Player     *Player
	ActionType ActionType
	Cards      []uint16 // 动作哪几张牌, 比如亮倒隐藏刻子有用
	Card       uint16   // 动作哪张牌
}

func (p *Player) Query(actions []ActionType) {
	// 模拟玩家动作
	p.Reader <- p.HandleAction(actions)

	return
}

func (p *Player) Read() (playerAction *PlayerAction) {
	playerAction = <-p.Reader
	return
}

func (p *Player) HandleAction(actions []ActionType) (playerAction *PlayerAction) {
	playerAction = &PlayerAction{
		Player:     p,
		ActionType: actions[0],
		Card:       100,
	}
	// 模拟玩家出牌需要1s
	time.Sleep(1 * time.Second)
	return
}

func NewPlayer() *Player{
	return &Player{
		Reader:make(chan *PlayerAction,10),
	}
}