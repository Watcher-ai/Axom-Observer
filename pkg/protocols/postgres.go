package protocols

import (
	"axom-observer/pkg/models"
	"net"
	"regexp"
	"strings"
	"time"
)

// ProcessPostgres parses PostgreSQL queries from raw packets.
// For production, use a proper PostgreSQL protocol parser.
func ProcessPostgres(packet []byte, src, dst net.Addr) (*models.Signal, error) {
	// TODO: Use a real PostgreSQL wire protocol parser for robust extraction.
	op, table := extractPostgresSQLOperation(packet)
	return &models.Signal{
		Timestamp:   time.Now(),
		Protocol:    "postgres",
		Source:      AddrToEndpoint(src),
		Destination: AddrToEndpoint(dst),
		DBOperation: op,
		DBTable:     table,
		RawRequest:  packet,
	}, nil
}

// extractPostgresSQLOperation tries to extract the SQL operation and table name from the packet.
// WARNING: This is a naive implementation based on simple string splitting.
// For production, use a real SQL or wire protocol parser to handle all edge cases and SQL dialects.
// TODO: Integrate a proper SQL parser or PostgreSQL wire protocol parser for robust extraction.
var (
	pgSelectRe = regexp.MustCompile(`(?i)^SELECT\s+.*\s+FROM\s+([^\s;]+)`)
	pgInsertRe = regexp.MustCompile(`(?i)^INSERT\s+INTO\s+([^\s(]+)`)
	pgUpdateRe = regexp.MustCompile(`(?i)^UPDATE\s+([^\s]+)`)
	pgDeleteRe = regexp.MustCompile(`(?i)^DELETE\s+FROM\s+([^\s;]+)`)
)

func extractPostgresSQLOperation(packet []byte) (string, string) {
	sql := strings.TrimSpace(string(packet))
	qUpper := strings.ToUpper(sql)
	switch {
	case strings.HasPrefix(qUpper, "SELECT"):
		if m := pgSelectRe.FindStringSubmatch(sql); m != nil {
			return "SELECT", m[1]
		}
		return "SELECT", ""
	case strings.HasPrefix(qUpper, "INSERT"):
		if m := pgInsertRe.FindStringSubmatch(sql); m != nil {
			return "INSERT", m[1]
		}
		return "INSERT", ""
	case strings.HasPrefix(qUpper, "UPDATE"):
		if m := pgUpdateRe.FindStringSubmatch(sql); m != nil {
			return "UPDATE", m[1]
		}
		return "UPDATE", ""
	case strings.HasPrefix(qUpper, "DELETE"):
		if m := pgDeleteRe.FindStringSubmatch(sql); m != nil {
			return "DELETE", m[1]
		}
		return "DELETE", ""
	default:
		return "", ""
	}
}
