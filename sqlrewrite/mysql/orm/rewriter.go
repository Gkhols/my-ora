package orm

import (
	"regexp"
	"strings"

	"github.com/Gkhols/my-ora/sqlrewrite/mysql/raw"
	"gorm.io/gorm"
)

func RawWithRewriter(db *gorm.DB, sqlOriginal string, args ...interface{}) *gorm.DB {
	upperSQL := strings.ToUpper(sqlOriginal)
	// Detect MySQL LIMIT ? OFFSET ? pattern
	needsSwap := needsLimitOffsetSwap(upperSQL)

	if needsSwap && len(args) >= 2 {
		args = swapLimitOffsetArgs(sqlOriginal, args)
	}
	sqlOriginal = raw.RewriteSQL(sqlOriginal)
	return db.Raw(sqlOriginal, args...)
}

func needsLimitOffsetSwap(originalQuery string) bool {
	re := regexp.MustCompile(`(?i)LIMIT\s+(\?|\:\d+)\s+OFFSET\s+(\?|\:\d+)`)
	return re.MatchString(originalQuery)
}

func swapLimitOffsetArgs(originalSQL string, args []interface{}) []interface{} {
	if len(args) < 2 {
		return args
	}

	upper := strings.ToUpper(originalSQL)
	limitIdx := strings.Index(upper, "LIMIT")
	offsetIdx := strings.Index(upper, "OFFSET")
	if limitIdx == -1 || offsetIdx == -1 || limitIdx > offsetIdx {
		return args
	}

	paramCount := 0
	limitArgIdx, offsetArgIdx := -1, -1

	for i := 0; i < len(originalSQL); i++ {
		if originalSQL[i] == '?' || originalSQL[i] == ':' {
			if i > limitIdx && limitArgIdx == -1 {
				limitArgIdx = paramCount
			} else if i > offsetIdx && offsetArgIdx == -1 {
				offsetArgIdx = paramCount
			}
			paramCount++
		}
	}

	if limitArgIdx >= 0 && offsetArgIdx >= 0 && limitArgIdx < len(args) && offsetArgIdx < len(args) {
		args[limitArgIdx], args[offsetArgIdx] = args[offsetArgIdx], args[limitArgIdx]
	}
	return args
}
