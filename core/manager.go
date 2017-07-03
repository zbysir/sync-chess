package core

import (
	"context"
	"log"
)

// 管理员管理整个打牌逻辑
// 命令该谁出牌
// 一个房间一个Manager
type Manager struct {
	Players         Players
	PlayerActionLog []*PlayerAction // 成功操作的玩家动作记录, 用与回放 恢复场景 和 检查杠上花等
}

type WaitActionPlayer struct {
	Player     *Player
	CanActions ActionTypes
}

// 开始监督
func (p *Manager) StartSupervise() {
	if len(p.Players) == 0 {
		panic("players is nil")
	}
	ctx := context.Background()

	// 设置第一个出牌者
	firstPlayer := p.Players[0]
	// 胡牌才会结束
	for {
	startRound:
		for _, player := range p.Players {
			p.ClosePlayerAction(player)
		}

		as := p.GetCanActions(firstPlayer, true, 0)
		p.NotifyNeedAction(firstPlayer, as)
	get:
		a, err := p.GetPlayerAction(ctx, firstPlayer, as)
		if err != nil {
			log.Printf("GetPlayerAction err %+v", err)
			goto get
		}
		log.Printf("GetPlayerAction %+v %+v", firstPlayer, a)

		switch a.ActionType {
		case AT_Play:
			// 出牌, 通知其他人
			card := a.Card
			waitActionPlayer := []WaitActionPlayer{}
			for _, player := range p.Players.Exclude(firstPlayer) {
				as := p.GetCanActions(player, false, card)
				if len(as) != 0 {
					p.NotifyNeedAction(player, as)
					// 胡牌优先选择
					if as.Contain(AT_Hu) {
						waitActionPlayer = append([]WaitActionPlayer{{Player: player, CanActions: as}}, waitActionPlayer...)
					} else {
						waitActionPlayer = append(waitActionPlayer, WaitActionPlayer{Player: player, CanActions: as})
					}
				}
			}
			isHasHu := false
			for _, wap := range waitActionPlayer {
				player := wap.Player
				a, _ := p.GetPlayerAction(ctx, player, wap.CanActions)
				log.Printf("GetPlayerAction %+v %+v", player, a)
				switch a.ActionType {
				case AT_Pass:
					continue
				case AT_Gang:
					// 胡牌和杠碰是互斥的
					// 胡过之后就不能杠碰
					if !isHasHu {
						// todo 摸牌
						firstPlayer = player
						goto startRound
					}
				case AT_Peng:
					if !isHasHu {
						firstPlayer = player
						goto startRound
					}
				case AT_Hu:
					isHasHu = true
				}
			}
			if isHasHu {
				// 有人胡牌就结束,可能是多个人胡
				goto end
			}

			// 若没人动作 下家摸牌
			firstPlayer = p.Players.After(firstPlayer)
			// todo 摸牌
			goto startRound
		case AT_Gang:
			// 抢杠胡(补杠)
			card := a.Card
			isBuGang := true
			if isBuGang {
				// 判断其他玩家有不有胡
				waitActionPlayer := []WaitActionPlayer{}
				for _, player := range p.Players.Exclude(firstPlayer) {
					as := p.GetCanActions(player, false, card)
					if as.Contain(AT_Hu) {
						p.NotifyNeedAction(player, ActionTypes{AT_Hu, AT_Pass})
						waitActionPlayer = append(waitActionPlayer, WaitActionPlayer{Player: player, CanActions: as})
					}
				}

				isHasHu := false
				// 有玩家可胡,就获取玩家操作
				for _, wap := range waitActionPlayer {
					player := wap.Player
					a, _ := p.GetPlayerAction(ctx, player, wap.CanActions)
					log.Printf("GetPlayerAction %+v %+v", player, a)
					switch a.ActionType {
					case AT_Pass:
						continue
					case AT_Hu:
						isHasHu = true
					}
				}
				if isHasHu {
					// 有人胡牌就结束,可能是多个人胡
					goto end
				}
			}

			// todo 摸牌
			goto startRound
		case AT_Peng:
			goto startRound
		case AT_Hu:
			// 胡牌就结束
			goto end
		}
	}
end:

	log.Printf("end ")
}

func (p *Manager) NotifyNeedAction(player *Player, actions []ActionType) {
	p.OpenPlayerAction(player)

	log.Printf("NeedAction %+v %+v", player, actions)

	player.Query(actions)
}

// isFirst 是否该他出牌
// card 能够吃的牌(其他人打来的)
func (p *Manager) GetCanActions(player *Player, isFirst bool, card uint16) (actions ActionTypes) {
	actions = player.GetCanActions(isFirst, card)
	return
}

// 阻塞获取玩家动作
func (p *Manager) GetPlayerAction(context context.Context, player *Player, canActions ActionTypes) (action *PlayerAction, err error) {
	for {
		select {
		case <-context.Done():
			err = context.Err()
			return
		case playerAction := <-player.Reader:
			log.Printf("[debug] GetPlayerAction: %+v %+v", player, playerAction)
			// 如果不在说明有误 再次获取
			if !canActions.Contain(playerAction.ActionType) {
				continue
			}
			// 接收成功一次就关闭接收
			p.ClosePlayerAction(player)
			action = playerAction
			return
		}
	}
}

// 开启接收玩家消息
func (p *Manager) OpenPlayerAction(player *Player) {
	player.IsCanReceive = true
	// 清空缓存
	for {
		select {
		case <-player.Reader:
		default:
			return
		}
	}
}

// 关闭接收玩家消息
func (p *Manager) ClosePlayerAction(player *Player) {
	player.IsCanReceive = false
}
