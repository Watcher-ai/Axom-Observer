package protocols

import (
	"net"
	"testing"
)

func TestExtractMySQLSQLOperation(t *testing.T) {
    tests := []struct {
        input     string
        wantOp    string
        wantTable string
    }{
        {"SELECT * FROM users", "SELECT", "users"},
        {"INSERT INTO orders VALUES (1)", "INSERT", "orders"},
        {"UPDATE products SET price=1", "UPDATE", "products"},
        {"DELETE FROM logs", "DELETE", "logs"},
        {"", "", ""},
    }
    for _, tt := range tests {
        op, table := extractMySQLSQLOperation([]byte(tt.input))
        if op != tt.wantOp || table != tt.wantTable {
            t.Errorf("input=%q got op=%q table=%q, want op=%q table=%q", tt.input, op, table, tt.wantOp, tt.wantTable)
        }
    }
}

func TestProcessMySQL(t *testing.T) {
	raw := []byte("SELECT * FROM users")
	src := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
	dst := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 3306}
	sig, err := ProcessMySQL(raw, src, dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sig.Protocol != "mysql" {
		t.Errorf("expected protocol mysql, got %s", sig.Protocol)
	}
	if sig.DBOperation != "SELECT" {
		t.Errorf("expected DBOperation SELECT, got %s", sig.DBOperation)
	}
}
