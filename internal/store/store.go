package store

import (
	"sync"
	"time"
)

// ShortURL 短链接实体
type ShortURL struct {
	Code       string    `json:"code"`
	LongURL    string    `json:"long_url"`
	CreatedAt  time.Time `json:"created_at"`
	AccessCount int64    `json:"access_count"`
}

// MemoryStore 基于内存的短链接存储（读多写少，用 RWMutex）
type MemoryStore struct {
	mu    sync.RWMutex
	codes map[string]*ShortURL // code -> entity
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		codes: make(map[string]*ShortURL),
	}
}

// Save 保存新短链接
func (s *MemoryStore) Save(entry *ShortURL) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[entry.Code] = entry
}

// Find 按 code 查找
func (s *MemoryStore) Find(code string) (*ShortURL, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.codes[code]
	return entry, ok
}

// IncrementAccess 原子增加访问计数
func (s *MemoryStore) IncrementAccess(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.codes[code]; ok {
		entry.AccessCount++
	}
}

// Count 返回存储的短链接总数
func (s *MemoryStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.codes)
}
