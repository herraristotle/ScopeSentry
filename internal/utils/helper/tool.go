// helper-------------------------------------
// @file      : tool.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/8 21:23
// -------------------------------------------

package helper

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/cespare/xxhash/v2"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/url"
	"strings"
)

func Sha256Hex(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

func HashXX64String(input string) string {
	hash := xxhash.Sum64String(input)
	return fmt.Sprintf("%x", hash)
}

func CalculateMD5FromContent(content string) string {
	hash := md5.New()
	hash.Write([]byte(content))
	return hex.EncodeToString(hash.Sum(nil))
}

func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func RemoveArrayDuplicates(input []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, str := range input {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}

func GetRootDomain(input string) (string, error) {
	u, err := SafeParseURL(input)
	if err != nil {
		return input, fmt.Errorf("URL parsing failed: %w", err)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return input, fmt.Errorf("unable to get Hostname")
	}

	// If it is an IP, return it directly
	if ip := net.ParseIP(hostname); ip != nil {
		return ip.String(), nil
	}

	// Extract the valid root domain
	rootDomain, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil {
		return input, fmt.Errorf("root domain parsing error: %w", err)
	}

	return rootDomain, nil
}

// SafeParseURL parses a URL safely, handling invalid %, IDN domains, spaces and similar issues
func SafeParseURL(input string) (*url.URL, error) {
	if input == "" {
		return nil, fmt.Errorf("input is empty")
	}

	input = strings.TrimSpace(input)
	input = sanitizePercent(input)
	input = strings.ReplaceAll(input, " ", "%20")

	// Add the protocol prefix if missing
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		input = "https://" + input
	}

	u, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("URL parsing failed: %w", err)
	}

	// IDNA-process IDN domains
	asciiHost, err := idna.ToASCII(u.Hostname())
	if err != nil {
		return nil, fmt.Errorf("domain IDNA conversion failed: %w", err)
	}

	// Append the port (if any)
	u.Host = asciiHost + portSuffix(u.Host)

	return u, nil
}

// sanitizePercent fixes invalid % encoding
func sanitizePercent(input string) string {
	var builder strings.Builder
	for i := 0; i < len(input); i++ {
		if input[i] == '%' {
			if i+2 >= len(input) || !isHexDigit(input[i+1]) || !isHexDigit(input[i+2]) {
				builder.WriteString("%25")
			} else {
				builder.WriteByte('%')
			}
		} else {
			builder.WriteByte(input[i])
		}
	}
	return builder.String()
}

// Determines whether a byte is a valid hex digit
func isHexDigit(b byte) bool {
	return ('0' <= b && b <= '9') || ('a' <= b && b <= 'f') || ('A' <= b && b <= 'F')
}

// Extracts the port from the original host (if any)
func portSuffix(host string) string {
	if colon := strings.LastIndex(host, ":"); colon != -1 && colon < len(host)-1 {
		port := host[colon+1:]
		if _, err := fmt.Sscanf(port, "%d", new(int)); err == nil {
			return ":" + port
		}
	}
	return ""
}
