package services

import (
	"errors"
	"fmt"
	"time"
)

type repoRateLimiter struct {
	// cacheMap contains a list of previous git requests that occured within a given time period; key is the repo URL.
	cacheMap map[string]*repoCacheEntry
}

type repoCacheEntry struct {
	requests []time.Time
}

const (
	MaxRequestsPerMinute = 10
	MaxDelayTime         = 5 * time.Second
)

func (cache *repoRateLimiter) EnsureGitRepoRateLimit(repoURL string) error {

	if repoURL == "" {
		return errors.New("repository resolved to an empty string")
	}

	// Get or create the cache entry if it doesn't exist
	cacheEntry := cache.cacheMap[repoURL]
	if cacheEntry == nil {
		cacheEntry = &repoCacheEntry{requests: []time.Time{}}
		cache.cacheMap[repoURL] = cacheEntry
	}

	// Add current request to the request list
	cacheEntry.requests = append(cacheEntry.requests, time.Now())

	fmt.Println("Initial request size", len(cacheEntry.requests))

	maxRequestsPerMinute := MaxRequestsPerMinute

	forLoopExpireTime := time.Now().Add(MaxDelayTime)

	for {
		oneMinuteAgo := time.Now().Add(-1 * time.Minute)

		// Remove requests that occurred more than one minute ago
		unexpired := []time.Time{}
		for _, req := range cacheEntry.requests {
			if req.After(oneMinuteAgo) {
				unexpired = append(unexpired, req)
			}
		}
		cacheEntry.requests = unexpired

		if len(cacheEntry.requests) > maxRequestsPerMinute {
			// If the number of requests that occured within the last minute is greater than the rate limit, then wait
			fmt.Println("Waiting.", len(unexpired))
			time.Sleep(1000 * time.Millisecond)
		} else {
			// If we are below the rate limit, then return
			break
		}

		// Wait at most 'forLoopExpireTime' seconds in the loop
		if time.Now().After(forLoopExpireTime) {
			break
		}
	}

	return nil
}
