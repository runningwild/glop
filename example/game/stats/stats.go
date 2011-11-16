package stats

import "game/base"
import "fmt"

type DynamicStats struct {
  Health   int
  Ap       int
}
type BaseStats struct {
  DynamicStats
  Attack   int
  Defense  int
  Atts     []string
}
type Stats struct {
  base  BaseStats
  cur   DynamicStats
  attmap map[string]Attributes
//buffs []Buff
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
  return s.base.Attack + processAttributes(s.base.Atts, s.attmap).AttackMods[t]
}
func (s *Stats) BaseDefense() int {
  return s.base.Defense
}
func (s *Stats) CurDefense(t base.Terrain) int {
  return s.base.Defense + processAttributes(s.base.Atts, s.attmap).DefenseMods[t]
}
func (s *Stats) Concealment(t base.Terrain) int {
  return processAttributes(s.base.Atts, s.attmap).LosMods[t]
}
func (s *Stats) LosDistance() int {
  return processAttributes(s.base.Atts, s.attmap).LosDistance
}
func (s *Stats) SpendAp(amt int) bool {
  if amt > s.cur.Ap { return false }
  s.cur.Ap -= amt
  return true
}
func (s *Stats) Setup() {
  s.cur = s.base.DynamicStats
  fmt.Printf("stats: %v\n", s)
}
func (s *Stats) Round() {
  s.cur.Ap = s.base.Ap
}
func (s *Stats) DoDamage(dmg int) {
  s.cur.Health -= dmg
  if s.cur.Health < 0 {
    s.cur.Health = 0
  }
}
func (s *Stats) MoveCost(t base.Terrain) int {
  r,ok := processAttributes(s.base.Atts, s.attmap).MoveMods[t]
  if !ok {
    return -1
  }
  return r
}
func MakeStats(base BaseStats, attmap map[string]Attributes) *Stats {
  return &Stats{
    base : base,
    cur : base.DynamicStats,
    attmap : attmap,
  }
}
