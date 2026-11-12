// ct-----------------------------------------
// @file      : matcher.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/01/24 21:00
// -------------------------------------------
// CT log monitoring SDK - high-performance domain matcher
// ============================================

package ct

import (
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/models"
	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/types"
	"golang.org/x/net/publicsuffix"
)

// DomainMatcher is the high-performance domain matcher
type DomainMatcher struct {
	config      *types.DomainMatcherConfig
	trieRoot    *models.DomainTrieNode
	bloomFilter *BloomFilter
	lruCache    *LRUCache
	domains     map[string]*models.Domain
	mu          sync.RWMutex

	// Statistics
	stats struct {
		totalMatches int64
		cacheHits    int64
		cacheMisses  int64
		bloomHits    int64
		bloomMisses  int64
		trieHits     int64
		trieMisses   int64
	}
}

// NewDomainMatcher creates a new domain matcher
func NewDomainMatcher(config *types.DomainMatcherConfig) *DomainMatcher {
	dm := &DomainMatcher{
		config:   config,
		trieRoot: models.NewDomainTrieNode(),
		domains:  make(map[string]*models.Domain),
	}

	// Initialize the bloom filter
	if config.EnableBloomFilter {
		dm.bloomFilter = NewBloomFilter(config.BloomFilterSize, config.BloomFilterHash)
	}

	// Initialize the LRU cache
	if config.EnableLRUCache {
		dm.lruCache = NewLRUCache(config.LRUCacheSize)
	}

	return dm
}

// AddDomain adds a domain to the matcher
func (dm *DomainMatcher) AddDomain(domain string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	domain = dm.normalizeDomain(domain)
	if domain == "" {
		return fmt.Errorf("invalid domain: %s", domain)
	}

	// Check whether it already exists
	if _, exists := dm.domains[domain]; exists {
		return nil
	}

	// Create the domain object
	domainObj := models.NewDomain(domain)

	// Add to the domain map
	dm.domains[domain] = domainObj

	// Add to the trie
	dm.addToTrie(domain, domainObj)

	// Add to the bloom filter
	if dm.bloomFilter != nil {
		dm.bloomFilter.Add(domain)
	}

	return nil
}

// RemoveDomain removes a domain from the matcher
func (dm *DomainMatcher) RemoveDomain(domain string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	domain = dm.normalizeDomain(domain)
	if domain == "" {
		return nil
	}

	// Remove from the domain map
	delete(dm.domains, domain)

	// Remove from the trie
	dm.removeFromTrie(domain)

	// Rebuild the bloom filter (simplified implementation)
	if dm.bloomFilter != nil {
		dm.rebuildBloomFilter()
	}

	// Clear the LRU cache
	if dm.lruCache != nil {
		dm.lruCache.Clear()
	}

	return nil
}

// HasDomain checks whether the given domain is contained
func (dm *DomainMatcher) HasDomain(domain string) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	domain = dm.normalizeDomain(domain)
	_, exists := dm.domains[domain]
	return exists
}

// GetAllDomains fetches all domains
func (dm *DomainMatcher) GetAllDomains() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	domains := make([]string, 0, len(dm.domains))
	for domain := range dm.domains {
		domains = append(domains, domain)
	}
	return domains
}

// MatchSubdomains matches subdomains (core method)
func (dm *DomainMatcher) MatchSubdomains(domains []string) []*models.DomainMatch {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var matches []*models.DomainMatch

	for _, domain := range domains {
		domain = dm.normalizeDomain(domain)
		if domain == "" {
			continue
		}

		// Extract the root domain (the input may be a subdomain; we match the root domain)
		rootDomain := extractRootDomain(domain)
		if rootDomain == "" {
			continue
		}

		// First filter layer: fast bloom filter check
		if dm.bloomFilter != nil {
			if !dm.bloomFilter.Contains(rootDomain) {
				dm.stats.bloomMisses++
				continue
			}
			dm.stats.bloomHits++
		}

		// Second filter layer: LRU cache check
		if dm.lruCache != nil {
			if _, found := dm.lruCache.Get(domain); found {
				dm.stats.cacheHits++
				continue
			}
			dm.stats.cacheMisses++
		}

		// Third filter layer: exact trie match
		if match := dm.matchInTrie(rootDomain); match != nil {
			// Update the subdomain info in the match result
			match.Subdomain = domain
			dm.stats.trieHits++
			dm.stats.totalMatches++

			matches = append(matches, match)

			// Add to the cache
			if dm.lruCache != nil {
				dm.lruCache.Put(domain, match)
			}
		} else {
			dm.stats.trieMisses++
		}
	}

	return matches
}

