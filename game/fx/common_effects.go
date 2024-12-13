package fx

import (
	"duel-masters/game/cnd"
	"duel-masters/game/match"
	"fmt"
	"math/rand"
)

func GiveTapAbilityToAllies(card *match.Card, ctx *match.Context, alliesFilter func(card *match.Card) bool, tapAbility func(card *match.Card, ctx *match.Context)) {
	// This is added for the case where the card is added to the field. There is another creature
	// that doesn't initially have a tap abbility but should receive one. The change doesn't propagate fast
	// enough to the FE and that creature doesn't get tap ability until another action takes places.
	// This is an ugly workaround.
	FindFilter(
		card.Player,
		match.BATTLEZONE,
		alliesFilter,
	).Map(func(x *match.Card) {
		x.AddUniqueSourceCondition(cnd.TapAbility, tapAbility, card.ID)
	})

	ctx.Match.ApplyPersistentEffect(func(ctx2 *match.Context, exit func()) {
		if card.Zone != match.BATTLEZONE {
			Find(
				card.Player,
				match.BATTLEZONE,
			).Map(func(x *match.Card) {
				x.RemoveConditionBySource(card.ID)
			})

			exit()
			return
		}

		FindFilter(
			card.Player,
			match.BATTLEZONE,
			alliesFilter,
		).Map(func(x *match.Card) {
			x.AddUniqueSourceCondition(cnd.TapAbility, tapAbility, card.ID)
		})
	})
}

func FilterShieldTriggers(ctx *match.Context, filter func(*match.Card) bool) {

	if event, ok := ctx.Event.(*match.ShieldTriggerEvent); ok {
		validCards, invalidCards := FilterCardList(event.Cards, filter)
		event.Cards = validCards
		event.UnplayableCards = append(event.UnplayableCards, invalidCards...)
	}

}

func OpponentDiscardsRandomCard(card *match.Card, ctx *match.Context) {

	hand, err := ctx.Match.Opponent(card.Player).Container(match.HAND)

	if err != nil || len(hand) < 1 {
		return
	}

	discardedCard, err := ctx.Match.Opponent(card.Player).MoveCard(hand[rand.Intn(len(hand))].ID, match.HAND, match.GRAVEYARD, card.ID)
	if err == nil {
		ctx.Match.ReportActionInChat(ctx.Match.Opponent(card.Player), fmt.Sprintf("%s was discarded from %s's hand by %s", discardedCard.Name, discardedCard.Player.Username(), card.Name))
	}

}

func OpponentDiscards2RandomCards(card *match.Card, ctx *match.Context) {
	OpponentDiscardsRandomCard(card, ctx)
	OpponentDiscardsRandomCard(card, ctx)
}

// To be used as part of a card effect, not for initial shuffle
func ShuffleDeck(card *match.Card, ctx *match.Context, forOpponent bool) {
	if !forOpponent {
		card.Player.ShuffleDeck()
		ctx.Match.ReportActionInChat(card.Player, fmt.Sprintf("%s shuffled their deck", card.Player.Username()))
	} else {
		opponent := ctx.Match.Opponent(card.Player)
		opponent.ShuffleDeck()
		ctx.Match.ReportActionInChat(opponent, fmt.Sprintf("%s deck shuffled by %s effect", opponent.Username(), card.Name))
	}

}

func CantBeBlockedByPowerUpTo(card *match.Card, ctx *match.Context, power int) {
	blockersList := BlockersList(ctx)
	var newBlockersList []*match.Card
	for _, blocker := range *blockersList {
		if ctx.Match.GetPower(blocker, false) > power {
			newBlockersList = append(newBlockersList, blocker)
		}
	}
	*blockersList = newBlockersList
}

func GiveOwnCreatureCantBeBlocked(card *match.Card, ctx *match.Context) {
	Select(card.Player, ctx.Match, card.Player, match.BATTLEZONE,
		"Choose a card to receive 'Can't be blocked this turn'", 1, 1, false,
	).Map(func(x *match.Card) {
		x.AddCondition(cnd.CantBeBlocked, nil, card.ID)
		ctx.Match.ReportActionInChat(card.Player,
			fmt.Sprintf("%s tap effect: %s can't be blocked this turn", card.Name, x.Name))
	})
}

func CantBeBlockedByPowerUpTo4000(card *match.Card, ctx *match.Context) {
	CantBeBlockedByPowerUpTo(card, ctx, 4000)
}

func CantBeBlockedByPowerUpTo5000(card *match.Card, ctx *match.Context) {
	CantBeBlockedByPowerUpTo(card, ctx, 5000)
}

func CantBeBlockedByPowerUpTo8000(card *match.Card, ctx *match.Context) {
	CantBeBlockedByPowerUpTo(card, ctx, 8000)
}

func CantBeBlockedByPowerUpTo3000(card *match.Card, ctx *match.Context) {
	CantBeBlockedByPowerUpTo(card, ctx, 3000)
}

