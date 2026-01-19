package db

import (
	"fmt"

	"github.com/febritecno/sqdesk/internal/config"
	"github.com/jmoiron/sqlx"
)

// PostgresConnector implements Connector for PostgreSQL
type PostgresConnector struct {
	BaseConnector
}

// NewPostgresConnector creates a new PostgreSQL connector
func NewPostgresConnector(cfg *config.DatabaseConfig) *PostgresConnector {
	return &PostgresConnector{
		BaseConnector: BaseConnector{
			config: cfg,
			driver: "postgres",
		},
	}
}

// Connect establishes connection to PostgreSQL database
func (c *PostgresConnector) Connect() error {
	sslmode := c.config.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.config.Host,
		c.config.Port,
		c.config.User,
		c.config.Password,
		c.config.Database,
		sslmode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	c.db = db
	return nil
}

// GetTables returns list of tables in the database
func (c *PostgresConnector) GetTables() ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
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
func (c *PostgresConnector) GetColumns(tableName string) ([]Column, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' as is_nullable,
			COALESCE(pk.is_pk, false) as is_pk
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT kcu.column_name, true as is_pk
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu 
				ON tc.constraint_name = kcu.constraint_name
			WHERE tc.table_name = $1 
			AND tc.constraint_type = 'PRIMARY KEY'
		) pk ON c.column_name = pk.column_name
		WHERE c.table_schema = 'public' 
		AND c.table_name = $1
		ORDER BY c.ordinal_position
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
func (c *PostgresConnector) GetSchema() (*Schema, error) {
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
func (c *PostgresConnector) GetDatabases() ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT datname 
		FROM pg_database 
		WHERE datistemplate = false 
		ORDER BY datname
	`

	var databases []string
	if err := c.db.Select(&databases, query); err != nil {
		return nil, fmt.Errorf("failed to get databases: %w", err)
	}

	return databases, nil
}

// SwitchDatabase switches to a different database
func (c *PostgresConnector) SwitchDatabase(dbName string) error {
	// Close current connection
	if c.db != nil {
		c.db.Close()
	}

	// Update config and reconnect
	c.config.Database = dbName
	return c.Connect()
}
