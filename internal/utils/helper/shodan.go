// helper-------------------------------------
// @file      : shodan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/12/21 19:47
// -------------------------------------------

package helper

import (
	"fmt"
	"regexp"
	"strings"
)

// ConvertShodanToCustomQuery converts a shodan query into the project's custom search expression
func ConvertShodanToCustomQuery(shodanQueryValue interface{}) string {
	var queries []string

	// Process array types
	if arr, ok := shodanQueryValue.([]interface{}); ok {
		for _, item := range arr {
			if str, ok := item.(string); ok {
				converted := parseShodanQuery(str)
				if converted != "" {
					queries = append(queries, converted)
				}
			}
		}
		// Join multiple conditions with &&
		if len(queries) > 0 {
			return strings.Join(queries, " && ")
		}
		return ""
	}

	// Process string types
	if str, ok := shodanQueryValue.(string); ok {
		return parseShodanQuery(str)
	}

	return ""
}

// parseShodanQuery parses a shodan query, supporting brackets and logical operators
func parseShodanQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	// Process brackets first, parsing recursively
	return parseShodanQueryWithBrackets(query)
}

// parseShodanQueryWithBrackets parses a bracketed shodan query
func parseShodanQueryWithBrackets(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	// Find and replace all bracketed content
	for {
		// Find the outermost opening bracket
		start := strings.Index(query, "(")
		if start == -1 {
			// No more brackets, parse the logical expression
			break
		}

		// Find the matching closing bracket
		depth := 0
		end := -1
		for i := start; i < len(query); i++ {
			if query[i] == '(' {
				depth++
			} else if query[i] == ')' {
				depth--
				if depth == 0 {
					end = i
					break
				}
			}
		}

		if end == -1 {
			// Unmatched bracket, break out of the loop and treat it as a normal query
			break
		}

		// Recursively parse the bracketed content
		innerQuery := query[start+1 : end]
		parsedInner := parseShodanQueryWithBrackets(innerQuery)

		// Replace the bracketed content with the parsed result
		query = query[:start] + "(" + parsedInner + ")" + query[end+1:]
	}

	// All brackets processed, parse the logical expression
	return parseShodanLogicExpression(query)
}

// parseShodanLogicExpression parses a logical expression (handling AND, OR, NOT)
func parseShodanLogicExpression(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	// Split by || (lowest precedence, symbolic form)
	if hasOperator(query, "||") {
		parts := splitByShodanOperatorSymbol(query, "||")
		var results []string
		for _, part := range parts {
			parsed := parseShodanLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " || ")
		}
		return ""
	}

	// Split by OR (lowest precedence, handling case; text form)
	if strings.Contains(strings.ToUpper(query), " OR ") {
		parts := splitByShodanOperator(query, " OR ")
		var results []string
		for _, part := range parts {
			parsed := parseShodanLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " || ")
		}
		return ""
	}

	// Split by && (symbolic form)
	if hasOperator(query, "&&") {
		parts := splitByShodanOperatorSymbol(query, "&&")
		var results []string
		for _, part := range parts {
			parsed := parseShodanLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " && ")
		}
		return ""
	}

	// Split by AND (case- and space-sensitive, text form)
	if strings.Contains(strings.ToUpper(query), " AND ") {
		parts := splitByShodanOperator(query, " AND ")
		var results []string
		for _, part := range parts {
			parsed := parseShodanLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " && ")
		}
		return ""
	}

	// No logical operator, parse a single condition (may contain multiple space-separated filters; the default is AND)
	return parseShodanFilters(query)
}

