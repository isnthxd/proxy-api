package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Database struct {
	Conn *sql.DB
}

type Proxy struct {
	Proxy        string
	Type         string
	ResponseTime int
}

type ProxySource struct {
	Type string
	Url  string
}

func (d *Database) Init(path string) error {
	var err error
	if d.Conn, err = sql.Open("sqlite", path); err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	tables := []string{
		`CREATE TABLE IF NOT EXISTS proxies (
			id INTEGER PRIMARY KEY,
			type TEXT NOT NULL,
			proxy TEXT NOT NULL,
			status TEXT NOT NULL,
			response_time INTEGER DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS sources (
			id INTEGER PRIMARY KEY,
			type TEXT NOT NULL,
			url TEXT NOT NULL
		);`,
	}

	for _, query := range tables {
		if _, err := d.Conn.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func (d *Database) Close() {
	if d.Conn != nil {
		d.Conn.Close()
	}
}

func (d *Database) InsertProxy(proxyType, proxy string) error {
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM proxies WHERE type = ? AND proxy = ?);`
	if err := d.Conn.QueryRow(checkQuery, proxyType, proxy).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check existing proxy: %w", err)
	}

	insertQuery := `INSERT INTO proxies (type, proxy, status, response_time) VALUES (?, ?, 'offline', 0);`
	if _, err := d.Conn.Exec(insertQuery, proxyType, proxy); err != nil && !exists {
		return fmt.Errorf("failed to insert proxy: %w", err)
	}

	return nil
}

func (d *Database) UpdateProxyStatus(proxyType, proxy, status string, responseTime int) error {
	query := `UPDATE proxies SET status = ?, response_time = ? WHERE type = ? AND proxy = ?;`
	if _, err := d.Conn.Exec(query, status, responseTime, proxyType, proxy); err != nil {
		return fmt.Errorf("failed to update proxy: %w", err)
	}
	return nil
}

func (d *Database) GetOnlineProxies(proxyType string, timeout int) []string {
	query := `SELECT proxy, response_time FROM proxies WHERE status = 'online'`
	var args []interface{}

	if proxyType != "all" {
		query += " AND type = ?"
		args = append(args, proxyType)
	}

	rows, err := d.Conn.Query(query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var proxies []string
	for rows.Next() {
		var proxy string
		var responseTime int
		if err := rows.Scan(&proxy, &responseTime); err != nil {
			return nil
		}
		if timeout == -1 || responseTime <= timeout {
			proxies = append(proxies, proxy)
		}
	}

	if err := rows.Err(); err != nil {
		return nil
	}

	return proxies
}

func (d *Database) GetAllProxies() []Proxy {
	query := `SELECT proxy, type FROM proxies;`
	rows, err := d.Conn.Query(query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var proxies []Proxy
	for rows.Next() {
		var proxy Proxy
		if err := rows.Scan(&proxy.Proxy, &proxy.Type); err != nil {
			return nil
		}
		proxies = append(proxies, proxy)
	}

	if err := rows.Err(); err != nil {
		return nil
	}

	return proxies
}

func (d *Database) GetAllProxySources() []ProxySource {
	query := `SELECT type, url FROM sources;`
	rows, err := d.Conn.Query(query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var sources []ProxySource
	for rows.Next() {
		var source ProxySource
		if err := rows.Scan(&source.Type, &source.Url); err != nil {
			return nil
		}
		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return nil
	}

	return sources
}
