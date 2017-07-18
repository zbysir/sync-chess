package chess

// Leader
type PlayerLeader interface {
	// 获取庄家
	Banker(players Players) (player Player)
	// 获取下家
	Next(currPlayer Player, players Players) (player Player)
	// 生成器
	PlayerCardsCreator() Player
}
