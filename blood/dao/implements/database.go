package implements

import (
	"context"
	"fmt"
	"log"
	"time"

	"blood/config"
	neo4j "blood/dao/neo4j"
	pg "blood/dao/pg"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库连接管理器
type Database struct {
	PostgresDB *gorm.DB
	NeuroDB    *neo4j.NeuroDB
	Config     *config.DatabaseConfig
}

// DefaultDatabaseConfig 返回默认数据库配置
func DefaultDatabaseConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Host:      "localhost",
		Port:      5432,
		User:      "postgres",
		Password:  "123456",
		DBName:    "hippo",
		SSLMode:   "disable",
		TimeZone:  "Asia/Shanghai",
		Dimension: 2560,
	}
}

// NewDatabase 创建新的数据库连接实例
func NewDatabase(config *config.DatabaseConfig) (*Database, error) {
	if config == nil {
		config = DefaultDatabaseConfig()
	}
	neuroDB := neo4j.NewNeuroDB(7687)
	db := &Database{
		Config:  config,
		NeuroDB: neuroDB,
	}
	err := CreateDatabase(db.Config)
	if err != nil {
		return nil, err
	}
	err = db.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	err = pg.InitDatabase(db.PostgresDB, db.Config.Dimension)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}
	return db, nil
}

// Connect 连接到PostgreSQL数据库
func (d *Database) Connect() error {
	// 构建数据库连接字符串
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		d.Config.Host,
		d.Config.User,
		d.Config.Password,
		d.Config.DBName,
		d.Config.Port,
		d.Config.SSLMode,
		d.Config.TimeZone,
	)

	// 配置GORM日志
	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second,   // 慢查询阈值
				LogLevel:                  logger.Silent, // 日志级别
				IgnoreRecordNotFoundError: true,          // 忽略记录未找到错误
				Colorful:                  false,         // 禁用彩色打印
			},
		),
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	d.PostgresDB = db

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间

	log.Printf("Successfully connected to PostgreSQL database: %s", d.Config.DBName)
	return nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	if d.PostgresDB == nil {
		return nil
	}

	sqlDB, err := d.PostgresDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	//d.NeuroDB.ExecuteQuery()
	log.Println("Database connection closed successfully")
	return nil
}

// Ping 测试数据库连接
func (d *Database) Ping() error {
	if d.PostgresDB == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := d.PostgresDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// GetPostgresSQL 获取GORM数据库实例
func (d *Database) GetPostgresSQL() *gorm.DB {
	return d.PostgresDB
}

func (d *Database) GetNeuroDB() *neo4j.NeuroDB {
	return d.NeuroDB
}

// AutoMigrate 自动迁移数据库表结构
func (d *Database) AutoMigrate(models ...interface{}) error {
	if d.PostgresDB == nil {
		return fmt.Errorf("database connection is nil")
	}

	err := d.PostgresDB.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Printf("Successfully migrated %d models", len(models))
	return nil
}

// CreateDatabase 创建数据库（如果不存在）
func CreateDatabase(config *config.DatabaseConfig) error {
	// 连接到postgres数据库来创建目标数据库
	tempConfig := *config
	tempConfig.DBName = "postgres"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		tempConfig.Host,
		tempConfig.User,
		tempConfig.Password,
		tempConfig.DBName,
		tempConfig.Port,
		tempConfig.SSLMode,
		tempConfig.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}

	// 检查数据库是否存在
	var exists bool
	err = db.Raw("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = ?)", config.DBName).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		// 创建数据库
		err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", config.DBName)).Error
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", config.DBName, err)
		}
		log.Printf("Database %s created successfully", config.DBName)
	} else {
		log.Printf("Database %s already exists", config.DBName)
	}

	// 关闭临时连接
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	return nil
}

// DropDatabase 删除数据库
func DropDatabase(config *config.DatabaseConfig) error {
	// 连接到postgres数据库来删除目标数据库
	tempConfig := *config
	tempConfig.DBName = "postgres"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		tempConfig.Host,
		tempConfig.User,
		tempConfig.Password,
		tempConfig.DBName,
		tempConfig.Port,
		tempConfig.SSLMode,
		tempConfig.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}

	// 删除数据库
	err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", config.DBName)).Error
	if err != nil {
		return fmt.Errorf("failed to drop database %s: %w", config.DBName, err)
	}

	log.Printf("Database %s dropped successfully", config.DBName)

	// 关闭临时连接
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	return nil
}

