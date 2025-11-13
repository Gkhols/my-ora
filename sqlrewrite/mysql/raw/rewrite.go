package raw

import (
	"regexp"
	"strconv"
	"strings"
)

// RewriteSQL converts MySQL-like SQL syntax to Oracle-compatible SQL
func RewriteSQL(query string) string {
	q := query

	// Basic replacements
	q = strings.ReplaceAll(q, "`", "\"")
	q = strings.ReplaceAll(q, "AUTO_INCREMENT", "GENERATED ALWAYS AS IDENTITY")
	q = strings.ReplaceAll(q, "BOOLEAN", "NUMBER(1)")
	q = strings.ReplaceAll(q, "TRUE", "1")
	q = strings.ReplaceAll(q, "FALSE", "0")

	// Replace functions
	q = rewriteFunctions(q)

	// Replace NOW() after functions
	q = strings.ReplaceAll(q, "NOW()", "SYSDATE")

	// Remove MySQL table options
	q = removeEngineCharset(q)

	// Handle LIMIT/OFFSET
	q = rewriteLimitOffset(q)

	// Replace ? placeholders → :1, :2 ...
	q = convertPlaceholders(q)

	return q
}

func rewriteFunctions(q string) string {
	replacements := map[*regexp.Regexp]string{
		regexp.MustCompile(`(?i)\bIFNULL\s*\(`):         "NVL(",
		regexp.MustCompile(`(?i)\bISNULL\s*\(`):         "NVL(",
		regexp.MustCompile(`(?i)\bSUBSTRING\s*\(`):      "SUBSTR(",
		regexp.MustCompile(`(?i)\bCHAR_LENGTH\s*\(`):    "LENGTH(",
		regexp.MustCompile(`(?i)\bLENGTH\s*\(`):         "LENGTH(",
		regexp.MustCompile(`(?i)\bREPLACE\s*\(`):        "REPLACE(",
		regexp.MustCompile(`(?i)\bLOCATE\s*\(`):         "INSTR(",
		regexp.MustCompile(`(?i)\bCURDATE\s*\(\)`):      "TRUNC(SYSDATE)",
		regexp.MustCompile(`(?i)\bCURRENT_DATE\s*\(\)`): "TRUNC(SYSDATE)",
		regexp.MustCompile(`(?i)\bCURTIME\s*\(\)`):      "TO_CHAR(SYSDATE,'HH24:MI:SS')",
		regexp.MustCompile(`(?i)\bDATE_FORMAT\s*\(`):    "TO_CHAR(",
		regexp.MustCompile(`(?i)\bDATEDIFF\s*\(`):       "(",
		regexp.MustCompile(`(?i)\bRAND\s*\(\)`):         "DBMS_RANDOM.VALUE",
		regexp.MustCompile(`(?i)\bFLOOR\s*\(`):          "FLOOR(",
		regexp.MustCompile(`(?i)\bCEIL\s*\(`):           "CEIL(",
		regexp.MustCompile(`(?i)\bROUND\s*\(`):          "ROUND(",
		regexp.MustCompile(`(?i)\bMOD\s*\(`):            "MOD(",
	}

	for re, repl := range replacements {
		q = re.ReplaceAllString(q, repl)
	}

	// CONCAT(a,b,c) → a || b || c
	q = rewriteConcat(q)

	return q
}

func rewriteConcat(q string) string {
	re := regexp.MustCompile(`(?i)CONCAT\s*\(([^)]+)\)`)
	return re.ReplaceAllStringFunc(q, func(match string) string {
		inner := re.FindStringSubmatch(match)[1]
		parts := strings.Split(inner, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return strings.Join(parts, " || ")
	})
}

func removeEngineCharset(q string) string {
	re := regexp.MustCompile(`(?i)ENGINE\s*=\s*\w+`)
	q = re.ReplaceAllString(q, "")
	re = regexp.MustCompile(`(?i)DEFAULT\s*CHARSET\s*=\s*\w+`)
	q = re.ReplaceAllString(q, "")
	return q
}

func rewriteLimitOffset(query string) string {
	// Case 1: LIMIT ? OFFSET ?
	re := regexp.MustCompile(`(?i)\s+LIMIT\s+([:\?\w\d]+)\s+OFFSET\s+([:\?\w\d]+)`)
	if re.MatchString(query) {
		return re.ReplaceAllString(query, " OFFSET $2 ROWS FETCH NEXT $1 ROWS ONLY")
	}

	// Case 2: OFFSET ? LIMIT ?
	re = regexp.MustCompile(`(?i)\s+OFFSET\s+([:\?\w\d]+)\s+LIMIT\s+([:\?\w\d]+)`)
	if re.MatchString(query) {
		return re.ReplaceAllString(query, " OFFSET $1 ROWS FETCH NEXT $2 ROWS ONLY")
	}

	// Case 3: LIMIT ? only
	re = regexp.MustCompile(`(?i)\s+LIMIT\s+([:\?\w\d]+)`)
	if re.MatchString(query) {
		return re.ReplaceAllString(query, " FETCH NEXT $1 ROWS ONLY")
	}

	return query
}

func convertPlaceholders(query string) string {
	var sb strings.Builder
	count := 1
	for _, ch := range query {
		if ch == '?' {
			sb.WriteString(":" + strconv.Itoa(count))
			count++
		} else {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}
