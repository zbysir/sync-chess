package core

type ActionType uint16
type ActionTypes []ActionType

const (
	AT_Get         ActionType = iota + 1 // 摸牌, 服务器不会下发这个命令 而是自动(比如杠牌后)下发通知告知玩家摸了那张牌
	AT_Play                              // 出牌
	AT_Peng                              // 碰
	AT_GangDian                          // 直杠
	AT_GangAn                            // 并且手上有三张了 就是暗杠;
	AT_GangBu                            // 是自摸的牌 并且是碰过的就是补杠
	AT_HuDian                            // 点炮
	AT_HuZiMo                            // 自摸
	AT_HuQiangGang                       // 抢杠胡
	AT_LiangDao                          // 亮倒
	AT_Pass                              // 过, 可以过 杠,碰,胡
)

func (p ActionType) String() (s string) {
	switch p {
	case AT_Get:
		s = "Get"
	case AT_Play:
		s = "Play"
	case AT_Peng:
		s = "Pong"
	case AT_GangDian:
		s = "GangDian"
	case AT_GangBu:
		s = "GangBu"
	case AT_GangAn:
		s = "GangAn"
	case AT_HuDian:
		s = "HuDian"
	case AT_HuZiMo:
		s = "HuZiMo"
	case AT_HuQiangGang:
		s = "QiangGang"
	case AT_LiangDao:
		s = "LiangDao"
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

type GangType int8

const (
	GT_Bu   GangType = iota + 1 // 补杠/扒杠
	GT_An                       // 暗杠/自杠
	GT_Dian                     // 点杠
)
