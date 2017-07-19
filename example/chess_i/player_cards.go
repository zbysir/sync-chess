package chess_i

import "github.com/bysir-zl/sync-chess/chess"

type Gangs []*Gang

type Gang struct {
	Card  chess.Card
	Score int32          // 分数, 杠需要记录扣分的人. 杠上杠的情况分数不一样
	Giver []string // 给予者Id
	Types GangType
}

type GangType int8

const (
	GT_Bu   GangType = iota + 1 // 补杠/扒杠
	GT_An                       // 暗杠/自杠
	GT_Dian                     // 点杠
)

// 牌型
type HuCardType int32

type Hu struct {
	IsHued    bool           // 是否胡了
	CardTypes []HuCardType   // 胡牌牌型
	Giver     []chess.Player // 给予者
}

func (p *Gangs) Add(gang *Gang) {
	*p = append(*p, gang)
	return
}
