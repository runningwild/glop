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
type Effect interface {
  // Can apply modifiers to BaseStats, but cannot actually change them, since we
  // need to keep the original data around
  ModifyStats(BaseStats) BaseStats

  // Can change DynamicStats however it wants, and may reference BaseStats to do
  // so.
  ModifyDynamicStats(*DynamicStats, BaseStats)

  // Given an amount of damage, returns a modified amount of damage that this
  // unit should take.
  ModifyDamage(int, BaseStats) int

  // Modifies an Attribute based on the terrain and its current value
  ModifyMovement(base.Terrain, int) int
  ModifyLos(base.Terrain, int) int
  ModifyAttack(base.Terrain, int) int
  ModifyDefense(base.Terrain, int) int

  // Called at the end of every round.
  Round()

  // Should return false when this effect is done so that it can be removed.
  Active() bool
}

type NullEffect struct {}
func (NullEffect) ModifyStats(b BaseStats) BaseStats { return b }
func (NullEffect) ModifyDynamicStats(*DynamicStats, BaseStats) { }
func (NullEffect) ModifyDamage(dmg int, b BaseStats) int { return dmg }
func (NullEffect) ModifyMovement(t base.Terrain, n int) int { return n }
func (NullEffect) ModifyLos(t base.Terrain, n int) int { return n }
func (NullEffect) ModifyAttack(t base.Terrain, n int) int { return n }
func (NullEffect) ModifyDefense(t base.Terrain, n int) int { return n }

// TimedEffect provides an easy way to make an effect last for a specific amount
// of time.  For an effect to only last until the end of the turn set
// rounds = 1, for it to last until the end of the next turn set rounds = 2
type TimedEffect int
func (e *TimedEffect) Round() {
  (*e)--
}
func (e *TimedEffect) Active() bool {
  return (*e) > 0
}

type Poison struct {
  NullEffect
  TimedEffect
  Power int
}
func (e *Poison) ModifyDynamicStats(d *DynamicStats, b BaseStats) {
  d.Health -= e.Power
}

type Stultify struct {
  NullEffect
  TimedEffect
  Power int
}
func (e *Stultify) ModifyDynamicStats(d *DynamicStats, b BaseStats) {
  d.Ap -= e.Power
}

type Confuse struct {
  NullEffect
  TimedEffect
  Power int
}
func (e *Confuse) ModifyStats(b BaseStats) BaseStats {
  b.Attack -= e.Power
  return b
}

type Blind struct {
  NullEffect
  TimedEffect
  Power int
}
func (e *Blind) ModifyStats(b BaseStats) BaseStats {
  b.LosDist -= e.Power
  return b
}

type Stun struct {
  NullEffect
  TimedEffect
  ApMod      int
  AttackMod  int
  DefenseMod int
}
func (e *Stun) ModifyStats(b BaseStats) BaseStats {
  b.Attack -= e.AttackMod
  b.Defense -= e.DefenseMod
  return b
}
func (e *Stun) ModifyDynamicStats(d *DynamicStats, b BaseStats) {
  d.Ap -= e.ApMod
}

type Slow struct {
  NullEffect
  TimedEffect
}
func (e *Slow) ModifyMovement(_ base.Terrain, cost int) int {
  return cost + 1
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