// HasTable 检查表是否存在
func (d *Database) HasTable(tableName string) bool {
	if d.PostgresDB == nil {
		return false
	}
	return d.PostgresDB.Migrator().HasTable(tableName)
}

// DropTable 删除表
func (d *Database) DropTable(models ...any) error {
	if d.PostgresDB == nil {
		return fmt.Errorf("database connection is nil")
	}

	err := d.PostgresDB.Migrator().DropTable(models...)
	if err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	log.Printf("Successfully dropped %d tables", len(models))
	return nil
}

// CreateIndex 创建索引
func (d *Database) CreateIndex(model any, indexName string, columns ...string) error {
	if d.PostgresDB == nil {
		return fmt.Errorf("database connection is nil")
	}

	err := d.PostgresDB.Migrator().CreateIndex(model, indexName)
	if err != nil {
		return fmt.Errorf("failed to create index %s: %w", indexName, err)
	}

	log.Printf("Successfully created index: %s", indexName)
	return nil
}

// DropIndex 删除索引
func (d *Database) DropIndex(model any, indexName string) error {
	if d.PostgresDB == nil {
		return fmt.Errorf("database connection is nil")
	}

	err := d.PostgresDB.Migrator().DropIndex(model, indexName)
	if err != nil {
		return fmt.Errorf("failed to drop index %s: %w", indexName, err)
	}

	log.Printf("Successfully dropped index: %s", indexName)
	return nil
}

// Stats DatabaseStats 数据库连接统计信息
type Stats struct {
	MaxOpenConnections int `json:"max_open_connections"`
	OpenConnections    int `json:"open_connections"`
	InUse              int `json:"in_use"`
	Idle               int `json:"idle"`
}

// GetStats 获取数据库连接统计信息
func (d *Database) GetStats() (*Stats, error) {
	if d.PostgresDB == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	sqlDB, err := d.PostgresDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
	return &Stats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
	}, nil
}

// SetConnectionPool 设置连接池参数
func (d *Database) SetConnectionPool(maxIdleConns, maxOpenConns int, connMaxLifetime time.Duration) error {
	if d.PostgresDB == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := d.PostgresDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	log.Printf("Connection pool configured: MaxIdle=%d, MaxOpen=%d, MaxLifetime=%v",
		maxIdleConns, maxOpenConns, connMaxLifetime)
	return nil
}

// IsHealthy 检查数据库健康状态
func (d *Database) IsHealthy() bool {
	if d.PostgresDB == nil {
		return false
	}

	err := d.Ping()
	return err == nil
}

// Transaction 执行事务
func (d *Database) Transaction(fn func(*gorm.DB) error) error {
	if d.PostgresDB == nil {
		return fmt.Errorf("database connection is nil")
	}

	return d.PostgresDB.Transaction(fn)
}

// WithContext 使用上下文
func (d *Database) WithContext(ctx context.Context) *gorm.DB {
	if d.PostgresDB == nil {
		return nil
	}
	return d.PostgresDB.WithContext(ctx)
}

// SetLogLevel 设置日志级别
func (d *Database) SetLogLevel(level logger.LogLevel) {
	if d.PostgresDB != nil {
		d.PostgresDB.Logger = d.PostgresDB.Logger.LogMode(level)
	}
}

// EnableDebugMode 启用调试模式
func (d *Database) EnableDebugMode() {
	if d.PostgresDB != nil {
		d.PostgresDB = d.PostgresDB.Debug()
	}
}

// GetConnectionString 获取连接字符串（隐藏密码）
func (d *Database) GetConnectionString() string {
	if d.Config == nil {
		return ""
	}

	return fmt.Sprintf("host=%s user=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		d.Config.Host,
		d.Config.User,
		d.Config.DBName,
		d.Config.Port,
		d.Config.SSLMode,
		d.Config.TimeZone,
	)
}

// Reconnect 重新连接数据库
func (d *Database) Reconnect() error {
	if d.PostgresDB != nil {
		// 关闭现有连接
		d.Close()
	}
	// 重新连接
	return d.Connect()
}
