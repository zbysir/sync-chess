package core

type Players []Player

// 排除
func (p *Players) Exclude(players ...Player) (ps Players) {
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
func (p *Players) After(currPlayer Player) (player Player) {
	player = (*p)[0]
	l := len(*p)
	for i := 0; i < l-1; i++ {
		if (*p)[i] == currPlayer {
			player = (*p)[i+1]
		}
	}

	return
}
