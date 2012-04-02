// In cases where Go's garbage collecter is unresponsive, or even broken, this
// allows you to easily manage blocks of memory.
package memory

import (
  "fmt"
  "sync"
)

type Manager struct {
  mutex  sync.Mutex
  blocks [][][]byte
  used   map[*byte]bool
}
func NewManager() *Manager {
  var m Manager
  m.blocks = make([][][]byte, 21)
  m.used = make(map[*byte]bool)
  return &m
}
func (m *Manager) GetBlock(n int) []byte {
  m.mutex.Lock()
  defer m.mutex.Unlock()
  c := 1024
  s := 0
  for c < n {
    c *= 2
    s++
  }
  for i := range m.blocks[s] {
    if !m.used[&m.blocks[s][i][0]] {
      m.used[&m.blocks[s][i][0]] = true
      for j := range m.blocks[s][i] {
        m.blocks[s][i][j] = 0
      }
      return m.blocks[s][i]
    }
  }
  new_block := make([]byte, c)
  m.blocks[s] = append(m.blocks[s], new_block)
  m.used[&new_block[0]] = true
  return new_block[0:n]
}
func (m *Manager) FreeBlock(b []byte) {
  m.mutex.Lock()
  defer m.mutex.Unlock()
  if _, ok := m.used[&b[0]]; !ok {
    panic("Tried to free an unused block")
  }
  delete(m.used, &b[0])
}
func (m *Manager) TotalAllocations() string {
  c := 1024
  var ret string
  total_used := 0
  total_allocated := 0
  for s := range m.blocks {
    used := 0
    for i := range m.blocks[s] {
      if m.used[&m.blocks[s][i][0]] {
        used++
      }
    }
    if used > 0 {
      ret += fmt.Sprintf("%d bytes: %d/%d blocks in use.\n", c, used, len(m.blocks[s]))
    }
    total_used += used * c
    total_allocated += len(m.blocks[s]) * c
    c *= 2
  }
  ret += fmt.Sprintf("Total memory used/allocated: %d/%d\n", total_used, total_allocated)
  return ret
}

var manager *Manager
func init() {
  manager = NewManager()
}
func GetBlock(n int) []byte {
  return manager.GetBlock(n)
}
func FreeBlock(b []byte) {
  manager.FreeBlock(b)
}
func TotalAllocations() string {
  return manager.TotalAllocations()
}