package otp

import (
	"sort"
	"sync"
	"time"
)

type MemoryStore struct {
	mu         sync.RWMutex
	challenges map[string]*Challenge
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		challenges: make(map[string]*Challenge),
	}
}

func (ms *MemoryStore) Create(challenge *Challenge) (*Challenge, error) {
	if challenge == nil {
		return nil, ErrNotFound
	}

	intent, err := NormalizeIntent(challenge.Intent)
	if err != nil {
		return nil, err
	}

	normalizedEmail := NormalizeEmail(challenge.Email)

	ms.mu.Lock()
	defer ms.mu.Unlock()

	copy := cloneChallenge(challenge)
	copy.Intent = intent
	copy.Email = normalizedEmail
	ms.challenges[copy.ID] = copy

	return cloneChallenge(copy), nil
}

func (ms *MemoryStore) GetByID(id string) (*Challenge, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	challenge, ok := ms.challenges[id]
	if !ok {
		return nil, ErrNotFound
	}

	return cloneChallenge(challenge), nil
}

func (ms *MemoryStore) GetLatestByEmailIntent(email, intent string) (*Challenge, error) {
	normalizedEmail := NormalizeEmail(email)
	normalizedIntent, err := NormalizeIntent(intent)
	if err != nil {
		return nil, err
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	matches := make([]*Challenge, 0)
	for _, challenge := range ms.challenges {
		if challenge.Email == normalizedEmail && challenge.Intent == normalizedIntent {
			matches = append(matches, challenge)
		}
	}

	if len(matches) == 0 {
		return nil, ErrNotFound
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].CreatedAt.Equal(matches[j].CreatedAt) {
			return matches[i].ID > matches[j].ID
		}
		return matches[i].CreatedAt.After(matches[j].CreatedAt)
	})

	return cloneChallenge(matches[0]), nil
}

func (ms *MemoryStore) IncrementAttempts(id string, updatedAt time.Time) (*Challenge, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	challenge, ok := ms.challenges[id]
	if !ok {
		return nil, ErrNotFound
	}

	challenge.AttemptCount++
	challenge.UpdatedAt = updatedAt.UTC()

	return cloneChallenge(challenge), nil
}

func (ms *MemoryStore) IncrementResendCount(id string, updatedAt time.Time) (*Challenge, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	challenge, ok := ms.challenges[id]
	if !ok {
		return nil, ErrNotFound
	}

	challenge.ResendCount++
	challenge.UpdatedAt = updatedAt.UTC()

	return cloneChallenge(challenge), nil
}

func (ms *MemoryStore) MarkVerifiedAndUsed(id string, verifiedAt time.Time) (*Challenge, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	challenge, ok := ms.challenges[id]
	if !ok {
		return nil, ErrNotFound
	}

	verifiedAtUTC := verifiedAt.UTC()
	challenge.VerifiedAt = &verifiedAtUTC
	challenge.UsedAt = &verifiedAtUTC
	challenge.UpdatedAt = verifiedAtUTC

	return cloneChallenge(challenge), nil
}

func (ms *MemoryStore) DeleteExpired(now time.Time) (int, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	cutoff := now.UTC()
	deleted := 0
	for id, challenge := range ms.challenges {
		if challenge.ExpiresAt.Before(cutoff) {
			delete(ms.challenges, id)
			deleted++
		}
	}

	return deleted, nil
}
