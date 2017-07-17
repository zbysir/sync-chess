package chess

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
	Id                   string                               // 管理员唯一标示
	Players              Players                              // 所有玩家
	LastPlayerNeedAction map[string]ActionTypes               // 最后一次需要玩家的动作, 用于玩家重连重新发送请求
	RoundStartPlayer     *Player                              // 每轮开始者
	CardGenerator        CardGenerator                        // 发牌器
	PlayerLeader         PlayerLeader                         // 玩家领导
	Storage              *Storage                             // 存档器
	IsInitFromStorage    bool                                 // 是否是从存档恢复
	isStartSupervise     bool                                 // 是否开始了游戏
	MessageHandler       MessageHandler                       // 消息通知
	PlayerActionRequestC map[string]chan *PlayerActionRequest // 玩家动作请求队列
}

type WaitActionPlayer struct {
	Player     *Player
	CanActions ActionTypes
}

var timeout = time.Second * 10000

// 开始监督(游戏进行中)
func (p *Manager) StartSupervise() {
	if p.isStartSupervise {
		return
	}
	p.isStartSupervise = true

	if len(p.Players) == 0 {
		panic("players is nil")
	}
	ctx := context.Background()

	if !p.IsInitFromStorage {
		p.startGame()
	}

	go func() {
		// 胡牌才会结束
	startRound:

	// 不需要读存档 说明不是第一次运行的现场恢复,则需要存档
		if !p.IsInitFromStorage {
			// 存档
			p.Storage.SnapShoot()
		} else {
			p.IsInitFromStorage = false
		}
		as := p.GetCanActions(p.RoundStartPlayer, true, 0)
		p.NotifyNeedAction(p.RoundStartPlayer, as)
		ctxTime, _ := context.WithTimeout(ctx, timeout)
	getRoundPlayerAction:
		a, err := p.GetPlayerAction(ctxTime, p.RoundStartPlayer, as, 0)
		if err != nil {
			p.NotifyNeedAction(p.RoundStartPlayer, as)
			goto getRoundPlayerAction
		}
		log.Info("GetPlayerAction", "%+v %+v", p.RoundStartPlayer, a)
		switch a.Types {
		case AT_Play:
			// 出牌
			err := p.RoundStartPlayer.PlayerI.DoAction(a, p.RoundStartPlayer)
			if err != nil {
				p.NotifyNeedAction(p.RoundStartPlayer, as)
				goto getRoundPlayerAction
			}
			p.DoActionAfter(p.RoundStartPlayer, a)

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
				a, err := p.GetPlayerAction(ctxTime, player, wap.CanActions, card)
				// 超时未响应则执行自动打牌
				if err != nil {
					p.NotifyNeedAction(player, wap.CanActions)
					goto getOtherPlayerAction
				}

				log.Info("GetPlayerAction", "%+v %+v", player, a)
				switch a.Types {
				case AT_Pass:
					p.DoActionAfter(player, a)
					continue
				case AT_GangDian:
					// 胡牌和杠碰是互斥的
					// 胡过之后就不能杠碰
					if !isHasHu {
						err := player.PlayerI.DoAction(a, p.RoundStartPlayer)
						if err != nil {
							p.NotifyNeedAction(player, wap.CanActions)
							goto getOtherPlayerAction
						}
						p.DoActionAfter(player, a)

						p.RoundStartPlayer = player
						err = p.GetCard(p.RoundStartPlayer)
						if err != nil {
							// 没牌了,直接结束
							goto end
						}
						goto startRound
					}
				case AT_Peng:
					if !isHasHu {
						err = player.PlayerI.DoAction(a, p.RoundStartPlayer)
						if err != nil {
							p.NotifyNeedAction(player, wap.CanActions)
							goto getOtherPlayerAction
						}
						p.DoActionAfter(player, a)
						p.RoundStartPlayer = player
						goto startRound
					}
				case AT_HuDian:
					isHasHu = true
					err := player.PlayerI.DoAction(a, p.RoundStartPlayer)
					if err != nil {
						p.NotifyNeedAction(player, wap.CanActions)
						goto getOtherPlayerAction
					}
					p.DoActionAfter(player, a)

				}
			}
			if isHasHu {
				// 有人胡牌就结束,可能是多个人胡
				goto end
			}

			// 若没人动作 下家摸牌
			p.RoundStartPlayer = p.PlayerLeader.Next(p.RoundStartPlayer, p.Players)
			err = p.GetCard(p.RoundStartPlayer)
			if err != nil {
				// 没牌了,直接结束
				goto end
			}

			goto startRound
		case AT_GangAn:
			err := p.RoundStartPlayer.PlayerI.DoAction(a, p.RoundStartPlayer)
			if err != nil {
				p.NotifyNeedAction(p.RoundStartPlayer, as)
				goto getRoundPlayerAction
			}
			p.DoActionAfter(p.RoundStartPlayer, a)

			// 摸牌
			err = p.GetCard(p.RoundStartPlayer)
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
				a, err := p.GetPlayerAction(ctxTime, player, wap.CanActions, card)
				if err != nil {
					p.NotifyNeedAction(player, wap.CanActions)
					goto getBuGangAction
				}
				log.Info("GetPlayerAction", "%+v %+v", player, a)
				switch a.Types {
				case AT_Pass:
					p.DoActionAfter(player, a)
					continue
				case AT_HuQiangGang:
					err := player.PlayerI.DoAction(a, p.RoundStartPlayer)
					if err != nil {
						// 抢杠胡都有错误? 那就当pass了吧
						continue
					}
					p.DoActionAfter(player, a)

					isHasHu = true
				}
			}
			if isHasHu {
				// 有人胡牌就结束,可能是多个人胡
				goto end
			} else {
				// 补杠逻辑
				err := p.RoundStartPlayer.PlayerI.DoAction(a, p.RoundStartPlayer)
				if err != nil {
					p.NotifyNeedAction(p.RoundStartPlayer, as)
					goto getRoundPlayerAction
				}
				p.DoActionAfter(p.RoundStartPlayer, a)

				// 摸牌
				err = p.GetCard(p.RoundStartPlayer)
				if err != nil {
					// 没牌了,直接结束
					goto end
				}
				log.Debug("goto startRound")
				goto startRound
			}
		case AT_HuZiMo:
			err := p.RoundStartPlayer.PlayerI.DoAction(a, p.RoundStartPlayer)
			if err != nil {
				p.NotifyNeedAction(p.RoundStartPlayer, as)
				goto getRoundPlayerAction
			}
			p.DoActionAfter(p.RoundStartPlayer, a)

			// 胡牌就结束
			goto end
		}

	end:
		p.Storage.Clean()
		log.Info("end ")
		return
	}()
}

