// helper-------------------------------------
// @file      : fofa.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/12/21 19:39
// -------------------------------------------

package helper

import (
	"fmt"
	"regexp"
	"strings"
)

// convertFofaToCustomQuery converts a fofa-query into the project's custom search expression
func ConvertFofaToCustomQuery(fofaQueryValue interface{}) string {
	var queries []string

	// Process array types
	if arr, ok := fofaQueryValue.([]interface{}); ok {
		for _, item := range arr {
			if str, ok := item.(string); ok {
				converted := parseFofaQuery(str)
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
	if str, ok := fofaQueryValue.(string); ok {
		return parseFofaQuery(str)
	}

	return ""
}

// parseFofaQuery parses a fofa query, supporting brackets and logical operators
func parseFofaQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	// Process brackets first, parsing recursively
	return parseFofaQueryWithBrackets(query)
}

// parseFofaQueryWithBrackets parses a bracketed fofa query
func parseFofaQueryWithBrackets(query string) string {
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
		parsedInner := parseFofaQueryWithBrackets(innerQuery)

		// Replace the bracketed content with the parsed result
		query = query[:start] + "(" + parsedInner + ")" + query[end+1:]
	}

	// All brackets processed, parse the logical expression
	return parseFofaLogicExpression(query)
}

// parseFofaLogicExpression parses a logical expression (handling && and ||)
func parseFofaLogicExpression(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	// Split by || (lowest precedence)
	if strings.Contains(query, " || ") {
		parts := splitByOperator(query, " || ")
		var results []string
		for _, part := range parts {
			parsed := parseFofaLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " || ")
		}
		return ""
	}

	// Split by &&
	if strings.Contains(query, " && ") {
		parts := splitByOperator(query, " && ")
		var results []string
		for _, part := range parts {
			parsed := parseFofaLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " && ")
		}
		return ""
	}

	// No logical operator, parse a single condition
	return parseFofaCondition(query)
}

// splitByOperator splits by the operator, taking care of bracketed content
func splitByOperator(query, operator string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for i := 0; i < len(query); i++ {
		if query[i] == '(' {
			depth++
			current.WriteByte(query[i])
		} else if query[i] == ')' {
			depth--
			current.WriteByte(query[i])
		} else if depth == 0 && i+len(operator) <= len(query) && query[i:i+len(operator)] == operator {
			parts = append(parts, current.String())
			current.Reset()
			i += len(operator) - 1 // Skip the operator
		} else {
			current.WriteByte(query[i])
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseFofaCondition parses a single fofa condition
func parseFofaCondition(condition string) string {
	condition = strings.TrimSpace(condition)
	if condition == "" {
		return ""
	}

	// Remove brackets (if the whole condition is wrapped in brackets)
	condition = strings.Trim(condition, "()")
	condition = strings.TrimSpace(condition)

	// Define the fofa field mappings (fofa field -> custom field)
	fieldMap := map[string]string{
		"app":           "app",
		"title":         "title",
		"body":          "body",
		"header":        "header",
		"icon_hash":     "icon",
		"protocol":      "service",
		"banner":        "banner",
		"banner_hash":   "banner",
		"base_protocol": "service",
		"status_code":   "statuscode",
		"js_name":       "body",
	}

	// Match the fofa query format, supporting =, ==, !=, *=
	// Sort by precedence: ==, *=, =, !=
	patterns := []struct {
		regex   *regexp.Regexp
		opIndex int // 0: ==, 1: *=, 2: =, 3: !=
	}{
		// Match field=="value" or field=='value' (exact match, quoted)
		{regexp.MustCompile(`^(\w+)\s*==\s*["']([^"']*)["']$`), 0},
		// Match field*="value" or field*='value' (fuzzy match, quoted)
		{regexp.MustCompile(`^(\w+)\s*\*=\s*["']([^"']*)["']$`), 1},
		// Match field="value" or field='value' (match, quoted)
		{regexp.MustCompile(`^(\w+)\s*=\s*["']([^"']*)["']$`), 2},
		// Match field!="value" or field!='value' (no match, quoted)
		{regexp.MustCompile(`^(\w+)\s*!=\s*["']([^"']*)["']$`), 3},
		// Match field==value (exact match, unquoted)
		{regexp.MustCompile(`^(\w+)\s*==\s*(.+)$`), 0},
		// Match field*=value (fuzzy match, unquoted)
		{regexp.MustCompile(`^(\w+)\s*\*=\s*(.+)$`), 1},
		// Match field=value (match, unquoted)
		{regexp.MustCompile(`^(\w+)\s*=\s*(.+)$`), 2},
		// Match field!=value (no match, unquoted)
		{regexp.MustCompile(`^(\w+)\s*!=\s*(.+)$`), 3},
	}

	operators := []string{"==", "=", "=", "!="} // *= becomes =

	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(condition)
		if len(matches) == 3 {
			field := strings.ToLower(matches[1])
			value := strings.TrimSpace(matches[2])
			// Remove quotes at both ends of the value (if present)
			value = strings.Trim(value, `"'`)

			// Check whether the field is in the mapping
			customField, exists := fieldMap[field]
			if !exists {
				// If the field is not in the mapping, default to body
				customField = "body"
			}

			operator := operators[pattern.opIndex]
			return fmt.Sprintf(`%s%s"%s"`, customField, operator, value)
		}
	}

	// If no pattern matched, check whether a known field name is present
	hasKnownField := false
	lowerCondition := strings.ToLower(condition)
	for field := range fieldMap {
		if strings.HasPrefix(lowerCondition, field+"=") ||
			strings.HasPrefix(lowerCondition, field+"==") ||
			strings.HasPrefix(lowerCondition, field+"!=") ||
			strings.HasPrefix(lowerCondition, field+"*=") {
			hasKnownField = true
			break
		}
	}

	// If no known field is present, default to a body query
	if !hasKnownField && condition != "" {
		// Remove possible quotes
		value := strings.Trim(condition, `"'`)
		// If the value starts with =, remove it
		value = strings.TrimPrefix(value, "=")
		value = strings.TrimPrefix(value, "==")
		value = strings.TrimPrefix(value, "!=")
		value = strings.TrimPrefix(value, "*=")
		value = strings.TrimSpace(value)
		if value != "" {
			return fmt.Sprintf(`body="%s"`, value)
		}
	}

	return ""
}
