package dm06

import (
	"duel-masters/game/civ"
	"duel-masters/game/cnd"
	"duel-masters/game/family"
	"duel-masters/game/fx"
	"duel-masters/game/match"
	"fmt"
)

func NeonCluster(c *match.Card) {

	c.Name = "Neon Cluster"
	c.Power = 4000
	c.Civ = civ.Water
	c.Family = []string{family.CyberCluster}
	c.ManaCost = 7
	c.ManaRequirement = []string{civ.Water}
	c.TapAbility = func(card *match.Card, ctx *match.Context) {
		ctx.Match.ReportActionInChat(card.Player, fmt.Sprintf("%s activated %s's tap ability to draw 2 cards", card.Player.Username(), card.Name))
		card.Player.DrawCards(2)
	}

	c.Use(fx.Creature, fx.TapAbility)
}

func OverloadCluster(c *match.Card) {

	c.Name = "Overload Cluster"
	c.Power = 4000
	c.Civ = civ.Water
	c.Family = []string{family.CyberCluster}
	c.ManaCost = 5
	c.ManaRequirement = []string{civ.Water}

	c.Use(fx.Creature, ReceiveBlockerWhenOpponentPlaysCreatureOrSpell)
}

func ReceiveBlockerWhenOpponentPlaysCreatureOrSpell(card *match.Card, ctx *match.Context) {
	if event, ok := ctx.Event.(*match.SpellCast); ok {
		ReceiveBlockerWhenOpponentPlaysCard(card, ctx, event.MatchPlayerID)
	}

	if fx.AnotherCreatureSummoned(card, ctx) {
		// This check can be removed once the card in CardMoved is passed as pointer
		// And MatchPlayerID is removed
		event, ok := ctx.Event.(*match.CardMoved)
		if !ok {
			return
		}
		ReceiveBlockerWhenOpponentPlaysCard(card, ctx, event.MatchPlayerID)
	}
}

func ReceiveBlockerWhenOpponentPlaysCard(card *match.Card, ctx *match.Context, playedCardPlayerId byte) {

	if card.Zone != match.BATTLEZONE || playedCardPlayerId == 0 {
		return
	}

	// Return if it's not the opponent that plays the card
	var playedCardPlayer *match.Player
	if playedCardPlayerId == 1 {
		playedCardPlayer = ctx.Match.Player1.Player
	} else {
		playedCardPlayer = ctx.Match.Player2.Player
	}
	if card.Player == playedCardPlayer {
		return
	}

	ctx.ScheduleAfter(func() {
		card.AddCondition(cnd.Blocker, nil, card.ID)
	})

}

func FortMegacluster(c *match.Card) {

	c.Name = "Fort Megacluster"
	c.Power = 5000
	c.Civ = civ.Water
	c.Family = []string{family.CyberCluster}
	c.ManaCost = 5
	c.ManaRequirement = []string{civ.Water}
	c.TapAbility = fortMegaclusterTapAbility

	c.Use(fx.Creature, fx.Evolution, fx.TapAbility,
		fx.When(fx.InTheBattlezone, func(card *match.Card, ctx *match.Context) {

			fx.GiveTapAbilityToAllies(
				card,
				ctx,
				func(x *match.Card) bool { return x.ID != card.ID && x.Civ == civ.Water },
				fortMegaclusterTapAbility,
			)

		}),
	)
}

func fortMegaclusterTapAbility(card *match.Card, ctx *match.Context) {
	card.Player.DrawCards(1)
	ctx.Match.ReportActionInChat(card.Player, fmt.Sprintf("%s activated %s's tap ability to draw 1 cards", card.Player.Username(), card.Name))
}
