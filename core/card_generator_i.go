package core

type CardGenerator interface {
	Reset(cards Cards)       // 重置牌局(初始化)
	Shuffle()                // 洗牌
	GetCards() (cards Cards) // 获取当前(未发送)的牌
}

