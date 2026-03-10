package cleaner

import "sync"

var analyzeDiscoveryCache struct {
	mu              sync.RWMutex
	sessions        map[string][]ConversationSession
	leftovers       map[string][]Candidate
	loadedSessions  map[string]bool
	loadedLeftovers map[string]bool
}

func resetAnalyzeDiscoveryCache() {
	analyzeDiscoveryCache.mu.Lock()
	defer analyzeDiscoveryCache.mu.Unlock()
	analyzeDiscoveryCache.sessions = map[string][]ConversationSession{}
	analyzeDiscoveryCache.leftovers = map[string][]Candidate{}
	analyzeDiscoveryCache.loadedSessions = map[string]bool{}
	analyzeDiscoveryCache.loadedLeftovers = map[string]bool{}
}

func discoverAssistantSessionsCached(assistant string) ([]ConversationSession, error) {
	analyzeDiscoveryCache.mu.RLock()
	if analyzeDiscoveryCache.loadedSessions[assistant] {
		cached := append([]ConversationSession(nil), analyzeDiscoveryCache.sessions[assistant]...)
		analyzeDiscoveryCache.mu.RUnlock()
		return cached, nil
	}
	analyzeDiscoveryCache.mu.RUnlock()

	sessions, err := discoverAssistantSessions(assistant)
	if err != nil {
		return nil, err
	}

	analyzeDiscoveryCache.mu.Lock()
	analyzeDiscoveryCache.sessions[assistant] = append([]ConversationSession(nil), sessions...)
	analyzeDiscoveryCache.loadedSessions[assistant] = true
	analyzeDiscoveryCache.mu.Unlock()

	return append([]ConversationSession(nil), sessions...), nil
}

func discoverAssistantLeftoversCached(assistant string) ([]Candidate, error) {
	analyzeDiscoveryCache.mu.RLock()
	if analyzeDiscoveryCache.loadedLeftovers[assistant] {
		cached := append([]Candidate(nil), analyzeDiscoveryCache.leftovers[assistant]...)
		analyzeDiscoveryCache.mu.RUnlock()
		return cached, nil
	}
	analyzeDiscoveryCache.mu.RUnlock()

	leftovers, err := discoverAssistantLeftovers(assistant)
	if err != nil {
		return nil, err
	}

	analyzeDiscoveryCache.mu.Lock()
	analyzeDiscoveryCache.leftovers[assistant] = append([]Candidate(nil), leftovers...)
	analyzeDiscoveryCache.loadedLeftovers[assistant] = true
	analyzeDiscoveryCache.mu.Unlock()

	return append([]Candidate(nil), leftovers...), nil
}

func invalidateAssistantSessionsCache(assistant string) {
	analyzeDiscoveryCache.mu.Lock()
	defer analyzeDiscoveryCache.mu.Unlock()
	delete(analyzeDiscoveryCache.sessions, assistant)
	delete(analyzeDiscoveryCache.loadedSessions, assistant)
}

func invalidateAssistantLeftoversCache(assistant string) {
	analyzeDiscoveryCache.mu.Lock()
	defer analyzeDiscoveryCache.mu.Unlock()
	delete(analyzeDiscoveryCache.leftovers, assistant)
	delete(analyzeDiscoveryCache.loadedLeftovers, assistant)
}
