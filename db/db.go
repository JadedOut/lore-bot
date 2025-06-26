package db

import (
	"database/sql"
	//"fmt"
	"log"
	"os"

	"github.com/go-gorp/gorp"
	_redis "github.com/go-redis/redis/v7"
	_ "github.com/lib/pq" //import postgres
)

// DB ...
type DB struct {
	*sql.DB
}

var db *gorp.DbMap

// Init ...
func Init() {

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("database_url doesnt exist")
	}

	var err error
	db, err = ConnectDB(databaseURL)
	if err != nil {
		log.Fatal("failed to connect to db: ", err)
	}

	log.Println("db connection success")

	TestPgVector()
}

// ConnectDB ...
func ConnectDB(dataSourceName string) (*gorp.DbMap, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	//dbmap.TraceOn("[gorp]", log.New(os.Stdout, "golang-gin:", log.Lmicroseconds)) //Trace database requests
	return dbmap, nil
}

// GetDB ...
func GetDB() *gorp.DbMap {
	return db
}

// Pgvector tests
func TestPgVector() {
	log.Println("------------------------testing pgvector------------------------------")

	_, err := db.Db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		log.Printf("Warning: pgvector error: %v ", err)
	}

	queries := []string{
		"DROP TABLE IF EXISTS test_items",
		"CREATE TABLE test_items (id bigserial PRIMARY KEY, embedding vector(3))",
		"INSERT INTO test_items (embedding) VALUES ('[1,2,3]'), ('[4,5,6]')",
	}

	for _, query := range queries {
		_, err := db.Db.Exec(query)
		if err != nil {
			log.Printf("Warning: execute query error: %v", err)
		}
	}

	// Similarity search
	rows, err := db.Db.Query("SELECT id, embedding FROM test_items ORDER BY embedding <-> '[3,1,2]' LIMIT 5")
	if err != nil {
		log.Printf("Warning: similarity search error: %v", err)
		return
	}
	defer rows.Close()

	log.Println("Similarity search results:")

	for rows.Next() {
		var id int
		var embedding string
		if err := rows.Scan(&id, &embedding); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		log.Printf("  ID: %d, Embedding: %s", id, embedding)
	}

	// Clean test table
	_, err = db.Db.Exec("DROP TABLE test_items")
	if err != nil {
		log.Printf("Warning: clean test table error: %v", err)
	}
}

// RedisClient ...
var RedisClient *_redis.Client

// InitRedis ...
func InitRedis(selectDB ...int) {

	var redisHost = os.Getenv("REDIS_HOST")
	var redisPassword = os.Getenv("REDIS_PASSWORD")

	if redisHost == "" {
		log.Println("skipping redis")
		return
	}

	RedisClient = _redis.NewClient(&_redis.Options{
		Addr:     redisHost,
		Password: redisPassword,
		DB:       selectDB[0],
		// DialTimeout:        10 * time.Second,
		// ReadTimeout:        30 * time.Second,
		// WriteTimeout:       30 * time.Second,
		// PoolSize:           10,
		// PoolTimeout:        30 * time.Second,
		// IdleTimeout:        500 * time.Millisecond,
		// IdleCheckFrequency: 500 * time.Millisecond,
		// TLSConfig: &tls.Config{
		// 	InsecureSkipVerify: true,
		// },
	})

}

// GetRedis ...
func GetRedis() *_redis.Client {
	return RedisClient
}
