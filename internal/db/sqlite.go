package db

import (
	"fmt"

	"github.com/febritecno/sqdesk-cli/internal/config"
	"github.com/jmoiron/sqlx"
)

// SQLiteConnector implements Connector for SQLite
type SQLiteConnector struct {
	BaseConnector
}

// NewSQLiteConnector creates a new SQLite connector
func NewSQLiteConnector(cfg *config.DatabaseConfig) *SQLiteConnector {
	return &SQLiteConnector{
		BaseConnector: BaseConnector{
			config: cfg,
			driver: "sqlite3",
		},
	}
}

// Connect establishes connection to SQLite database
func (c *SQLiteConnector) Connect() error {
	// For SQLite, the database path is stored in Host or Database field
	dbPath := c.config.Database
	if dbPath == "" {
		dbPath = c.config.Host
	}

	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	c.db = db
	return nil
}

// GetTables returns list of tables in the database
func (c *SQLiteConnector) GetTables() ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT name 
		FROM sqlite_master 
		WHERE type = 'table' 
		AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`

	var tables []string
	if err := c.db.Select(&tables, query); err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	return tables, nil
}

// GetColumns returns columns for a specific table
func (c *SQLiteConnector) GetColumns(tableName string) ([]Column, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	rows, err := c.db.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue interface{}

		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		columns = append(columns, Column{
			Name:     name,
			Type:     colType,
			Nullable: notNull == 0,
			IsPK:     pk == 1,
		})
	}

	return columns, nil
}

// GetSchema returns the complete database schema
func (c *SQLiteConnector) GetSchema() (*Schema, error) {
	tables, err := c.GetTables()
	if err != nil {
		return nil, err
	}

	schema := &Schema{
		Tables: make(map[string]Table),
	}

	for _, tableName := range tables {
		columns, err := c.GetColumns(tableName)
		if err != nil {
			continue // Skip tables we can't read
		}
		schema.Tables[tableName] = Table{
			Name:    tableName,
			Columns: columns,
		}
	}

	return schema, nil
}

// GetDatabases returns list of databases (SQLite has only one)
func (c *SQLiteConnector) GetDatabases() ([]string, error) {
	return []string{c.config.Database}, nil
}

// SwitchDatabase is not supported for SQLite
func (c *SQLiteConnector) SwitchDatabase(dbName string) error {
	return fmt.Errorf("SQLite does not support switching databases")
}
