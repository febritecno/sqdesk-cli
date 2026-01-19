package db

import (
	"fmt"

	"github.com/febritecno/sqdesk/internal/config"
	"github.com/jmoiron/sqlx"
)

// MySQLConnector implements Connector for MySQL
type MySQLConnector struct {
	BaseConnector
}

// NewMySQLConnector creates a new MySQL connector
func NewMySQLConnector(cfg *config.DatabaseConfig) *MySQLConnector {
	return &MySQLConnector{
		BaseConnector: BaseConnector{
			config: cfg,
			driver: "mysql",
		},
	}
}

// Connect establishes connection to MySQL database
func (c *MySQLConnector) Connect() error {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.config.User,
		c.config.Password,
		c.config.Host,
		c.config.Port,
		c.config.Database,
	)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	c.db = db
	return nil
}

// GetTables returns list of tables in the database
func (c *MySQLConnector) GetTables() ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	var tables []string
	if err := c.db.Select(&tables, query); err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	return tables, nil
}

// GetColumns returns columns for a specific table
func (c *MySQLConnector) GetColumns(tableName string) ([]Column, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT 
			COLUMN_NAME as name,
			DATA_TYPE as type,
			IS_NULLABLE = 'YES' as is_nullable,
			COLUMN_KEY = 'PRI' as is_pk
		FROM information_schema.columns 
		WHERE table_schema = DATABASE()
		AND table_name = ?
		ORDER BY ordinal_position
	`

	rows, err := c.db.Queryx(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		var isNullable, isPK bool
		if err := rows.Scan(&col.Name, &col.Type, &isNullable, &isPK); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		col.Nullable = isNullable
		col.IsPK = isPK
		columns = append(columns, col)
	}

	return columns, nil
}

// GetSchema returns the complete database schema
func (c *MySQLConnector) GetSchema() (*Schema, error) {
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

// GetDatabases returns list of all databases on the server
func (c *MySQLConnector) GetDatabases() ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `SHOW DATABASES`

	var databases []string
	if err := c.db.Select(&databases, query); err != nil {
		return nil, fmt.Errorf("failed to get databases: %w", err)
	}

	return databases, nil
}

// SwitchDatabase switches to a different database
func (c *MySQLConnector) SwitchDatabase(dbName string) error {
	if c.db == nil {
		return fmt.Errorf("not connected")
	}

	// Use USE statement for MySQL
	_, err := c.db.Exec("USE " + dbName)
	if err != nil {
		return fmt.Errorf("failed to switch database: %w", err)
	}

	c.config.Database = dbName
	return nil
}
