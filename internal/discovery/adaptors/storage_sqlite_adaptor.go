package adaptors

import (
	"database/sql"
	"github.com/aminsalami/repartido/internal/discovery"
	"github.com/aminsalami/repartido/internal/discovery/ports"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SqliteCacheStorage implements ports.CacheStorage
type SqliteCacheStorage struct {
	db *sql.DB
}

var logger *zap.Logger

// -------------------

func init() {
	logger, _ = zap.NewDevelopment()
}

func NewSqliteCacheStorage() ports.CacheStorage {
	// TODO get dbName from the config
	dbName := "cache_nodes.sql.db"
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		logger.Fatal(err.Error())
	}
	if err != nil {
		logger.Fatal(err.Error())
	}
	// Migrate all the ".sql" files oder by their name
	err = sqlMigrate(db)
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Info("Successful Migration.")
	// TODO handle db.close() method gracefully
	return &SqliteCacheStorage{
		db: db,
	}
}

func sqlMigrate(db *sql.DB) error {
	// Default location is in current directory
	return filepath.WalkDir("./internal/discovery/adaptors", func(path string, d fs.DirEntry, err error) error {
		// Ignore directories. Ignore any file that is not ".sql"
		if d.IsDir() || strings.Index(path, ".sql") == -1 {
			return nil
		}

		logger.Info("Migrating " + path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = db.Exec(string(data))
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *SqliteCacheStorage) Save(node *discovery.CacheNode) error {
	logger.Info("Saving into table `cache_node`", zap.Object("node", node))
	_, err := s.db.Exec(
		"Insert INTO cache_node (name, host, port, last_ping, ram_size) values (?, ?, ?, ?, ?)",
		node.Name, node.Host, node.Port, node.LastPing, node.RamSize,
	)
	return err
}

func (s *SqliteCacheStorage) Get() (discovery.CacheNode, error) {
	logger.Info("Return empty CacheNode from sqlite!")
	return discovery.CacheNode{}, nil
}

func (s *SqliteCacheStorage) List() ([]*discovery.CacheNode, error) {
	logger.Info("Return a list of empty CacheNodes from sqlite!")
	return []*discovery.CacheNode{}, nil
}

func (s *SqliteCacheStorage) Delete(node *discovery.CacheNode) error {
	logger.Info("Deleting the node")
	_, err := s.db.Exec("DELETE FROM cache_node WHERE id = ? AND host = ? AND port = ?", node.Id, node.Host, node.Port)
	return err
}

func (s *SqliteCacheStorage) Close() error {
	e := s.db.Close()
	if e != nil {
		logger.Error(e.Error())
	}
	return e
}
