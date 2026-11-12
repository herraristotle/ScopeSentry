// ct/models-----------------------------------
// @file      : domain.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/01/24 21:00
// -------------------------------------------
// CT log monitoring SDK - domain-related models
// ============================================

package models

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// Domain is the domain object
type Domain struct {
	Name       string    `json:"name"`        // Domain name
	IsWildcard bool      `json:"is_wildcard"` // Whether it is a wildcard domain
	RootDomain string    `json:"root_domain"` // Root domain
	Subdomains []string  `json:"subdomains"`  // Subdomain list (serialization only)
	AddedAt    time.Time `json:"added_at"`    // Add time
	LastSeen   time.Time `json:"last_seen"`   // Last discovery time
	CertCount  int64     `json:"cert_count"`  // Number of associated certificates
	HitCount   int64     `json:"hit_count"`   // Hit count
}

// NewDomain creates a new domain object
func NewDomain(name string) *Domain {
	now := time.Now()
	return &Domain{
		Name:     name,
		AddedAt:  now,
		LastSeen: now,
	}
}

// IsSubdomainOf determines whether this is a subdomain of the given domain
func (d *Domain) IsSubdomainOf(rootDomain string) bool {
	if d.Name == rootDomain {
		return false // The same domain does not count as a subdomain
	}

	// Check whether it ends with the root domain
	if !strings.HasSuffix(d.Name, "."+rootDomain) {
		return false
	}

	// Validate the subdomain format
	subParts := strings.Split(d.Name, ".")
	rootParts := strings.Split(rootDomain, ".")

	// A subdomain must have at least one more label than the root domain
	return len(subParts) > len(rootParts)
}

// UpdateLastSeen updates the last discovery time
func (d *Domain) UpdateLastSeen() {
	d.LastSeen = time.Now()
	d.HitCount++
}

// IncrementCertCount increments the certificate count
func (d *Domain) IncrementCertCount() {
	d.CertCount++
}

// DomainMatch is the domain match result
type DomainMatch struct {
	Subdomain   string    `json:"subdomain"`   // Discovered subdomain
	RootDomain  string    `json:"root_domain"` // Matched root domain
	Certificate *CertInfo `json:"certificate"` // Certificate information
	MatchedAt   time.Time `json:"matched_at"`  // Match time
	MatchType   string    `json:"match_type"`  // Match type (exact/prefix/suffix)
	Confidence  float64   `json:"confidence"`  // Match confidence 0-1
	HitCount    int64     `json:"hit_count"`   // Hit count
}

// NewDomainMatch creates a new domain match result
func NewDomainMatch(subdomain, rootDomain string, cert *CertInfo) *DomainMatch {
	return &DomainMatch{
		Subdomain:   subdomain,
		RootDomain:  rootDomain,
		Certificate: cert,
		MatchedAt:   time.Now(),
		MatchType:   "prefix",
		Confidence:  0.9,
		HitCount:    1,
	}
}

// DomainTrieNode is a trie node (used for efficient domain matching)
type DomainTrieNode struct {
	Children  map[rune]*DomainTrieNode `json:"-"`         // Child nodes (not serialized)
	IsEnd     bool                     `json:"is_end"`    // Whether it is an end node
	Domain    string                   `json:"domain"`    // Full domain
	AddedAt   time.Time                `json:"added_at"`  // Add time
	HitCount  int64                    `json:"hit_count"` // Hit count
	DomainObj *Domain                  `json:"-"`         // Domain object (not serialized)
}

// NewDomainTrieNode creates a new trie node
func NewDomainTrieNode() *DomainTrieNode {
	return &DomainTrieNode{
		Children: make(map[rune]*DomainTrieNode),
		IsEnd:    false,
	}
}

// DomainStats holds domain statistics
type DomainStats struct {
	TotalDomains      int64   `json:"total_domains"`      // Total domains
	UniqueSubdomains  int64   `json:"unique_subdomains"`  // Unique subdomains
	NewDiscoveries    int64   `json:"new_discoveries"`    // New discoveries
	RootDomainCount   int64   `json:"root_domain_count"`  // Root domain count
	WildcardDomains   int64   `json:"wildcard_domains"`   // Wildcard domain count
	TotalMatches      int64   `json:"total_matches"`      // Total matches
	AverageConfidence float64 `json:"average_confidence"` // Average match confidence
	CacheHitRate      float64 `json:"cache_hit_rate"`     // Cache hit rate
}

// DomainMatcherStats holds domain matcher statistics
type DomainMatcherStats struct {
	DomainStats
	BloomFilterStats BloomFilterStats `json:"bloom_filter_stats"` // Bloom filter statistics
	TrieStats        TrieStats        `json:"trie_stats"`         // Trie statistics
	CacheStats       CacheStats       `json:"cache_stats"`        // Cache statistics
}

// BloomFilterStats holds bloom filter statistics
type BloomFilterStats struct {
	Size              uint    `json:"size"`                // Filter size
	HashCount         uint    `json:"hash_count"`          // Number of hash functions
	ItemCount         uint64  `json:"item_count"`          // Item count
	FalsePositiveRate float64 `json:"false_positive_rate"` // False positive rate
}

// TrieStats holds trie statistics
type TrieStats struct {
	NodeCount     int64 `json:"node_count"`      // Node count
	MaxDepth      int   `json:"max_depth"`       // Maximum depth
	DomainCount   int64 `json:"domain_count"`    // Domain count
	TotalHitCount int64 `json:"total_hit_count"` // Total hit count
}

// CacheStats holds cache statistics
type CacheStats struct {
	Size      int     `json:"size"`       // Cache size
	ItemCount int     `json:"item_count"` // Item count
	HitCount  int64   `json:"hit_count"`  // Hit count
	MissCount int64   `json:"miss_count"` // Miss count
	HitRate   float64 `json:"hit_rate"`   // Hit rate
}

// DomainMatcherInterface is the domain matcher interface
type DomainMatcherInterface interface {
	// Domain management
	AddDomain(domain string) error
	RemoveDomain(domain string) error
	HasDomain(domain string) bool
	GetAllDomains() []string

	// Matching
	MatchSubdomains(domains []string) []*DomainMatch
	IsSubdomainOf(subdomain, rootDomain string) bool

	// Statistics
	GetStats() *DomainMatcherStats
	Clear() error
}

// NormalizeDomain normalizes a domain
func NormalizeDomain(domain string, caseSensitive bool) string {
	if domain == "" {
		return ""
	}

	domain = strings.TrimSpace(domain)

	if !caseSensitive {
		domain = strings.ToLower(domain)
	}

	// Remove the trailing dot
	domain = strings.TrimSuffix(domain, ".")

	return domain
}

// ValidateDomain validates the domain format
func ValidateDomain(domain string, maxLength int) error {
	if domain == "" {
		return fmt.Errorf("domain must not be empty")
	}

	if len(domain) > maxLength {
		return fmt.Errorf("domain length exceeds the maximum limit of %d", maxLength)
	}

	// Validate the characters
	for _, r := range domain {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '.' && r != '*' {
			return fmt.Errorf("domain contains an invalid character: %c", r)
		}
	}

	return nil
}
