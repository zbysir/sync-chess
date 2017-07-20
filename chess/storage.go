package chess

import (
	"github.com/bysir-zl/bygo/log"
	"bytes"
	"encoding/json"
	"errors"
)

// 用于保存牌局, 实现down机恢复, 重放等功能
type Storage struct {
	manager       *Manager
	playerActionC map[string]chan *PlayerActionRequest // 玩家记录存档
}

type SnapShoot struct {
	Players            Players
	RoundStartPlayerId string
	SurplusCards       Cards
}

type Step struct {
	PlayerId      string
	ActionRequest *PlayerActionRequest
}

func (p *Storage) Mount(m *Manager) {
	p.manager = m
}

// 保存每轮开始(出牌玩家开始出牌之前)快照
func (p *Storage) SnapShoot() {
	players := p.manager.Players
	roundStartPlayer := p.manager.RoundStartPlayer
	surplusCards := p.manager.CardGenerator.GetCardsSurplus()

	s := SnapShoot{
		Players:            players,
		RoundStartPlayerId: roundStartPlayer.GetId(),
		SurplusCards:       surplusCards,
	}
	sBs, err := s.Marshal()
	if err != nil {
		panic(err)
	}
	bs := make([]byte, len(sBs)+1)
	bs[0] = 1
	copy(bs[1:], sBs)

	Redis.RPUSH(p.manager.Id, bs)

	log.Info("storage SnapShoot ", s)
}

// 恢复快照,并且读取待运行的操作
// 拉取所有记录, 找到最近一次SnapShoot并恢复
// 将Step记录转换为manager.AutoAction, 当有AutoAction时manager不在询问player而是自动action
func (p *Storage) Recovery() (has bool) {
	ss, err := Redis.LRANGE(p.manager.Id, 0, -1)
	if err != nil {
		log.Error("Recovery ERR: ", err)
		return
	}

	steps := []*Step{}
	for i := len(ss) - 1; i >= 0; i-- {
		bs := ss[i].([]byte)
		switch bs[0] {
		case 1:
			// SnapShoot
			snap := SnapShoot{}
			err := snap.Unmarshal(bs[1:], p.manager.playerLeader.PlayerCreator)
			if err != nil {
				log.Error("snap.Unmarshal Err:", err)
				return
			}
			p.manager.Players = snap.Players
			p.manager.RoundStartPlayer, _ = p.manager.Players.Find(snap.RoundStartPlayerId)
			p.manager.CardGenerator.SetCardsSurplus(snap.SurplusCards)

			log.Info("storage Recovery", snap)
			has = true
		case 2:
			// Step
			step := &Step{}
			err := step.Unmarshal(bs[1:])
			if err != nil {
				log.Error("step.Unmarshal Err:", err)
				return
			}
			steps = append(steps, step)
		}
		if has {
			break
		}
	}

	// 存储step
	for _, step := range steps {
		if _, ok := p.playerActionC[step.PlayerId]; !ok {
			p.playerActionC[step.PlayerId] = make(chan *PlayerActionRequest, 100)
		}
		p.playerActionC[step.PlayerId] <- step.ActionRequest
		log.Info("storage Recovery Step", step)
	}

	return
}

// 保存玩家操作日志
func (p *Storage) Step(playerId string, request *PlayerActionRequest) {
	s := Step{
		PlayerId:      playerId,
		ActionRequest: request,
	}
	sBs, err := json.Marshal(&s)
	if err != nil {
		panic(err)
	}
	bs := make([]byte, len(sBs)+1)
	bs[0] = 2
	copy(bs[1:], sBs)

	Redis.RPUSH(p.manager.Id, bs)

	log.Info("storage Step ", playerId, request)
}

// 清空这局存档
func (p *Storage) Clean() {
	Redis.DEL(p.manager.Id)

	log.Info("storage Cleaned")
	return
}

// 获取玩家动作存档
func (p *Storage) HasStep(playerId string) (has bool) {
	c, ok := p.playerActionC[playerId]
	if !ok {
		return
	}
	has = len(c) != 0
	return
}

// 获取玩家动作存档
func (p *Storage) PopStep(playerId string) (action *PlayerActionRequest, has bool) {
	c, ok := p.playerActionC[playerId]
	if !ok {
		return
	}
	select {
	case action = <-c:
		has = true
		return
	default:
		return
	}
	return
}

// ---------------

var sp = []byte("@#$%$#@")
var spPlayer = []byte("^&*(*&^")

func (s *SnapShoot) Marshal() (bs []byte, err error) {
	var buff bytes.Buffer
	playerBs := [][]byte{}
	for _, player := range s.Players {
		pbs, e := player.Marshal()
		if e != nil {
			err = e
			return
		}
		playerBs = append(playerBs, pbs)
	}
	// 写入玩家
	buff.Write(bytes.Join(playerBs, spPlayer))
	// 写入局头人
	buff.Write(sp)
	buff.Write([]byte(s.RoundStartPlayerId))
	// 写入剩余卡牌
	buff.Write(sp)
	cardsBs, err := json.Marshal(s.SurplusCards)
	if err != nil {
		return
	}
	buff.Write(cardsBs)

	bs = buff.Bytes()

	return
}

func (s *SnapShoot) Unmarshal(bs []byte, PlayerCreator func(string) Player) (err error) {
	bsp := bytes.Split(bs, sp)
	if len(bsp) != 3 {
		err = errors.New("bad format")
		return
	}
	// 读取玩家
	if len(bsp[0]) == 0 {
		err = errors.New("bad format: player")
		return
	}
	playerBs := bytes.Split(bsp[0], spPlayer)
	if len(playerBs) == 0 {
		err = errors.New("bad format: player len is zero")
		return
	}
	lenPlayer := len(playerBs)
	players := make(Players, lenPlayer)
	for i, pbs := range playerBs {
		player := PlayerCreator("")
		err = player.Unmarshal(pbs)
		if err != nil {
			return
		}
		players[i] = player
	}
	s.Players = players
	// 读取id
	s.RoundStartPlayerId = string(bsp[1])
	// 剩余卡牌
	err = json.Unmarshal(bsp[2], &s.SurplusCards)
	if err != nil {
		return
	}

	return
}

func (s *Step) Marshal() (bs []byte, err error) {
	bs, err = json.Marshal(s)
	return
}

func (s *Step) Unmarshal(bs []byte) (err error) {
	err = json.Unmarshal(bs, s)
	return
}

func NewStorage() *Storage {
	return &Storage{
		playerActionC: map[string]chan *PlayerActionRequest{},
	}
}