// 玩家成功动作后记录,并响应玩家
func (p *Manager) DoActionAfter(player *Player, action *PlayerActionRequest, ) {
	if action.ActionFrom != AF_Storage {
		p.Storage.Step(player, action)
		rsp := &PlayerActionResponse{
			ActionFrom: action.ActionFrom,
			Card:       action.Card,
			Types:      action.Types,
		}
		notice := &PlayerActionNotice{
			Types:      action.Types,
			Card:       action.Card,
			PlayerFrom: player,
		}
		p.MessageHandler.NotifyActionResponse(player.Id, rsp)
		p.NotifyOtherPlayerAction(player, notice)
	}
}

// 摸牌
func (p *Manager) GetCard(player *Player) (err error) {
	cards, ok := p.CardGenerator.GetCards(1)
	if !ok {
		err = errors.New("not cards")
		return
	}
	card := cards[0]
	action := &PlayerActionRequest{
		Card:  card,
		Types: AT_Get,
	}
	err = player.PlayerI.DoAction(action, player)
	if err != nil {
		return
	}
	rsp := &PlayerActionResponse{
		ActionFrom: action.ActionFrom,
		Card:       action.Card,
		Types:      action.Types,
	}
	notice := &PlayerActionNotice{
		Types:      action.Types,
		Card:       action.Card,
		PlayerFrom: player,
	}
	p.MessageHandler.NotifyActionResponse(player.Id, rsp)
	p.NotifyOtherPlayerAction(player, notice)
	return
}

// 通知其他人消息
func (p *Manager) NotifyOtherPlayerAction(currPlayer *Player, notice *PlayerActionNotice) {
	// pass 消息不需要告知其他人
	if notice.Types == AT_Pass {
		return
	}

	if notice.Types == AT_Get {
		notice.Card = 0
	}

	otherPlayer := p.Players.Exclude(currPlayer)
	for _, player := range otherPlayer {
		p.MessageHandler.NotifyFromOtherPlayerAction(player.Id, notice)
	}

	return
}

