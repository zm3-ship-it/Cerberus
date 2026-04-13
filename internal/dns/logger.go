package dns

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type QueryRecord struct {
	ID        int64  `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Domain    string `json:"domain"`
	Type      string `json:"qtype"`
	ClientIP  string `json:"client_ip"`
	ClientMAC string `json:"client_mac"`
	Device    string `json:"device"`
	Blocked   bool   `json:"blocked"`
	Answer    string `json:"answer"`
}

type Logger struct {
	db *sql.DB
	mu sync.Mutex
}

func NewLogger(dbPath string) (*Logger, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// Set pragmas separately — modernc/sqlite DSN doesn't use query params
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=5000")

	schema := `
	CREATE TABLE IF NOT EXISTS dns_queries (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp  INTEGER NOT NULL,
		domain     TEXT NOT NULL,
		qtype      TEXT NOT NULL DEFAULT 'A',
		client_ip  TEXT NOT NULL,
		client_mac TEXT NOT NULL DEFAULT '',
		device     TEXT NOT NULL DEFAULT '',
		blocked    INTEGER NOT NULL DEFAULT 0,
		answer     TEXT NOT NULL DEFAULT ''
	);
	CREATE INDEX IF NOT EXISTS idx_queries_ts ON dns_queries(timestamp);
	CREATE INDEX IF NOT EXISTS idx_queries_domain ON dns_queries(domain);
	CREATE INDEX IF NOT EXISTS idx_queries_client ON dns_queries(client_ip);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}

	return &Logger{db: db}, nil
}

func (l *Logger) Log(rec QueryRecord) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if rec.Timestamp == 0 {
		rec.Timestamp = time.Now().Unix()
	}

	_, err := l.db.Exec(
		`INSERT INTO dns_queries (timestamp, domain, qtype, client_ip, client_mac, device, blocked, answer)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.Timestamp, rec.Domain, rec.Type, rec.ClientIP, rec.ClientMAC, rec.Device, rec.Blocked, rec.Answer,
	)
	return err
}

func (l *Logger) Query(since, until int64, clientIP, domain string, limit int) ([]QueryRecord, error) {
	if limit <= 0 || limit > 10000 {
		limit = 500
	}

	q := `SELECT id, timestamp, domain, qtype, client_ip, client_mac, device, blocked, answer
	      FROM dns_queries WHERE timestamp >= ? AND timestamp <= ?`
	args := []interface{}{since, until}

	if clientIP != "" {
		q += " AND client_ip = ?"
		args = append(args, clientIP)
	}
	if domain != "" {
		q += " AND domain LIKE ?"
		args = append(args, "%"+domain+"%")
	}

	q += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := l.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []QueryRecord
	for rows.Next() {
		var r QueryRecord
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.Domain, &r.Type, &r.ClientIP, &r.ClientMAC, &r.Device, &r.Blocked, &r.Answer); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (l *Logger) TopDomains(since int64, clientIP string, limit int) ([]DomainCount, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	q := `SELECT domain, COUNT(*) as cnt FROM dns_queries WHERE timestamp >= ?`
	args := []interface{}{since}

	if clientIP != "" {
		q += " AND client_ip = ?"
		args = append(args, clientIP)
	}

	q += " GROUP BY domain ORDER BY cnt DESC LIMIT ?"
	args = append(args, limit)

	rows, err := l.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []DomainCount
	for rows.Next() {
		var d DomainCount
		if err := rows.Scan(&d.Domain, &d.Count); err != nil {
			return nil, err
		}
		results = append(results, d)
	}
	return results, rows.Err()
}

type DomainCount struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

func (l *Logger) Stats(since int64) (*StatsResult, error) {
	var s StatsResult

	err := l.db.QueryRow(
		`SELECT COUNT(*), COALESCE(COUNT(DISTINCT domain),0), COALESCE(COUNT(DISTINCT client_ip),0), COALESCE(SUM(blocked),0)
		 FROM dns_queries WHERE timestamp >= ?`, since,
	).Scan(&s.TotalQueries, &s.UniqueDomains, &s.UniqueClients, &s.BlockedQueries)

	return &s, err
}

type StatsResult struct {
	TotalQueries   int `json:"total_queries"`
	UniqueDomains  int `json:"unique_domains"`
	UniqueClients  int `json:"unique_clients"`
	BlockedQueries int `json:"blocked_queries"`
}

func (l *Logger) Purge(olderThan int64) (int64, error) {
	res, err := l.db.Exec("DELETE FROM dns_queries WHERE timestamp < ?", olderThan)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (l *Logger) Close() error {
	return l.db.Close()
}
