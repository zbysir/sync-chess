package core

// 用于保存牌局, 实现down机恢复, 重放等功能
type Storage struct {
}

// 保存每轮开始(出牌玩家开始出牌之前)快照
func (p *Storage) SnapShoot() {

}

// 保存玩家操作日志
func (p *Storage) Step() {

}
