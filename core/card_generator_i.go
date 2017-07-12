package core

type CardGenerator interface {
	Reset()                        // 重置牌局(初始化)
	Shuffle()                      // 洗牌
	GetCards(length int) (cards Cards, ok bool)       // 获取当前(未发送)的牌
	GetCardsSurplus() (cards Cards) // 获取一张牌(摸牌)
}
