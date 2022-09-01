package configuration

type Config struct {
	Port       int    `envconfig:"GOLINKS_PORT" default:"8080"`
	DbType     string `envconfig:"GOLINKS_DB_TYPE" default:"sqllite"`
	DbDSN      string `envconfig:"GOLINKS_DB" default:"./data/golinks.db"`
	StaticPath string `envconfig:"GOLINKS_STATIC" default:"./static/"`
}
