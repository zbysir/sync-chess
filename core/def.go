package core

type GangType int8

const (
	GT_Bu   GangType = iota + 1 // 补杠/扒杠
	GT_An                       // 暗杠/自杠
	GT_Dian                     // 点杠
)
