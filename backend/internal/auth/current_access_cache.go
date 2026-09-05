package auth

import (
	"context"
	"sync"
	"time"
)

const defaultCurrentAccessCacheTTL = 2 * time.Second

type cachedCurrentAccessEntry struct {
	access    *CurrentAccess
	expiresAt time.Time
}

type cachedCurrentAccessResolver struct {
	inner   CurrentAccessResolver
	ttl     time.Duration
	mu      sync.Mutex
	byStaff map[uint64]cachedCurrentAccessEntry
}

// NewCachedCurrentAccessResolver memoizes Resolve by staff ID for ttl.
// Failed lookups are not cached. Callers must treat the returned snapshot as
// request-local; slices are copied on hit.
func NewCachedCurrentAccessResolver(
	inner CurrentAccessResolver,
	ttl time.Duration,
) CurrentAccessResolver {
	if ttl <= 0 {
		ttl = defaultCurrentAccessCacheTTL
	}
	return &cachedCurrentAccessResolver{
		inner:   inner,
		ttl:     ttl,
		byStaff: make(map[uint64]cachedCurrentAccessEntry),
	}
}

func (c *cachedCurrentAccessResolver) Resolve(
	ctx context.Context,
	staffID uint64,
) (*CurrentAccess, error) {
	now := time.Now()
	c.mu.Lock()
	if entry, ok := c.byStaff[staffID]; ok && now.Before(entry.expiresAt) {
		cloned := cloneCurrentAccess(entry.access)
		c.mu.Unlock()
		return cloned, nil
	}
	c.mu.Unlock()

	access, err := c.inner.Resolve(ctx, staffID)
	if err != nil {
		c.mu.Lock()
		delete(c.byStaff, staffID)
		c.mu.Unlock()
		return nil, err
	}

	stored := cloneCurrentAccess(access)
	c.mu.Lock()
	c.byStaff[staffID] = cachedCurrentAccessEntry{
		access:    stored,
		expiresAt: now.Add(c.ttl),
	}
	c.mu.Unlock()
	return cloneCurrentAccess(access), nil
}

func cloneCurrentAccess(access *CurrentAccess) *CurrentAccess {
	if access == nil {
		return nil
	}
	cloned := *access
	cloned.ClinicIDs = append([]uint64(nil), access.ClinicIDs...)
	return &cloned
}
