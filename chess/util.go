package chess

import (
	"sort"
	"github.com/bysir-zl/bygo/log"
)

// 可能有用的工具

func IsHu(cards Cards, cheatCardLen int) (bool) {

	count := map[Card]int{}
	for _, c := range cards {
		count[c]++
	}

	T := []Cards{}
	for card, c := range count {
		cardsTemp := make(Cards, len(cards))
		copy(cardsTemp, cards)
		sort.Sort(&cardsTemp)

		if c >= 2 {
			// 移除两张对子
			cardsTemp.Delete(card)
			cardsTemp.Delete(card)
			T = append(T, cardsTemp)
		}
	}

	for _, cards := range T {
		log.Info("test-start", cards)

		// 移除3连子
	removeLian:
		if cards.Len() == 0 {
			return true
		}

		count2 := map[Card]int{}
		for _, c := range cards {
			count2[c]++
		}

		for _, card := range cards {
			if count2[card+1] > 0 && count2[card+2] > 0 {
				cards.Delete(card)
				cards.Delete(card + 1)
				cards.Delete(card + 2)
				log.Info("test-lian", card, card+1, card+2)

				goto removeLian
			}
		}

		log.Info("test-lian-end", cards)

		if cards.Len() == 0 {
			return true
		}

		// 移除3张相同的
		count2 = map[Card]int{}
		for _, c := range cards {
			count2[c]++
		}

		for card2, c2 := range count2 {
			if c2 >= 3 {
				cards.Delete(card2)
				cards.Delete(card2)
				cards.Delete(card2)
				log.Info("test-kezi", card2)
				log.Info("test-kezi-end", cards)
				goto removeLian
			}
		}
	}

	return false
}

func cardsRemoveKezi(cards Cards) (ok bool) {
	count := map[Card]int{}
	for _, c := range cards {
		count[c]++
	}

	return
}
