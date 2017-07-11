package core

import "log"

// 用于保存牌局, 实现down机恢复, 重放等功能
type Storage struct {
	manager *Manager
	//map[] // 存储用户操作
}

// 保存每轮开始(出牌玩家开始出牌之前)快照
func (p *Storage) SnapShoot() {
	players := p.manager.Players
	roundStartPlayer := p.manager.RoundStartPlayer
	surplusCards := p.manager.CardGenerator.GetCards()

	log.Print("Storage SnapShoot ", players, roundStartPlayer, surplusCards)
}

// 恢复快照,并且读取待运行的操作
func (p *Storage) Recovery() {
	//players := p.manager.Players
	//roundStartPlayer := p.manager.RoundStartPlayer
	//surplusCards := p.manager.CardGenerator.GetCards()

	log.Print("Storage Recoveryed " )
}

// 保存玩家操作日志
func (p *Storage) Step(player Player, request *PlayerActionRequest) {
	log.Print("Storage Step ", player, request)
}

func NewStorage(manager *Manager) *Storage {
	return &Storage{manager: manager}
}
