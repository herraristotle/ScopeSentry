// ct-----------------------------------------
// @file      : parser.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/01/24 21:00
// -------------------------------------------
// CT log monitoring SDK - streaming certificate parser
// ============================================

package ct

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/models"
	"github.com/google/certificate-transparency-go/x509"
)

// CertificateParser is the certificate parser (streaming)
type CertificateParser struct {
	// Parsing options
	maxDomains int // Maximum number of domains
}

// NewCertificateParser creates a new certificate parser
func NewCertificateParser() *CertificateParser {
	return &CertificateParser{
		maxDomains: 100, // Default maximum of 100 domains
	}
}

// ParseCertificateStream parses a certificate in a streaming manner (core method)
// Only extracts the necessary information; does not store certificate content
func (p *CertificateParser) ParseCertificateStream(rawData []byte, index int64, timestamp time.Time, logURL string) (*models.CertInfo, error) {
	if len(rawData) == 0 {
		return nil, fmt.Errorf("certificate data is empty")
	}

	certInfo := &models.CertInfo{
		LogURL:    logURL,
		Index:     index,
		Timestamp: timestamp,
	}

	// Quickly extract domains (the most important information)
	domains, err := p.ExtractDomainsFast(rawData)
	if err != nil {
		// Domain extraction failed, but processing continues
		domains = []string{}
	}

	// Limit the number of domains
	if len(domains) > p.maxDomains {
		domains = domains[:p.maxDomains]
	}

	certInfo.AllDomains = domains

	// Try to extract the CN (Common Name)
	if len(domains) > 0 {
		certInfo.CommonName = domains[0]
	}

	return certInfo, nil
}

// ExtractDomainsFast quickly extracts domains from a certificate
// Uses the Google CT library's x509 parser for better CT log certificate support
func (p *CertificateParser) ExtractDomainsFast(certData []byte) ([]string, error) {
	domains := make([]string, 0, 10)

	// Parse the X.509 certificate (using the Google CT library's x509 parser)
	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		// If parsing fails, fall back to pattern matching
		return p.extractDomainsFromPatterns(certData), nil
	}

	// Fetch domains from DNSNames (Subject Alternative Name extension)
	// Ensure DNSNames is not nil (a nil value serializes to null in JSON)
	if cert.DNSNames != nil {
		domains = append(domains, cert.DNSNames...)
	}

	// Handle the Common Name (if not a CA certificate and the CN is not empty)
	if cert.Subject.CommonName != "" && !cert.IsCA {
		// Check whether the CN is already in the domain list
		cnAlreadyAdded := false
		for _, domain := range domains {
			if domain == cert.Subject.CommonName {
				cnAlreadyAdded = true
				break
			}
		}

		// If the CN is not in the list, add it
		if !cnAlreadyAdded {
			domains = append(domains, cert.Subject.CommonName)
		}
	}

	// Validate and clean the domains
	validDomains := make([]string, 0, len(domains))
	for _, domain := range domains {
		if p.isValidDomain(domain) {
			validDomains = append(validDomains, strings.ToLower(strings.TrimSpace(domain)))
		}
	}

	// Deduplicate
	validDomains = p.deduplicateDomains(validDomains)

	return validDomains, nil
}

// extractDomainsFromPatterns extracts domains via pattern matching (fallback)
// Used when standard X.509 parsing fails
func (p *CertificateParser) extractDomainsFromPatterns(data []byte) []string {
	domains := make([]string, 0, 10)
	dataStr := string(data)

	// Common domain separators and patterns
	separators := []string{"\x00", "\n", "\r", "\t", " ", ",", ";", "|"}

	for _, sep := range separators {
		parts := strings.Split(dataStr, sep)
		for _, part := range parts {
			if p.isValidDomain(part) {
				domains = append(domains, strings.ToLower(strings.TrimSpace(part)))
			}
		}
	}

	// Deduplicate
	domains = p.deduplicateDomains(domains)

	return domains
}

// isValidDomain validates the domain format
func (p *CertificateParser) isValidDomain(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 4 || len(s) > 253 { // Domain length limit
		return false
	}

	// Must contain a dot
	if !strings.Contains(s, ".") {
		return false
	}

	// Check character validity (characters allowed in domains)
	for _, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') &&
			!(r >= '0' && r <= '9') && r != '.' && r != '-' {
			return false
		}
	}

	// Check the TLD (top-level domain)
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return false
	}

	tld := parts[len(parts)-1]
	if len(tld) < 2 || len(tld) > 6 {
		return false
	}

	return true
}

// deduplicateDomains deduplicates the domain list
func (p *CertificateParser) deduplicateDomains(domains []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(domains))

	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain != "" && !seen[domain] {
			seen[domain] = true
			result = append(result, domain)
		}
	}

	return result
}

