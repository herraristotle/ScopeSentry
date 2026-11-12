// helper-------------------------------------
// @file      : task.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/6/2 21:08
// -------------------------------------------

package helper

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// GenerateIPRange generates an IP range
func GenerateIPRange(startIP, endIP string) ([]string, error) {
	start := net.ParseIP(startIP)
	end := net.ParseIP(endIP)
	if start == nil || end == nil {
		return nil, fmt.Errorf("invalid IP range")
	}

	var ipList []string
	for ip := start; !ip.Equal(end); ip = NextIP(ip) {
		ipList = append(ipList, ip.String())
	}
	ipList = append(ipList, end.String()) // Include the end IP
	return ipList, nil
}

// NextIP generates the next IP (IPv4 only)
func NextIP(ip net.IP) net.IP {
	ip = ip.To4()
	result := make(net.IP, len(ip))
	copy(result, ip)
	for i := len(result) - 1; i >= 0; i-- {
		result[i]++
		if result[i] != 0 {
			break
		}
	}
	return result
}

// GenerateTarget parses target, which can be a single IP, IP range, CIDR or URL
func GenerateTarget(target string) []string {
	target = strings.TrimSpace(target)
	if strings.Contains(target, "://") {
		return []string{target}
	}
	if strings.Contains(target, "-") {
		parts := strings.SplitN(target, "-", 2)
		if len(parts) != 2 {
			return []string{target}
		}
		ipRange, err := GenerateIPRange(parts[0], parts[1])
		if err != nil {
			return []string{target}
		}
		return ipRange
	} else if strings.Contains(target, "/") {
		_, network, err := net.ParseCIDR(target)
		if err != nil {
			return []string{target}
		}
		var result []string
		for ip := network.IP.Mask(network.Mask); network.Contains(ip); ip = NextIP(ip) {
			result = append(result, ip.String())
		}
		// Remove the network address and broadcast address
		if len(result) >= 2 {
			result = result[1 : len(result)-1]
		}
		return result
	}
	return []string{target}
}

// GenerateIgnore parses the ignore list, returning exact ignore IPs and the regex ignore list
func GenerateIgnore(ignore string) ([]string, []string) {
	if ignore == "" {
		return []string{}, []string{}
	}
	var ignoreList []string
	var regexList []string

	lines := strings.Split(ignore, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(line, "http://", ""), "https://", ""))
		if line == "" {
			continue
		}
		if !strings.Contains(line, "*") {
			targets := GenerateTarget(line)
			ignoreList = append(ignoreList, targets...)
		} else {
			escaped := regexp.QuoteMeta(line)
			regexList = append(regexList, strings.ReplaceAll(escaped, `\*`, ".*"))
		}
	}
	return ignoreList, regexList
}

// GetTargetList generates a deduplicated, order-preserving target list from the raw target list and the ignore list
func GetTargetList(rawTarget, ignore string) ([]string, error) {
	ignoreList, regexList := GenerateIgnore(ignore)
	ignoreSet := make(map[string]struct{})
	for _, ip := range ignoreList {
		ignoreSet[ip] = struct{}{}
	}

	seen := make(map[string]struct{})
	var finalList []string

	targetLines := strings.Split(rawTarget, "\n")
	for _, line := range targetLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		targets := GenerateTarget(line)
		for _, t := range targets {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			if _, found := ignoreSet[t]; found {
				continue
			}
			shouldIgnore := false
			for _, pattern := range regexList {
				matched, _ := regexp.MatchString(pattern, t)
				if matched {
					shouldIgnore = true
					break
				}
			}
			if !shouldIgnore {
				if _, exists := seen[t]; !exists {
					seen[t] = struct{}{}
					finalList = append(finalList, t)
				}
			}
		}
	}
	return finalList, nil
}