// 写入玩家动作
func (p *Manager) WritePlayerAction(playerId string, action *PlayerActionRequest) (err error) {
	_, index := p.Players.Find(playerId)
	if index == -1 {
		err = errors.New("not find player id is " + playerId)
		return
	}

	if _, ok := p.PlayerActionRequestC[playerId]; !ok {
		err = errors.New("player " + playerId + " not open receive")
		return
	}

	select {
	case p.PlayerActionRequestC[playerId] <- action:
	default:
		err = errors.New("player " + playerId + " not open receive")
	}
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
			player.PlayerI.SetCards(cards)
		} else {
			cards, _ := p.CardGenerator.GetCards(13)
			player.PlayerI.SetCards(cards)
		}
	}
}

func (p *Manager) NotifyNeedAction(player *Player, actions ActionTypes) {
	// 有动作存档 就不发送通知,而是直接读取
	playerId := player.Id
	if p.Storage.HasStep(playerId) {
		return
	}

	// 开启通道
	if c, ok := p.PlayerActionRequestC[playerId]; ok {
		<-c
	} else {
		p.PlayerActionRequestC[playerId] = make(chan *PlayerActionRequest, 1)
	}
	// 保存玩家需要的动作, 用于重连时重发
	p.LastPlayerNeedAction[playerId] = actions
	p.MessageHandler.NotifyNeedAction(playerId, actions)
}

// isFirst 是否该他出牌
// card 能够吃的牌(其他人打来的)
func (p *Manager) GetCanActions(player *Player, isFirst bool, card Card) (actions ActionTypes) {
	// 有动作存档 就不需要或者CanAction
	if p.Storage.HasStep(player.Id) {
		return
	}
	actions = player.PlayerI.CanActions(isFirst, card)
	return
}

// 阻塞获取玩家动作
func (p *Manager) GetPlayerAction(ctx context.Context, player *Player, canActions ActionTypes, card Card) (action *PlayerActionRequest, err error) {
	playerId := player.Id
	// 有动作存档 直接读取
	if a, ok := p.Storage.PopStep(playerId); ok {
		action = a
		action.ActionFrom = AF_Storage
		return
	}
	select {
	case action = <-p.PlayerActionRequestC[playerId]:
		p.PlayerActionRequestC[playerId] <- nil
		delete(p.LastPlayerNeedAction,playerId)

		// 错误的动作, 重新获取
		if !canActions.Contain(action.Types) {
			err = ERR_BadActionTypeNeedRetry
			return
		}
		action.ActionFrom = AF_Player
	case <-ctx.Done():
		// 写入一个空 占满通道让他关闭接收消息
		p.PlayerActionRequestC[playerId] <- nil
		delete(p.LastPlayerNeedAction,playerId)

		// 超时就自动打牌
		action = p.GetPlayerActionAuto(player, canActions, card)
	}

	return
}

// 获取玩家自动动作
func (p *Manager) GetPlayerActionAuto(player *Player, canActions ActionTypes, card Card) (action *PlayerActionRequest) {
	action = player.PlayerI.RequestActionAuto(canActions, card)
	action.ActionFrom = AF_Auto
	return
}

// 添加玩家
func (p *Manager) AddPlayer(playerId string) (err error) {
	_, index := p.Players.Find(playerId)
	if index != -1 {
		err = ERR_JoinRoomPlayerExist
		return
	}

	// 如果游戏已经开始, 则不能新加玩家
	if p.isStartSupervise {
		err = errors.New("game is started")
		return
	}

	player := &Player{
		Id:      playerId,
		PlayerI: p.PlayerLeader.PlayerCreator(),
	}
	if !p.Players.Add(player) {
		err = errors.New("add player error")
		return
	}
	return
}

// 删除玩家
func (p *Manager) RemovePlayer(playerId string) (err error) {
	// 如果游戏已经开始, 则不能删除玩家
	if p.isStartSupervise {
		err = errors.New("game is started")
		return
	}
	_, index := p.Players.Find(playerId)
	if index == -1 {
		err = errors.New("this player is not in room")
		return
	}

	if !p.Players.RemoveById(playerId) {
		err = errors.New("remove player error")
		return
	}
	return
}

func NewManager(id string, cardGenerator CardGenerator, playerLeader PlayerLeader, messageHandler MessageHandler) *Manager {
	m := &Manager{
		Id:                   id,
		Players:              Players{},
		CardGenerator:        cardGenerator,
		PlayerLeader:         playerLeader,
		MessageHandler:       messageHandler,
		LastPlayerNeedAction: map[string]ActionTypes{},
		PlayerActionRequestC: map[string]chan *PlayerActionRequest{},
	}
	m.Storage = NewStorage(m)
	// 尝试读档
	m.IsInitFromStorage = m.Storage.Recovery()
	if m.IsInitFromStorage {
		m.StartSupervise()
	}
	return m
}
