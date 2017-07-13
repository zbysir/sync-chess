package core

import (
	"context"
	"time"
	"errors"
	"github.com/bysir-zl/bygo/log"
)

// 管理员管理整个打牌逻辑
// 命令该谁出牌
// 一个房间一个Manager
type Manager struct {
	Id                   string  // 管理员唯一标示
	Players              Players // 所有玩家
	PlayerCreator        func() Player
	LastPlayerNeedAction map[string]ActionTypes // 最后一次需要玩家的动作, 用于玩家重连重新发送请求
	RoundStartPlayer     Player                 // 每轮开始者
	CardGenerator        CardGenerator          // 发牌器
	PlayerLeader         PlayerLeader           // 玩家领导
	*Storage
}

type WaitActionPlayer struct {
	Player     Player
	CanActions ActionTypes
}

var timeout = time.Second * 5

// 开始监督
func (p *Manager) StartSupervise() {
	if len(p.Players) == 0 {
		panic("players is nil")
	}
	ctx := context.Background()

	// 读取存档, 用于down机恢复
	hasStorage := p.Recovery()
	if !hasStorage {
		// 没有存档 则是新的一局,初始化牌桌
		p.startGame()
	}

	go func() {
		// 胡牌才会结束
		for {
		startRound:

		// 不需要读存档 说明不是第一次运行的现场恢复,则需要存档
			if !hasStorage {
				// 存档
				p.SnapShoot()
			} else {
				hasStorage = false
			}
			ctxTime, _ := context.WithTimeout(ctx, timeout)

			as := p.GetCanActions(p.RoundStartPlayer, true, 0)
			p.NotifyNeedAction(p.RoundStartPlayer, as)

		getRoundPlayerAction:
			a, err := p.GetPlayerAction(ctxTime, p.RoundStartPlayer, as)
			// 超时未响应则执行自动打牌
			if err != nil {
				if err == context.DeadlineExceeded {
					a = p.RoundStartPlayer.RequestActionAuto(as, 0)
				} else {
					p.NotifyNeedAction(p.RoundStartPlayer, as)
					goto getRoundPlayerAction
				}
			}
			log.Info("GetPlayerAction", "%+v %+v", p.RoundStartPlayer, a)

			switch a.Types {
			case AT_Play:
				// 出牌
				rsp := p.RoundStartPlayer.DoAction(a, p.RoundStartPlayer)
				if rsp.Err != nil {
					p.NotifyNeedAction(p.RoundStartPlayer, as)
					goto getRoundPlayerAction
				}
				rsp.Types = a.Types
				p.Step(p.RoundStartPlayer, a)
				p.RoundStartPlayer.ResponseAction(rsp)
				p.Players.NotifyOtherPlayerAction(p.RoundStartPlayer, a)

				card := a.Card
				waitActionPlayer := []WaitActionPlayer{}
				for _, player := range p.Players.Exclude(p.RoundStartPlayer) {
					as := p.GetCanActions(player, false, card)
					if len(as) == 0 {
						continue
					}
					// 只有过就不通知用户了
					if len(as) == 1 && as[0] == AT_Pass {
						continue
					}
					p.NotifyNeedAction(player, as)
					// 胡牌优先选择
					if as.Contain(AT_HuDian) {
						waitActionPlayer = append([]WaitActionPlayer{{Player: player, CanActions: as}}, waitActionPlayer...)
					} else {
						waitActionPlayer = append(waitActionPlayer, WaitActionPlayer{Player: player, CanActions: as})
					}
				}
				isHasHu := false
				for _, wap := range waitActionPlayer {
					player := wap.Player
					ctxTime, _ := context.WithTimeout(ctx, timeout)

				getOtherPlayerAction:
					a, err := p.GetPlayerAction(ctxTime, player, wap.CanActions)
					// 超时未响应则执行自动打牌
					if err != nil {
						if err == context.DeadlineExceeded {
							a = player.RequestActionAuto(wap.CanActions, card)
						} else {
							p.NotifyNeedAction(player, wap.CanActions)
							goto getOtherPlayerAction
						}
					}
					log.Info("GetPlayerAction", "%+v %+v", player, a)
					switch a.Types {
					case AT_Pass:
						p.Step(player, a)
						passRsp := &PlayerActionResponse{Types: AT_Pass}
						player.ResponseAction(passRsp)
						continue
					case AT_GangDian:
						// 胡牌和杠碰是互斥的
						// 胡过之后就不能杠碰
						if !isHasHu {
							rsp := player.DoAction(a, p.RoundStartPlayer)
							if rsp.Err != nil {
								p.NotifyNeedAction(player, wap.CanActions)
								goto getOtherPlayerAction
							}
							p.Step(player, a)
							rsp.Types = a.Types
							player.ResponseAction(rsp)
							p.Players.NotifyOtherPlayerAction(player, a)

							p.RoundStartPlayer = player
							err := p.GetCard(p.RoundStartPlayer)
							if err != nil {
								// 没牌了,直接结束
								goto end
							}
							goto startRound
						}
					case AT_Peng:
						if !isHasHu {
							rsp := player.DoAction(a, p.RoundStartPlayer)
							if rsp.Err != nil {
								p.NotifyNeedAction(player, wap.CanActions)
								goto getOtherPlayerAction
							}
							p.Step(player, a)
							rsp.Types = a.Types
							player.ResponseAction(rsp)
							p.Players.NotifyOtherPlayerAction(player, a)

							p.RoundStartPlayer = player
							goto startRound
						}
					case AT_HuDian:
						isHasHu = true
						rsp := player.DoAction(a, p.RoundStartPlayer)
						if rsp.Err != nil {
							p.NotifyNeedAction(player, wap.CanActions)
							goto getOtherPlayerAction
						}
						p.Step(player, a)

						rsp.Types = a.Types
						player.ResponseAction(rsp)
						p.Players.NotifyOtherPlayerAction(player, a)
					}
				}
				if isHasHu {
					// 有人胡牌就结束,可能是多个人胡
					goto end
				}

				// 若没人动作 下家摸牌
				p.RoundStartPlayer = p.PlayerLeader.Next(p.RoundStartPlayer, p.Players)
				err := p.GetCard(p.RoundStartPlayer)
				if err != nil {
					// 没牌了,直接结束
					goto end
				}

				goto startRound
			case AT_GangAn:
				rsp := p.RoundStartPlayer.DoAction(a, p.RoundStartPlayer)
				if rsp.Err != nil {
					p.NotifyNeedAction(p.RoundStartPlayer, as)
					goto getRoundPlayerAction
				}
				p.Step(p.RoundStartPlayer, a)
				rsp.Types = a.Types
				p.RoundStartPlayer.ResponseAction(rsp)
				p.Players.NotifyOtherPlayerAction(p.RoundStartPlayer, a)

				// 摸牌
				err := p.GetCard(p.RoundStartPlayer)
				if err != nil {
					// 没牌了,直接结束
					goto end
				}
			case AT_GangBu:
				// 抢杠胡(补杠)
				card := a.Card
				// 判断其他玩家有不有胡
				waitActionPlayer := []WaitActionPlayer{}
				for _, player := range p.Players.Exclude(p.RoundStartPlayer) {
					as := p.GetCanActions(player, false, card)
					if as.Contain(AT_HuDian) {
						as = ActionTypes{AT_HuQiangGang, AT_Pass}
						p.NotifyNeedAction(player, as)
						waitActionPlayer = append(waitActionPlayer, WaitActionPlayer{Player: player, CanActions: as})
					}
				}

				isHasHu := false
				// 有玩家可胡,就获取玩家操作
				for _, wap := range waitActionPlayer {
					player := wap.Player
					ctxTime, _ := context.WithTimeout(ctx, timeout)
				getBuGangAction:
					a, err := p.GetPlayerAction(ctxTime, player, wap.CanActions)
					if err != nil {
						if err == context.DeadlineExceeded {
							a = player.RequestActionAuto(wap.CanActions, card)
						} else {
							p.NotifyNeedAction(player, wap.CanActions)
							goto getBuGangAction
						}
					}
					log.Info("GetPlayerAction", "%+v %+v", player, a)
					switch a.Types {
					case AT_Pass:
						p.Step(p.RoundStartPlayer, a)
						passRsp := &PlayerActionResponse{Types: AT_Pass}
						player.ResponseAction(passRsp)
						continue
					case AT_HuQiangGang:
						rsp := player.DoAction(a, p.RoundStartPlayer)
						if rsp.Err != nil {
							// 抢杠胡都有错误? 那就当pass了吧
							continue
						}

						p.Step(p.RoundStartPlayer, a)
						rsp.Types = a.Types
						player.ResponseAction(rsp)
						p.Players.NotifyOtherPlayerAction(player, a)
						isHasHu = true
					}
				}
				if isHasHu {
					// 有人胡牌就结束,可能是多个人胡
					goto end
				} else {
					// 补杠逻辑
					rsp := p.RoundStartPlayer.DoAction(a, p.RoundStartPlayer)
					if rsp.Err != nil {
						p.NotifyNeedAction(p.RoundStartPlayer, as)
						goto getRoundPlayerAction
					}
					p.Step(p.RoundStartPlayer, a)
					rsp.Types = a.Types
					p.RoundStartPlayer.ResponseAction(rsp)
					p.Players.NotifyOtherPlayerAction(p.RoundStartPlayer, a)

					// 摸牌
					err := p.GetCard(p.RoundStartPlayer)
					if err != nil {
						// 没牌了,直接结束
						goto end
					}

					goto startRound
				}

			case AT_HuZiMo:
				rsp := p.RoundStartPlayer.DoAction(a, p.RoundStartPlayer)
				if rsp.Err != nil {
					p.NotifyNeedAction(p.RoundStartPlayer, as)
					goto getRoundPlayerAction
				}
				p.Step(p.RoundStartPlayer, a)
				rsp.Types = a.Types
				p.RoundStartPlayer.ResponseAction(rsp)
				p.Players.NotifyOtherPlayerAction(p.RoundStartPlayer, a)

				// 胡牌就结束
				goto end
			}
			continue
		end:
			p.Clean()
			log.Info("end ")
			return
		}
	}()
}