// IsSubdomainOf determines whether a domain is a subdomain of the given root domain
func (dm *DomainMatcher) IsSubdomainOf(subdomain, rootDomain string) bool {
	subdomain = dm.normalizeDomain(subdomain)
	rootDomain = dm.normalizeDomain(rootDomain)

	if subdomain == "" || rootDomain == "" {
		return false
	}

	// Exact match
	if subdomain == rootDomain {
		return false
	}

	// Check whether it ends with the root domain
	if !strings.HasSuffix(subdomain, "."+rootDomain) {
		return false
	}

	// Validate the subdomain format
	subParts := strings.Split(subdomain, ".")
	rootParts := strings.Split(rootDomain, ".")

	// A subdomain must have at least one more label than the root domain
	return len(subParts) > len(rootParts)
}

// GetStats fetches statistics
func (dm *DomainMatcher) GetStats() *models.DomainMatcherStats {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	stats := &models.DomainMatcherStats{
		DomainStats: models.DomainStats{
			TotalDomains:    int64(len(dm.domains)),
			WildcardDomains: 0,
		},
	}

	// Count wildcard domains
	for _, domain := range dm.domains {
		if domain.IsWildcard {
			stats.WildcardDomains++
		}
	}

	// Bloom filter statistics
	if dm.bloomFilter != nil {
		stats.BloomFilterStats = models.BloomFilterStats{
			Size:      dm.bloomFilter.size,
			HashCount: dm.bloomFilter.hashCount,
			ItemCount: dm.bloomFilter.itemCount,
		}
	}

	// Trie statistics
	stats.TrieStats = dm.getTrieStats()

	// Cache statistics
	if dm.lruCache != nil {
		stats.CacheStats = models.CacheStats{
			Size:      dm.lruCache.capacity,
			ItemCount: len(dm.lruCache.items),
			HitCount:  dm.stats.cacheHits,
			MissCount: dm.stats.cacheMisses,
		}
		if dm.stats.cacheHits+dm.stats.cacheMisses > 0 {
			stats.CacheStats.HitRate = float64(dm.stats.cacheHits) / float64(dm.stats.cacheHits+dm.stats.cacheMisses)
		}
	}

	return stats
}

// Clear clears all domains
func (dm *DomainMatcher) Clear() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.domains = make(map[string]*models.Domain)
	dm.trieRoot = models.NewDomainTrieNode()

	if dm.bloomFilter != nil {
		dm.bloomFilter.Clear()
	}

	if dm.lruCache != nil {
		dm.lruCache.Clear()
	}

	// Reset the statistics
	dm.stats = struct {
		totalMatches int64
		cacheHits    int64
		cacheMisses  int64
		bloomHits    int64
		bloomMisses  int64
		trieHits     int64
		trieMisses   int64
	}{}

	return nil
}

// Private methods

// normalizeDomain normalizes a domain
func (dm *DomainMatcher) normalizeDomain(domain string) string {
	return models.NormalizeDomain(domain, dm.config.CaseSensitive)
}

// addToTrie adds a domain to the trie
func (dm *DomainMatcher) addToTrie(domain string, domainObj *models.Domain) {
	node := dm.trieRoot
	runes := []rune(domain)

	for _, r := range runes {
		if node.Children[r] == nil {
			node.Children[r] = models.NewDomainTrieNode()
		}
		node = node.Children[r]
	}

	node.IsEnd = true
	node.Domain = domain
	node.AddedAt = time.Now()
	node.DomainObj = domainObj
}

// removeFromTrie removes a domain from the trie
func (dm *DomainMatcher) removeFromTrie(domain string) {
	// Simplified implementation: rebuild the whole trie
	dm.rebuildTrie()
}

// rebuildTrie rebuilds the trie
func (dm *DomainMatcher) rebuildTrie() {
	dm.trieRoot = models.NewDomainTrieNode()
	for domain, domainObj := range dm.domains {
		dm.addToTrie(domain, domainObj)
	}
}

// rebuildBloomFilter rebuilds the bloom filter
func (dm *DomainMatcher) rebuildBloomFilter() {
	if dm.bloomFilter == nil {
		return
	}

	dm.bloomFilter.Clear()
	for domain := range dm.domains {
		dm.bloomFilter.Add(domain)
	}
}

// matchInTrie matches the root domain in the trie
func (dm *DomainMatcher) matchInTrie(rootDomain string) *models.DomainMatch {
	// Look up the root domain in the trie
	if node := dm.searchTrie(rootDomain); node != nil && node.IsEnd {
		node.HitCount++
		return &models.DomainMatch{
			Subdomain:  "", // Set at the call site
			RootDomain: rootDomain,
			MatchedAt:  time.Now(),
			MatchType:  "root_domain",
			Confidence: 1.0,
			HitCount:   node.HitCount,
		}
	}

	return nil
}

