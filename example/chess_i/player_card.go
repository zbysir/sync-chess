package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
)

type PlayerCards struct {
	Cards chess.Cards         // 手上的牌
	Pong  chess.Cards         // 碰的牌
	Gang  map[chess.Card]Gang // 杠的牌
}

type Gang struct {
	Score    int32          // 分数, 杠需要记录扣分的人. 杠上杠的情况分数不一样
	Receiver chess.Player   // 接收者
	Giver    []chess.Player // 给予者
	Types    GangType
}

type GangType int8

const (
	GT_Bu   GangType = iota + 1 // 补杠/扒杠
	GT_An                       // 暗杠/自杠
	GT_Dian                     // 点杠
)
