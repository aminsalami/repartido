package adaptors

import (
	"database/sql"
	"github.com/aminsalami/repartido/internal/discovery"
	"github.com/aminsalami/repartido/internal/discovery/ports"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// SqliteCacheStorage implements ports.CacheStorage
type SqliteCacheStorage struct {
	db *sql.DB
}

var logger = zap.NewExample().Sugar()

func NewSqliteCacheStorage() ports.CacheStorage {
	dbPath := viper.GetString("db.path")
	if dbPath == "" {
		dbPath = "./"
	}
	dbName := viper.GetString("db.name")
	if dbName == "" {
		dbName = "discovery.conf"
	}
	db, err := sql.Open("sqlite3", dbPath+dbName)
	if err != nil {
		logger.Fatal(err.Error())
	}
	err = db.Ping()
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Info("Successfully pinged the db on " + dbPath + dbName)
	// TODO handle db.close() method gracefully
	return &SqliteCacheStorage{
		db: db,
	}
}

func (s *SqliteCacheStorage) Save(node *discovery.CacheNode) error {
	logger.Info("Saving into table `cache_node`", zap.Object("node", node))
	_, err := s.db.Exec(
		"Insert INTO cache_node (node_id, name, host, port, last_ping, ram_size) values (?, ?, ?, ?, ?, ?)",
		node.Id, node.Name, node.Host, node.Port, node.LastPing, node.RamSize,
	)
	return err
}

func (s *SqliteCacheStorage) Get() (discovery.CacheNode, error) {
	logger.Info("Return empty CacheNode from sqlite!")
	return discovery.CacheNode{}, nil
}

func (s *SqliteCacheStorage) GetById(id string) (*discovery.CacheNode, error) {
	var i int
	node := discovery.CacheNode{}
	res := s.db.QueryRow("SELECT * FROM cache_node WHERE node_id = ?", id)
	if err := res.Scan(&i, &node.Name, &node.Host, &node.Port, &node.LastPing, &node.RamSize); err != nil {
		return &discovery.CacheNode{}, err
	}

	return &node, nil
}

func (s *SqliteCacheStorage) List() ([]*discovery.CacheNode, error) {
	logger.Info("Return a list of empty CacheNodes from sqlite!")
	return []*discovery.CacheNode{}, nil
}

func (s *SqliteCacheStorage) Delete(node *discovery.CacheNode) error {
	logger.Infow("Deleting the node", "node_id", node.Id)
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