func (p *Manager) GetCard(player Player) (err error) {
	cards, ok := p.CardGenerator.GetCards(1)
	if !ok {
		err = errors.New("not cards")
		return
	}
	card := cards[0]
	addCardAction := &PlayerActionRequest{
		Card:  card,
		Types: AT_Get,
	}
	rsp := player.DoAction(addCardAction, player)
	if rsp.Err != nil {
		err = rsp.Err
	}
	rsp.Types = AT_Get
	player.ResponseAction(rsp)
	p.Players.NotifyOtherPlayerAction(player, addCardAction)
	return
}

func (p *Manager) startGame() {
	// 选庄家
	p.RoundStartPlayer = p.Players[0]
	p.CardGenerator.Reset()
	p.CardGenerator.Shuffle()

	// 发牌
	for _, player := range p.Players {
		if player == p.RoundStartPlayer {
			cards, _ := p.CardGenerator.GetCards(14)
			player.SetCards(cards)
		} else {
			cards, _ := p.CardGenerator.GetCards(13)
			player.SetCards(cards)
		}
	}
}

func (p *Manager) NotifyNeedAction(player Player, actions ActionTypes) {
	// 保存玩家需要的动作, 用于重连时重发
	player.SetValue("NeedAction", actions)

	player.NotifyNeedAction(actions)
}

// isFirst 是否该他出牌
// card 能够吃的牌(其他人打来的)
func (p *Manager) GetCanActions(player Player, isFirst bool, card Card) (actions ActionTypes) {
	actions = player.CanActions(isFirst, card)
	return
}

// 阻塞获取玩家动作
func (p *Manager) GetPlayerAction(context context.Context, player Player, canActions ActionTypes) (action *PlayerActionRequest, err error) {
	action, err = player.WaitAction(context)
	if err != nil {
		return
	}
	// 错误的动作, 重新获取
	if !canActions.Contain(action.Types) {
		err = errors.New("except action")
	}
	return
}
