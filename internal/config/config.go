package config

type DatabaseConfig struct {
	DBURL    string `env:"DB_URL,required"`
	MinConns int32  `env:"PG_POOL_MIN_CONN,default=1"`
	MaxConns int32  `env:"PG_POOL_MAX_CONN,default=10"`
}
