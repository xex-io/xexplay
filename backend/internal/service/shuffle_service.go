package service

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type ShuffleService struct{}

func NewShuffleService() *ShuffleService {
	return &ShuffleService{}
}

// Shuffle returns a deterministic permutation of [0, cardCount) based on hash(userID + date).
// The same user on the same date always gets the same card order (Fisher-Yates shuffle).
func (s *ShuffleService) Shuffle(userID uuid.UUID, date time.Time, cardCount int) []int {
	// Build deterministic seed from userID + date
	dateStr := date.Format("2006-01-02")
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", userID.String(), dateStr)))
	seed := int64(binary.BigEndian.Uint64(hash[:8]))

	// Fisher-Yates shuffle with deterministic RNG
	rng := rand.New(rand.NewSource(seed))

	perm := make([]int, cardCount)
	for i := range perm {
		perm[i] = i
	}
	for i := cardCount - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		perm[i], perm[j] = perm[j], perm[i]
	}

	return perm
}
