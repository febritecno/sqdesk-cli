package db

import (
	"fmt"

	"github.com/febritecno/sqdesk/internal/config"
	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Column represents a database column
type Column struct {
	Name     string
	Type     string
	Nullable bool
	IsPK     bool
}

// Table represents a database table with its columns
type Table struct {
	Name    string
	Columns []Column
}

// Schema represents the database schema
type Schema struct {
	Tables map[string]Table
}

// Connector interface for database operations
type Connector interface {
	Connect() error
	Close() error
	IsConnected() bool
	Query(sql string) ([]map[string]interface{}, []string, error)
	Execute(sql string) (int64, error)
	GetTables() ([]string, error)
	GetColumns(tableName string) ([]Column, error)
	GetSchema() (*Schema, error)
	GetDatabases() ([]string, error)
	SwitchDatabase(dbName string) error
	GetDriverName() string
	GetDatabaseName() string
}

// BaseConnector implements common database operations
type BaseConnector struct {
	config *config.DatabaseConfig
	db     *sqlx.DB
	driver string
}

// NewConnector creates a new database connector based on driver type
func NewConnector(cfg *config.DatabaseConfig) (Connector, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is nil")
	}

	switch cfg.Driver {
	case "postgres", "postgresql":
		return NewPostgresConnector(cfg), nil
	case "mysql":
		return NewMySQLConnector(cfg), nil
	case "sqlite", "sqlite3":
		return NewSQLiteConnector(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
}

// IsConnected checks if database is connected
func (c *BaseConnector) IsConnected() bool {
	if c.db == nil {
		return false
	}
	return c.db.Ping() == nil
}

// Close closes the database connection
func (c *BaseConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// GetDriverName returns the driver name
func (c *BaseConnector) GetDriverName() string {
	return c.driver
}

// GetDatabaseName returns the database name
func (c *BaseConnector) GetDatabaseName() string {
	return c.config.Database
}

// Query executes a SELECT query and returns results
func (c *BaseConnector) Query(sql string) ([]map[string]interface{}, []string, error) {
	if c.db == nil {
		return nil, nil, fmt.Errorf("not connected to database")
	}

	rows, err := c.db.Queryx(sql)
	if err != nil {
		return nil, nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			return nil, nil, fmt.Errorf("scan error: %w", err)
		}
		
		// Convert []byte to string for better display
		for k, v := range row {
			if b, ok := v.([]byte); ok {
				row[k] = string(b)
			}
		}
		
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("rows error: %w", err)
	}

	return results, columns, nil
}

// Execute runs an INSERT/UPDATE/DELETE query
func (c *BaseConnector) Execute(sql string) (int64, error) {
	if c.db == nil {
		return 0, fmt.Errorf("not connected to database")
	}

	result, err := c.db.Exec(sql)
	if err != nil {
		return 0, fmt.Errorf("execute error: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return affected, nil
}


