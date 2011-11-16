package stats

import "game/base"

type BaseStats struct {
  Health   int
  Ap       int
  Attack   int
  Defense  int
  Atts     []string
}

type Stats struct {
  base   BaseStats
  attmap map[string]Attributes
//buffs []Buff
}

func (s *Stats) BaseHealth() int {
  return 3
}
func (s *Stats) CurHealth() int {
  return 3
}
func (s *Stats) BaseAp() int {
  return 3
}
func (s *Stats) CurAp() int {
  return 3
}
func (s *Stats) BaseAttack() int {
  return 3
}
func (s *Stats) CurAttack(t base.Terrain) int {
  return 3
}
func (s *Stats) BaseDefense() int {
  return 3
}
func (s *Stats) CurDefense(t base.Terrain) int {
  return 3
}
func (s *Stats) Concealment(t base.Terrain) int {
  return 1
}
func (s *Stats) LosDistance() int {
  return 30
}
func (s *Stats) SpendAp(amt int) {
}
func (s *Stats) Setup() {
}
func (s *Stats) Round() {
}
func (s *Stats) DoDamage(dmg int) {
}
func (s *Stats) MoveCost(t base.Terrain) int {
  return 1
}
func MakeStats(base BaseStats, attmap map[string]Attributes) *Stats {
  return &Stats{}
}
