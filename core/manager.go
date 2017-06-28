package core

import "context"

// 管理员管理整个打牌逻辑
// 命令该谁出牌
// 一个房间一个Manager
type Manager struct {
	PlayerActionC   chan *PlayerAction
	Players         []*Player
	PlayerActionLog []*PlayerAction // 成功操作的玩家动作记录, 用与回放 和 检查杠上花等
}

// 开始监督
func (p *Manager) StartSupervise() {
	if len(p.Players) == 0 {
		panic("players is nil")
	}

	// 读取player输入
	for _, player := range p.Players {
		go func() {
			for {
				p.PlayerActionC <- player.Read()
			}
		}()
	}

	// 设置第一个出牌者
	firstPlayer := p.Players[0]
	// 开始监督打牌
	for {
		p.NotifyNeedAction(firstPlayer, p.GetCanAction(firstPlayer, true))
		p.GetPlayerAction(context.Background(), firstPlayer)

	}
}

func (p *Manager) NotifyNeedAction(player *Player, actions []ActionType) {
	player.Query(actions)
}

func (p *Manager) GetCanAction(player *Player, isFirst bool) (actions []ActionType) {
	if isFirst {
		actions = []ActionType{AT_Gang, AT_Play}
		return
	}
	actions = []ActionType{AT_Gang, AT_Pong, AT_Pass}
	return
}

// 阻塞获取玩家动作
func (p *Manager) GetPlayerAction(context context.Context, player *Player) (action *PlayerAction, err error) {
	for {
		select {
		case <-context.Done():
			err = context.Err()
			return
		case playerAction := <-p.PlayerActionC:
			if playerAction.Player != player {
				// 无效动作 重新读动作
				break
			}
			action = playerAction
			return
		}
	}
}
