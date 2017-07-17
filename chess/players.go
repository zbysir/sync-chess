package chess

type Players []*Player

// 排除
func (p *Players) Exclude(players ...*Player) (ps Players) {
	ps = Players{}
	for _, player := range *p {
		isExclude := false
		for _, eplayer := range players {
			if eplayer == player {
				isExclude = true
			}
		}
		if !isExclude {
			ps = append(ps, player)
		}
	}
	return
}

// 得到下家
func (p *Players) After(currPlayer *Player) (player *Player) {
	player = (*p)[0]
	l := len(*p)
	for i := 0; i < l-1; i++ {
		if (*p)[i] == currPlayer {
			player = (*p)[i+1]
		}
	}

	return
}

func (p *Players) Add(player *Player) (ok bool) {
	if index := p.Index(player); index != -1 {
		return
	}

	ok = true
	*p = append(*p, player)
	return
}

func (p *Players) RemoveById(playerId string) (ok bool) {
	for i, player := range *p {
		if player.Id == playerId {
			ok = true
			*p = append((*p)[0:i], (*p)[i+1:]...)
			return
		}
	}
	return
}

func (p *Players) Index(player *Player) (index int) {
	index = -1
	for i, playerI := range *p {
		if playerI.Id == player.Id {
			index = i
			return
		}
	}
	return
}

func (p *Players) Find(id string) (player *Player, index int) {
	index = -1
	for i, playerI := range *p {
		if playerI.Id == id {
			index = i
			player = (*p)[i]
			return
		}
	}
	return
}
