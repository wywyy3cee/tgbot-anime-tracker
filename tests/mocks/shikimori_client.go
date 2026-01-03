package mocks

import (
	"fmt"
	"sync"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

type MockShikimoriClient struct {
	mu              sync.RWMutex
	searchResults   map[string][]models.Anime
	animeDetails    map[int]*models.Anime
	searchCallCount int
	getByIDCallCount int
}

func NewMockShikimoriClient() *MockShikimoriClient {
	return &MockShikimoriClient{
		searchResults: make(map[string][]models.Anime),
		animeDetails:  make(map[int]*models.Anime),
	}
}

func (m *MockShikimoriClient) SearchAnime(query string, limit int) ([]models.Anime, error) {
	m.mu.Lock()
	m.searchCallCount++
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if results, ok := m.searchResults[query]; ok && len(results) > 0 {
		// Return limited results
		if limit > 0 && len(results) > limit {
			return results[:limit], nil
		}
		return results, nil
	}

	return nil, fmt.Errorf("no animes found for query: %s", query)
}

func (m *MockShikimoriClient) GetAnimeById(id int) (*models.Anime, error) {
	m.mu.Lock()
	m.getByIDCallCount++
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if anime, ok := m.animeDetails[id]; ok {
		return anime, nil
	}

	return nil, fmt.Errorf("anime with id %d not found", id)
}

func (m *MockShikimoriClient) SetSearchResults(query string, animes []models.Anime) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.searchResults[query] = animes
}

func (m *MockShikimoriClient) SetAnimeDetails(id int, anime *models.Anime) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.animeDetails[id] = anime
}

func (m *MockShikimoriClient) ResetCallCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.searchCallCount = 0
	m.getByIDCallCount = 0
}

func (m *MockShikimoriClient) GetSearchCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.searchCallCount
}

func (m *MockShikimoriClient) GetByIDCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getByIDCallCount
}
