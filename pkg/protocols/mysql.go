package protocols

import (
	"net"
	"strings"
	"time"
   "regexp"
	"axom-observer/pkg/models"
)

// ProcessMySQL parses MySQL queries from raw packets.
// For production, use a proper MySQL protocol parser.
func ProcessMySQL(packet []byte, src, dst net.Addr) (*models.Signal, error) {
	// TODO: Use a real MySQL wire protocol parser for robust extraction.
	op, table := extractMySQLSQLOperation(packet)
	return &models.Signal{
		Timestamp:   time.Now(),
		Protocol:    "mysql",
		Source:      AddrToEndpoint(src),
		Destination: AddrToEndpoint(dst),
		DBOperation: op,
		DBTable:     table,
		RawRequest:  packet,
	}, nil
}

// extractMySQLSQLOperation tries to extract the SQL operation and table name from the packet.
// WARNING: This is a naive implementation based on simple string splitting.
// For production, use a real SQL or wire protocol parser to handle all edge cases and SQL dialects.
// TODO: Integrate a proper SQL parser or MySQL wire protocol parser for robust extraction.
var (
    selectRe = regexp.MustCompile(`(?i)^SELECT\s+.*\s+FROM\s+([^\s;]+)`)
    insertRe = regexp.MustCompile(`(?i)^INSERT\s+INTO\s+([^\s(]+)`)
    updateRe = regexp.MustCompile(`(?i)^UPDATE\s+([^\s]+)`)
    deleteRe = regexp.MustCompile(`(?i)^DELETE\s+FROM\s+([^\s;]+)`)
)

func extractMySQLSQLOperation(packet []byte) (string, string) {
    sql := strings.TrimSpace(string(packet))
    qUpper := strings.ToUpper(sql)
    switch {
    case strings.HasPrefix(qUpper, "SELECT"):
        if m := selectRe.FindStringSubmatch(sql); m != nil {
            return "SELECT", m[1]
        }
        return "SELECT", ""
    case strings.HasPrefix(qUpper, "INSERT"):
        if m := insertRe.FindStringSubmatch(sql); m != nil {
            return "INSERT", m[1]
        }
        return "INSERT", ""
    case strings.HasPrefix(qUpper, "UPDATE"):
        if m := updateRe.FindStringSubmatch(sql); m != nil {
            return "UPDATE", m[1]
        }
        return "UPDATE", ""
    case strings.HasPrefix(qUpper, "DELETE"):
        if m := deleteRe.FindStringSubmatch(sql); m != nil {
            return "DELETE", m[1]
        }
        return "DELETE", ""
    default:
        return "", ""
    }
}
