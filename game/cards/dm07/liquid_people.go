package dm07

import (
	"duel-masters/game/civ"
	"duel-masters/game/family"
	"duel-masters/game/fx"
	"duel-masters/game/match"
)

func AquaAgent(c *match.Card) {

	c.Name = "Aqua Agent"
	c.Power = 2000
	c.Civ = civ.Water
	c.Family = []string{family.LiquidPeople}
	c.ManaCost = 6
	c.ManaRequirement = []string{civ.Water}

	c.Use(fx.Creature, fx.WaterStealth, fx.When(fx.WouldBeDestroyed, fx.MayReturnToHand))
}

func AquaFencer(c *match.Card) {

	c.Name = "Aqua Fencer"
	c.Power = 3000
	c.Civ = civ.Water
	c.Family = []string{family.LiquidPeople}
	c.ManaCost = 7
	c.ManaRequirement = []string{civ.Water}
	c.TapAbility = fx.ReturnOpCardFromMZToHand

	c.Use(fx.Creature, fx.TapAbility)
}
