package stats

import (
  "game/base"
  "glop/util/algorithm"
  "gob"
)

type Stats interface {
  BaseHealth() int
  BaseAp() int
  BaseAttack() int
  BaseDefense() int
  BaseLosDist() int
  AddEffect(e Effect, apply_now bool)
  CurHealth() int
  CurAp() int
  CurAttack(t base.Terrain) int
  CurDefense(t base.Terrain) int
  Concealment(t base.Terrain) int
  CurLosDist(t base.Terrain) int
  SpendAp(amt int) bool
  Setup()
  Round()
  DoDamage(dmg int)
  MoveCost(t base.Terrain) int
}

type dynamicStats struct {
  Health   int
  Ap       int
}

type baseStats struct {
  Health   int
  Ap       int
  Attack   int
  Defense  int
  LosDist  int
  Atts     []string
}

func (bs *baseStats) dynamic() dynamicStats {
  return dynamicStats{
    Health: bs.Health,
    Ap: bs.Ap,
  }
}

type stats struct {
  Base    baseStats
  Cur     dynamicStats
  Effects []Effect
}
func init() {
  gob.Register(&stats{})
}

// Global map from Attribute name to Attribute
var attmap map[string]Attributes
func SetAttmap(_attmap map[string]Attributes) {
  attmap = _attmap
}

func (s *stats) BaseHealth() int {
  return s.Base.Health
}
func (s *stats) BaseAp() int {
  return s.Base.Ap
}
func (s *stats) BaseAttack() int {
  return s.Base.Attack
}
func (s *stats) BaseDefense() int {
  return s.Base.Defense
}
func (s *stats) BaseLosDist() int {
  return s.Base.LosDist
}

func (s *stats) AddEffect(e Effect, apply_now bool) {
  if apply_now {
    e.ModifyDynamicStats(&s.Cur, s.Base)
    if s.Cur.Health > s.BaseHealth() {
      s.Cur.Health = s.BaseHealth()
    }
    if !e.Active() {
      return
    }
  }
  for i := range s.Effects {
    if s.Effects[i].Name() == e.Name() {
      s.Effects[i] = e
      return
    }
  }
  s.Effects = append(s.Effects, e)
}
func (s *stats) CurHealth() int {
  return s.Cur.Health
}
func (s *stats) CurAp() int {
  return s.Cur.Ap
}
func (s *stats) CurAttack(t base.Terrain) int {
  attack := s.BaseAttack()
  attack += processAttributes(s.Base.Atts).AttackMods[t]
  for _, effect := range s.Effects {
    attack = effect.ModifyAttack(t, attack)
  }
  return attack
}
func (s *stats) CurDefense(t base.Terrain) int {
  defense := s.BaseDefense()
  defense += processAttributes(s.Base.Atts).DefenseMods[t]
  for _, effect := range s.Effects {
    defense = effect.ModifyDefense(t, defense)
  }
  return defense
}
func (s *stats) Concealment(t base.Terrain) int {
  return processAttributes(s.Base.Atts).LosMods[t]
}
func (s *stats) CurLosDist(t base.Terrain) int {
  los := s.BaseLosDist()
  for _, effect := range s.Effects {
    los = effect.ModifyLos(t, los)
  }
  return los
}
func (s *stats) SpendAp(amt int) bool {
  if amt > s.Cur.Ap { return false }
  s.Cur.Ap -= amt
  return true
}
func (s *stats) Setup() {
  s.Cur = s.Base.dynamic()
}
func (s *stats) Round() {
  s.Cur.Ap = s.BaseAp()
  for i := range s.Effects {
    s.Effects[i].ModifyDynamicStats(&s.Cur, s.Base)
    s.Effects[i].Round()
  }
  if s.Cur.Health > s.BaseHealth() {
    s.Cur.Health = s.BaseHealth()
  }
  s.Effects = algorithm.Choose(s.Effects, func(a interface{}) bool {
    return a.(Effect).Active()
  }).([]Effect)
}
func (s *stats) DoDamage(dmg int) {
  for _,effect := range s.Effects {
    dmg = effect.ModifyIncomingDamage(dmg, s.Base)
  }
  s.Cur.Health -= dmg
  if s.Cur.Health < 0 {
    s.Cur.Health = 0
  }
}
func (s *stats) MoveCost(t base.Terrain) int {
  move := 1
  move += processAttributes(s.Base.Atts).MoveMods[t]
  for _, effect := range s.Effects {
    move = effect.ModifyMovement(t, move)
  }
  return move
}

func MakeStats(Health, Ap, Attack, Defense, LosDist int, Atts []string) Stats {
  base := baseStats{
    Health: Health,
    Ap: Ap,
    Attack: Attack,
    Defense: Defense,
    LosDist: LosDist,
    Atts: Atts,
  }
  return &stats{
    Base : base,
    Cur : base.dynamic(),
  }
}
