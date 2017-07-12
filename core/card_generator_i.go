package core

type CardGenerator interface {
	Reset()                        // 重置牌局(初始化)
	Shuffle()                      // 洗牌
	GetCards() (cards Cards)       // 获取当前(未发送)的牌
	GetCard() (card Card, ok bool) // 获取一张牌(摸牌)
}