// searchTrie searches the trie
func (dm *DomainMatcher) searchTrie(domain string) *models.DomainTrieNode {
	node := dm.trieRoot
	runes := []rune(domain)

	for _, r := range runes {
		if node.Children[r] == nil {
			return nil
		}
		node = node.Children[r]
	}

	return node
}

// getTrieStats fetches trie statistics
func (dm *DomainMatcher) getTrieStats() models.TrieStats {
	stats := models.TrieStats{}
	dm.traverseTrie(dm.trieRoot, 0, &stats)
	return stats
}

// traverseTrie traverses the trie for statistics
func (dm *DomainMatcher) traverseTrie(node *models.DomainTrieNode, depth int, stats *models.TrieStats) {
	stats.NodeCount++

	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}

	if node.IsEnd {
		stats.DomainCount++
		stats.TotalHitCount += node.HitCount
	}

	for _, child := range node.Children {
		dm.traverseTrie(child, depth+1, stats)
	}
}

// BloomFilter is a simple bloom filter implementation
type BloomFilter struct {
	bitset    []bool
	hashes    []uint
	size      uint
	hashCount uint
	itemCount uint64
}

func NewBloomFilter(size, hashCount uint) *BloomFilter {
	return &BloomFilter{
		bitset:    make([]bool, size),
		hashes:    make([]uint, hashCount),
		size:      size,
		hashCount: hashCount,
	}
}

func (bf *BloomFilter) Add(item string) {
	for i := uint(0); i < bf.hashCount; i++ {
		hash := bf.hash(item, i)
		bf.bitset[hash%bf.size] = true
	}
	bf.itemCount++
}

func (bf *BloomFilter) Contains(item string) bool {
	for i := uint(0); i < bf.hashCount; i++ {
		hash := bf.hash(item, i)
		if !bf.bitset[hash%bf.size] {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) Clear() {
	bf.bitset = make([]bool, bf.size)
	bf.itemCount = 0
}

// hash uses double hashing to generate multiple independent hash values
func (bf *BloomFilter) hash(item string, seed uint) uint {
	// Use FNV-1a as the base hash
	h1 := fnv.New32a()
	h1.Write([]byte(item))
	hash1 := uint64(h1.Sum32())

	// Use a second different hash function
	h2 := fnv.New32()
	h2.Write([]byte(item))
	hash2 := uint64(h2.Sum32())

	// Double hashing: hash_i = hash1 + i * hash2
	// This generates multiple independent hash values
	hash := hash1 + uint64(seed)*hash2
	return uint(hash % uint64(bf.size))
}

// extractRootDomain extracts the root (registered) domain of a domain
// For example: www.example.com -> example.com, sub.example.co.uk -> example.co.uk
func extractRootDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return ""
	}

	// Use the publicsuffix library to obtain the valid suffix
	suffix, icann := publicsuffix.PublicSuffix(domain)
	if suffix == "" || !icann {
		// If the suffix cannot be determined, fall back to the last two labels
		parts := strings.Split(domain, ".")
		if len(parts) >= 2 {
			return strings.Join(parts[len(parts)-2:], ".")
		}
		return domain
	}

	// Fetch the root (registered) domain
	root, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		// If it fails, fall back to the last two labels
		parts := strings.Split(domain, ".")
		if len(parts) >= 2 {
			return strings.Join(parts[len(parts)-2:], ".")
		}
		return domain
	}

	return root
}

// LRUCache is a simple LRU cache implementation
type LRUCache struct {
	capacity int
	items    map[string]interface{}
	keys     []string
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]interface{}),
		keys:     make([]string, 0, capacity),
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	if value, exists := c.items[key]; exists {
		// Move to the end (most recently used)
		c.moveToEnd(key)
		return value, true
	}
	return nil, false
}

func (c *LRUCache) Put(key string, value interface{}) {
	if _, exists := c.items[key]; exists {
		c.items[key] = value
		c.moveToEnd(key)
		return
	}

	// Check capacity
	if len(c.items) >= c.capacity {
		// Evict the least recently used item
		oldest := c.keys[0]
		delete(c.items, oldest)
		c.keys = c.keys[1:]
	}

	c.items[key] = value
	c.keys = append(c.keys, key)
}

func (c *LRUCache) Clear() {
	c.items = make(map[string]interface{})
	c.keys = c.keys[:0]
}

func (c *LRUCache) moveToEnd(key string) {
	for i, k := range c.keys {
		if k == key {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			c.keys = append(c.keys, key)
			break
		}
	}
}
