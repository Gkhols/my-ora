package raw

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"

	"github.com/godror/godror"
)

// -------------------- DRIVER --------------------

type rewritedDriver struct{}

func (d *rewritedDriver) Open(dsn string) (driver.Conn, error) {
	params, err := godror.ParseConnString(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN: %w", err)
	}

	connector := godror.NewConnector(params)
	conn, err := connector.Connect(context.Background())
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return &rewritedConn{inner: conn}, nil
}

func Register() {
	sql.Register("my-ora", &rewritedDriver{})
}

// -------------------- CONNECTION --------------------

type rewritedConn struct {
	inner driver.Conn
}

func (c *rewritedConn) Prepare(query string) (driver.Stmt, error) {
	newQuery := RewriteSQL(query)
	needsSwap := needsLimitOffsetSwap(query) // check original MySQL query
	stmt, err := c.inner.Prepare(newQuery)
	if err != nil {
		return nil, err
	}

	return &rewritedStmt{
		inner:     stmt,
		query:     newQuery,
		needsSwap: needsSwap,
		origQuery: query,
	}, nil
}

func (c *rewritedConn) Close() error              { return c.inner.Close() }
func (c *rewritedConn) Begin() (driver.Tx, error) { return c.inner.Begin() }

// -------------------- CONTEXT EXEC/QUERY --------------------

func (c *rewritedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if execer, ok := c.inner.(driver.ExecerContext); ok {
		newQuery := RewriteSQL(query)
		args = swapLimitOffsetArgs(query, args)
		return execer.ExecContext(ctx, newQuery, args)
	}
	return nil, driver.ErrSkip
}

func (c *rewritedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if queryer, ok := c.inner.(driver.QueryerContext); ok {
		newQuery := RewriteSQL(query)
		args = swapLimitOffsetArgs(query, args)
		return queryer.QueryContext(ctx, newQuery, args)
	}
	return nil, driver.ErrSkip
}

// -------------------- STATEMENT WRAPPER --------------------

type rewritedStmt struct {
	inner     driver.Stmt
	query     string
	needsSwap bool
	origQuery string
}

func (s *rewritedStmt) Close() error { return s.inner.Close() }
func (s *rewritedStmt) NumInput() int {
	return s.inner.NumInput()
}

// Classic Exec
func (s *rewritedStmt) Exec(args []driver.Value) (driver.Result, error) {
	if s.needsSwap && len(args) >= 2 {
		re := regexp.MustCompile(`(?i)LIMIT\s+(\?|\:\d+)\s+OFFSET\s+(\?|\:\d+)`)
		m := re.FindStringSubmatchIndex(s.origQuery)
		if len(m) >= 4 {
			// find the position of LIMIT and OFFSET placeholders
			limitPos := m[2]
			offsetPos := m[4]
			if limitPos < offsetPos && len(args) >= 2 {
				// only swap the last two args (limit and offset)
				last := len(args) - 1
				args[last-1], args[last] = args[last], args[last-1]
			}
		}
	}
	return s.inner.Exec(args)
}

// Classic Query
func (s *rewritedStmt) Query(args []driver.Value) (driver.Rows, error) {
	if s.needsSwap && len(args) >= 2 {
		re := regexp.MustCompile(`(?i)LIMIT\s+(\?|\:\d+)\s+OFFSET\s+(\?|\:\d+)`)
		m := re.FindStringSubmatchIndex(s.origQuery)
		if len(m) >= 4 {
			limitPos := m[2]
			offsetPos := m[4]
			if limitPos < offsetPos && len(args) >= 2 {
				last := len(args) - 1
				args[last-1], args[last] = args[last], args[last-1]
			}
		}
	}
	return s.inner.Query(args)
}

// -------------------- HELPER --------------------

// Detect if query uses LIMIT ? OFFSET ? or OFFSET ? LIMIT ?
func needsLimitOffsetSwap(originalQuery string) bool {
	re := regexp.MustCompile(`(?i)LIMIT\s+(\?|\:\d+)\s+OFFSET\s+(\?|\:\d+)`)
	return re.MatchString(originalQuery)
}

// Swap LIMIT/OFFSET arguments if query is rewritten to Oracle style
func swapLimitOffsetArgs(originalQuery string, args []driver.NamedValue) []driver.NamedValue {
	if len(args) < 2 {
		return args
	}

	// Find placeholder order
	upper := strings.ToUpper(originalQuery)
	limitIdx := strings.Index(upper, "LIMIT")
	offsetIdx := strings.Index(upper, "OFFSET")

	// Only proceed if LIMIT appears before OFFSET
	if limitIdx == -1 || offsetIdx == -1 || limitIdx > offsetIdx {
		return args
	}

	// Count parameter order in query text
	paramCount := 0
	var limitArgIdx, offsetArgIdx int = -1, -1

	for i := 0; i < len(originalQuery); i++ {
		if originalQuery[i] == '?' || originalQuery[i] == ':' {
			paramCount++
			if i > limitIdx && limitArgIdx == -1 {
				limitArgIdx = paramCount - 1
			} else if i > offsetIdx && offsetArgIdx == -1 {
				offsetArgIdx = paramCount - 1
			}
		}
	}

	if limitArgIdx >= 0 && offsetArgIdx >= 0 &&
		limitArgIdx < len(args) && offsetArgIdx < len(args) {
		args[limitArgIdx], args[offsetArgIdx] = args[offsetArgIdx], args[limitArgIdx]
	}

	return args
}