// splitByShodanOperator splits by the operator, taking care of bracketed content and case
func splitByShodanOperator(query, operator string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	operatorUpper := strings.ToUpper(operator)
	operatorLen := len(operator)

	for i := 0; i < len(query); i++ {
		if query[i] == '(' {
			depth++
			current.WriteByte(query[i])
		} else if query[i] == ')' {
			depth--
			current.WriteByte(query[i])
		} else if depth == 0 {
			// Check whether the operator matches (case-insensitive)
			if i+operatorLen <= len(query) {
				substr := strings.ToUpper(query[i : i+operatorLen])
				if substr == operatorUpper {
					parts = append(parts, current.String())
					current.Reset()
					i += operatorLen - 1 // Skip the operator
					continue
				}
			}
			current.WriteByte(query[i])
		} else {
			current.WriteByte(query[i])
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// hasOperator checks whether the query contains an operator (outside quotes)
func hasOperator(query, operator string) bool {
	inQuotes := false
	quoteChar := byte(0)
	depth := 0

	for i := 0; i <= len(query)-len(operator); i++ {
		char := query[i]

		// Handle quotes
		if (char == '"' || char == '\'') && (i == 0 || query[i-1] != '\\') {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
			continue
		}

		// Skip if inside quotes
		if inQuotes {
			continue
		}

		// Handle bracket depth
		if char == '(' {
			depth++
			continue
		} else if char == ')' {
			depth--
			continue
		}

		// Check the operator (outside quotes and brackets)
		if depth == 0 && query[i:i+len(operator)] == operator {
			return true
		}
	}

	return false
}

// splitByShodanOperatorSymbol splits by a symbolic operator (such as || or &&), taking care of bracketed and quoted content
func splitByShodanOperatorSymbol(query, operator string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	inQuotes := false
	quoteChar := byte(0)
	operatorLen := len(operator)

	for i := 0; i < len(query); i++ {
		char := query[i]

		// Handle quotes
		if (char == '"' || char == '\'') && (i == 0 || query[i-1] != '\\') {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
			current.WriteByte(char)
			continue
		}

		// Append the character directly if inside quotes
		if inQuotes {
			current.WriteByte(char)
			continue
		}

		// Handle brackets
		if char == '(' {
			depth++
			current.WriteByte(char)
		} else if char == ')' {
			depth--
			current.WriteByte(char)
		} else if depth == 0 {
			// Check whether the operator matches (outside quotes and brackets)
			if i+operatorLen <= len(query) && query[i:i+operatorLen] == operator {
				parts = append(parts, current.String())
				current.Reset()
				i += operatorLen - 1 // Skip the operator
				continue
			}
			current.WriteByte(char)
		} else {
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseShodanFilters parses shodan filters (may contain multiple space-separated filters)
func parseShodanFilters(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	// Remove brackets (if the whole query is wrapped in brackets)
	query = strings.Trim(query, "()")
	query = strings.TrimSpace(query)

	// Split the filters (by space, taking care of quoted content)
	filters := splitShodanFilters(query)
	if len(filters) == 0 {
		return ""
	}

	var results []string
	for _, filter := range filters {
		filter = strings.TrimSpace(filter)
		if filter == "" {
			continue
		}
		parsed := parseShodanFilter(filter)
		if parsed != "" {
			results = append(results, parsed)
		}
	}

	if len(results) == 0 {
		return ""
	}

	// Join multiple filters with &&
	if len(results) == 1 {
		return results[0]
	}
	return strings.Join(results, " && ")
}

// splitShodanFilters splits shodan filters, taking care of quoted content
// Note: this function only handles space-separated filters; OR and AND are handled in parseShodanLogicExpression
func splitShodanFilters(query string) []string {
	var filters []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(query); i++ {
		char := query[i]

		// Handle quotes
		if (char == '"' || char == '\'') && (i == 0 || query[i-1] != '\\') {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
				current.WriteByte(char)
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
				current.WriteByte(char)
			} else {
				current.WriteByte(char)
			}
			continue
		}

		// Append the character directly if inside quotes
		if inQuotes {
			current.WriteByte(char)
			continue
		}

		// Space-separated (but outside quotes)
		if char == ' ' {
			if current.Len() > 0 {
				filters = append(filters, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(char)
	}

	if current.Len() > 0 {
		filters = append(filters, current.String())
	}

	return filters
}

// parseShodanFilter parses a single shodan filter
func parseShodanFilter(filter string) string {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return ""
	}

	// Define the shodan field mappings (shodan field -> custom field)
	fieldMap := map[string]string{
		"http.title":        "title",
		"http.html":         "body",
		"http.component":    "app",
		"http.server":       "header",
		"http.favicon.hash": "icon",
		"http.status":       "statuscode",
		"title":             "title",
		"html":              "body",
		"component":         "app",
		"server":            "header",
		"favicon.hash":      "icon",
		"status":            "statuscode",
	}

	// Check whether it is a NOT query (starts with - or NOT, case-insensitive)
	isNot := false
	if strings.HasPrefix(filter, "-") {
		isNot = true
		filter = strings.TrimPrefix(filter, "-")
		filter = strings.TrimSpace(filter)
	} else {
		// Check whether it starts with NOT (case-insensitive)
		filterUpper := strings.ToUpper(filter)
		if strings.HasPrefix(filterUpper, "NOT ") {
			isNot = true
			// Find the position of NOT in the original string (case-insensitive)
			notIndex := strings.Index(filterUpper, "NOT ")
			if notIndex == 0 {
				filter = filter[4:] // Remove "NOT " (4 characters)
				filter = strings.TrimSpace(filter)
			}
		}
	}

	// Match the shodan filter format: filter:value or "filter:value"
	// Support quoted values
	pattern := regexp.MustCompile(`^([\w.]+)\s*:\s*(.+)$`)
	matches := pattern.FindStringSubmatch(filter)
	if len(matches) == 3 {
		field := strings.ToLower(matches[1])
		value := strings.TrimSpace(matches[2])

		// If the field is country, skip it directly (the country field is not supported)
		if field == "country" {
			return ""
		}

		// Remove quotes at both ends of the value (if present)
		value = strings.Trim(value, `"'`)

		// Check whether the field is in the mapping
		customField, exists := fieldMap[field]
		if !exists {
			// If the field is not in the mapping, default to body
			customField = "body"
		}

		// Choose the operator based on whether it is a NOT query
		operator := "="
		if isNot {
			operator = "!="
		}

		return fmt.Sprintf(`%s%s"%s"`, customField, operator, value)
	}

	// If no filter format matched, check whether it is a plain text query
	// Plain text queries default to a body query
	if !strings.Contains(filter, ":") {
		value := strings.Trim(filter, `"'`)
		operator := "="
		if isNot {
			operator = "!="
		}
		return fmt.Sprintf(`body%s"%s"`, operator, value)
	}

	return ""
}
