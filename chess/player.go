package chess

import (
	"time"
)

type Player struct {
	Name         string
	Reader       chan *PlayerAction
	IsCanReceive bool // 只有在 接收 中的玩家能向服务器发送数据
	CanActions   ActionTypes
}

type Players []*Player

// 排除
func (p *Players) Exclude(players ...*Player) (ps Players) {
	ps = Players{}
	for _, player := range *p {
		isExclude := false
		for _, eplayer := range players {
			if eplayer == player {
				isExclude = true
			}
		}
		if !isExclude {
			ps = append(ps, player)
		}
	}
	return
}

func (p *Player) GetCanActions(isMyPlay bool, card uint16) ActionTypes {
	return p.CanActions
}

func (p *Player) String() (s string) {
	s = p.Name
	return
}

func (p *Player) WriteAction(action *PlayerAction) bool {
	if !p.IsCanReceive {
		return false
	}
	select {
	case p.Reader <- action:
	default:
		return false
	}
	return true
}

func (p *Player) Query(actions []ActionType) {
	return
}

func (p *Player) Read() (playerAction *PlayerAction) {
	playerAction = <-p.Reader
	return
}

func (p *Player) HandleAction(actions []ActionType) (playerAction *PlayerAction) {
	playerAction = &PlayerAction{
		ActionType: actions[0],
		Card:       100,
	}
	// 模拟玩家出牌需要1s
	time.Sleep(1 * time.Second)
	return
}

func NewPlayer() *Player {
	return &Player{
		Reader: make(chan *PlayerAction, 1),
	}
}
