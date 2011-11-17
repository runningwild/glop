package stats

import (
  "game/base"
  "glop/util/algorithm"
)

type DynamicStats struct {
  Health   int
  Ap       int
}
type BaseStats struct {
  DynamicStats
  Attack   int
  Defense  int
  LosDist  int
  Atts     []string
}

type Stats struct {
  base    BaseStats
  cur     DynamicStats
  attmap  map[string]Attributes
  effects []Effect
}

func (s *Stats) AddEffect(e Effect) {
  s.effects = append(s.effects, e)
}
func (s *Stats) BaseHealth() int {
  return s.base.Health
}
func (s *Stats) CurHealth() int {
  return s.cur.Health
}
func (s *Stats) BaseAp() int {
  return s.base.Ap
}
func (s *Stats) CurAp() int {
  return s.cur.Ap
}
func (s *Stats) BaseAttack() int {
  return s.base.Attack
}
func (s *Stats) CurAttack(t base.Terrain) int {
  attack := s.base.Attack
  attack += processAttributes(s.base.Atts, s.attmap).AttackMods[t]
  for _, effect := range s.effects {
    attack = effect.ModifyAttack(t, attack)
  }
  return attack
}
func (s *Stats) BaseDefense() int {
  return s.base.Defense
}
func (s *Stats) CurDefense(t base.Terrain) int {
  defense := s.base.Defense
  defense += processAttributes(s.base.Atts, s.attmap).DefenseMods[t]
  for _, effect := range s.effects {
    defense = effect.ModifyDefense(t, defense)
  }
  return defense
}
func (s *Stats) Concealment(t base.Terrain) int {
  return processAttributes(s.base.Atts, s.attmap).LosMods[t]
}
func (s *Stats) BaseLosDist() int {
  return s.base.LosDist
}
func (s *Stats) CurLosDist(t base.Terrain) int {
  los := s.base.LosDist
  for _, effect := range s.effects {
    los = effect.ModifyLos(t, los)
  }
  return los
}
func (s *Stats) SpendAp(amt int) bool {
  if amt > s.cur.Ap { return false }
  s.cur.Ap -= amt
  return true
}
func (s *Stats) Setup() {
  s.cur = s.base.DynamicStats
}
func (s *Stats) Round() {
  s.cur.Ap = s.base.Ap
  for i := range s.effects {
    s.effects[i].Round()
  }
  s.effects = algorithm.Choose(s.effects, func(a interface{}) bool {
    return a.(Effect).Active()
  }).([]Effect)
}
func (s *Stats) DoDamage(dmg int) {
  s.cur.Health -= dmg
  if s.cur.Health < 0 {
    s.cur.Health = 0
  }
}
func (s *Stats) MoveCost(t base.Terrain) int {
  move := 1
  move += processAttributes(s.base.Atts, s.attmap).MoveMods[t]
  for _, effect := range s.effects {
    move = effect.ModifyMovement(t, move)
  }
  return move
}
func MakeStats(base BaseStats, attmap map[string]Attributes) *Stats {
  return &Stats{
    base : base,
    cur : base.DynamicStats,
    attmap : attmap,
  }
}