// CalculateCertHash computes the certificate hash (for deduplication)
func (p *CertificateParser) CalculateCertHash(certData []byte) string {
	hash := sha256.Sum256(certData)
	return fmt.Sprintf("%x", hash)
}

// ValidateCertificateData validates the certificate data format
func (p *CertificateParser) ValidateCertificateData(data []byte) error {
	if len(data) < 100 {
		return fmt.Errorf("certificate data is too small")
	}

	if len(data) > 100*1024 { // 100KB limit
		return fmt.Errorf("certificate data is too large")
	}

	// Check whether it is plausibly a DER-encoded certificate
	if data[0] != 0x30 {
		return fmt.Errorf("not a valid DER-encoded certificate")
	}

	return nil
}

// GetParserStats fetches the parser statistics
func (p *CertificateParser) GetParserStats() *ParserStats {
	return &ParserStats{
		MaxDomains: p.maxDomains,
	}
}

// ParserStats holds parser statistics
type ParserStats struct {
	MaxDomains      int   `json:"max_domains"`
	TotalParsed     int64 `json:"total_parsed"`
	ParseErrors     int64 `json:"parse_errors"`
	DomainExtracted int64 `json:"domain_extracted"`
}

// StreamProcessor is the stream processor
type StreamProcessor struct {
	parser   *CertificateParser
	matcher  *DomainMatcher
	eventBus *models.EventBus

	// Processing statistics
	stats struct {
		totalProcessed int64
		totalMatched   int64
		totalErrors    int64
	}
}

// NewStreamProcessor creates a new stream processor
func NewStreamProcessor(matcher *DomainMatcher, eventBus *models.EventBus) *StreamProcessor {
	return &StreamProcessor{
		parser:   NewCertificateParser(),
		matcher:  matcher,
		eventBus: eventBus,
	}
}

// ProcessCertificate processes a single certificate (streaming)
func (sp *StreamProcessor) ProcessCertificate(rawData []byte, index int64, timestamp time.Time, logURL string) error {
	// Parse the certificate information
	certInfo, err := sp.parser.ParseCertificateStream(rawData, index, timestamp, logURL)
	if err != nil {
		sp.stats.totalErrors++
		return fmt.Errorf("certificate parsing failed: %w", err)
	}

	sp.stats.totalProcessed++

	// Publish the certificate processing event
	certEvent := &models.CertificateProcessedEvent{
		BaseEvent: models.BaseEvent{
			Type:      models.EventTypeCertificateProcessed,
			ID:        models.GenerateEventID(),
			Timestamp: time.Now(),
		},
		Certificate:    certInfo,
		ProcessingTime: 0, // Processing time statistics can be added later
		Success:        true,
		Error:          "",
	}

	// Publish the certificate processing event asynchronously
	go func(evt *models.CertificateProcessedEvent) {
		sp.eventBus.Publish(evt)
	}(certEvent)

	// Skip if there are no domains
	if len(certInfo.AllDomains) == 0 {
		return nil
	}

	// Match the domains
	matches := sp.matcher.MatchSubdomains(certInfo.AllDomains)
	if len(matches) == 0 {
		return nil // No match, discard directly
	}

	sp.stats.totalMatched++

	// Generate and publish the event
	for _, match := range matches {
		match.Certificate = certInfo

		event := &models.SubdomainDiscoveredEvent{
			BaseEvent: models.BaseEvent{
				Type:      models.EventTypeSubdomainDiscovered,
				ID:        models.GenerateEventID(),
				Timestamp: time.Now(),
			},
			SubdomainEvent: models.SubdomainEvent{
				Subdomain:    match.Subdomain,
				RootDomain:   match.RootDomain,
				Certificate:  certInfo,
				DiscoveredAt: match.MatchedAt,
				EventID:      models.GenerateEventID(),
				IsNew:        true,
			},
		}

		// Publish the event asynchronously
		go func(evt *models.SubdomainDiscoveredEvent) {
			sp.eventBus.Publish(evt)
		}(event)
	}

	return nil
}

// GetStats fetches the stream processor statistics
func (sp *StreamProcessor) GetStats() *StreamProcessorStats {
	return &StreamProcessorStats{
		TotalProcessed: sp.stats.totalProcessed,
		TotalMatched:   sp.stats.totalMatched,
		TotalErrors:    sp.stats.totalErrors,
	}
}

// StreamProcessorStats holds stream processor statistics
type StreamProcessorStats struct {
	TotalProcessed int64 `json:"total_processed"`
	TotalMatched   int64 `json:"total_matched"`
	TotalErrors    int64 `json:"total_errors"`
}