func RemoveBlockerFromList(card *match.Card, ctx *match.Context) {
	blockersList := BlockersList(ctx)
	var newBlockersList []*match.Card
	for _, blocker := range *blockersList {
		if blocker.ID != card.ID {
			newBlockersList = append(newBlockersList, blocker)
		}
	}
	*blockersList = newBlockersList
}
func BlockerWhenNoShields(card *match.Card, ctx *match.Context) {
	condition := &match.Condition{ID: cnd.Blocker, Val: true, Src: nil}
	HaveSelfConditionsWhenNoShields(card, ctx, []*match.Condition{condition})
}

func HaveSelfConditionsWhenNoShields(card *match.Card, ctx *match.Context, conditions []*match.Condition) {

	ctx.Match.ApplyPersistentEffect(func(ctx2 *match.Context, exit func()) {

		notInTheBZ := card.Zone != match.BATTLEZONE
		if notInTheBZ || IHaveShields(card, ctx2) {
			for _, cond := range conditions {
				card.RemoveSpecificConditionBySource(cond.ID, card.ID)
			}
		}

		if notInTheBZ {
			exit()
			return
		}

		if IDontHaveShields(card, ctx2) {
			for _, cond := range conditions {
				card.AddUniqueSourceCondition(cond.ID, cond.Val, card.ID)
			}
		}

	})
}

func RotateShields(card *match.Card, ctx *match.Context, max int) {

	nrShields, err := card.Player.Container(match.SHIELDZONE)
	if err != nil {
		return
	}

	if len(nrShields) < 1 {
		return
	}

	toShield := Select(
		card.Player,
		ctx.Match,
		card.Player,
		match.HAND,
		fmt.Sprintf("%s: You may select up to %d card(s) from your hand and put it into the shield zone", card.Name, max),
		0, max, true,
	)

	cardsMoved := len(toShield)
	if cardsMoved < 1 {
		return
	}

	for _, c := range toShield {
		c.Player.MoveCard(c.ID, match.HAND, match.SHIELDZONE, card.ID)
	}

	toHand := SelectBackside(
		card.Player,
		ctx.Match,
		card.Player,
		match.SHIELDZONE,
		fmt.Sprintf("%s: Select %d of your shields that will be moved to your hand", card.Name, cardsMoved),
		cardsMoved,
		cardsMoved,
		false,
	)

	for _, c := range toHand {
		c.Player.MoveCard(c.ID, match.SHIELDZONE, match.HAND, card.ID)
	}

}

func DestoryOpShield(card *match.Card, ctx *match.Context) {
	opponent := ctx.Match.Opponent(card.Player)

	ctx.Match.BreakShields(SelectBackside(
		card.Player,
		ctx.Match,
		opponent,
		match.SHIELDZONE,
		fmt.Sprintf("%s effect: select shield to break", card.Name),
		1,
		1,
		false,
	), card)

	ctx.Match.ReportActionInChat(ctx.Match.Opponent(card.Player),
		fmt.Sprintf("%s effect broke one of %s's shields", card.Name, opponent.Username()))

}

func OpDiscardsXCards(x int) match.HandlerFunc {
	return func(card *match.Card, ctx *match.Context) {

		min := 0
		handCount := ctx.Match.Opponent(card.Player).Denormalized().HandCount

		if x > handCount {
			min = handCount
		} else {
			min = x
		}

		Select(
			ctx.Match.Opponent(card.Player),
			ctx.Match,
			ctx.Match.Opponent(card.Player),
			match.HAND,
			fmt.Sprintf("%s: Select %d card(s) from your hand that will be sent to your graveyard", card.Name, x),
			min,
			x,
			false,
		).Map(func(x *match.Card) {
			x.Player.MoveCard(x.ID, match.HAND, match.GRAVEYARD, card.ID)
			ctx.Match.ReportActionInChat(ctx.Match.Opponent(card.Player), fmt.Sprintf("%s was moved from %s's hand to his graveyard by %s", x.Name, x.Player.Username(), card.Name))
		})
	}
}

// Look at opponent's x shields
func ShowXShields(x int) match.HandlerFunc {
	return func(card *match.Card, ctx *match.Context) {

		shieldsID := []string{}

		SelectBackside(
			card.Player,
			ctx.Match,
			ctx.Match.Opponent(card.Player),
			match.SHIELDZONE,
			fmt.Sprintf("%s: Select %d of your opponent's shields that will be shown to you", card.Name, x),
			1,
			x,
			true,
		).Map(func(shields *match.Card) {
			shieldsID = append(shieldsID, shields.ImageID)
		})

		ctx.Match.ShowCards(
			card.Player,
			"Your opponent's shield:",
			shieldsID,
		)
	}

}

func OpponentDiscardsHand(card *match.Card, ctx *match.Context) {

	Find(
		ctx.Match.Opponent(card.Player),
		match.HAND,
	).Map(func(x *match.Card) {
		ctx.Match.Opponent(card.Player).MoveCard(x.ID, match.HAND, match.GRAVEYARD, card.ID)
	})

	ctx.Match.ReportActionInChat(ctx.Match.Opponent(card.Player), fmt.Sprintf("%s's hand was discarded by %s", ctx.Match.Opponent(card.Player).Username(), card.Name))

}
