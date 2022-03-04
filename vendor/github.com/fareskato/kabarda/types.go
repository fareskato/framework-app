package kabarda

import "database/sql"

//////////////////////////
// Kabarda framework types
//////////////////////////

// config type: Kabarda configuration type
type config struct {
	port          string
	templateEngin string
	cookie        cookieConfig // holds all scs.SessionManage configuration fields
	sessionType   string       // session store type: redis, mysql ...ect
	dbConfig      databaseConfig
	redis         RedisConfig
}

// Server kabarda server
type Server struct {
	ServerName string
	Port       string
	Secure     bool
	URL        string
}

// initPaths type: contains the application root path and all application dirs
type initPaths struct {
	rootPath    string
	folderNames []string
}

// cookieConfig holds sec session cookie config fields: refer to scs docs
type cookieConfig struct {
	name     string
	lifetime string
	persist  string
	secure   string
	domain   string
}

// databaseConfig holds the dsn(database connection string and database type(postgres, mysql ...etc))
type databaseConfig struct {
	dsn      string
	database string
}

// DataBase holds the database type and the connection pool
type DataBase struct {
	DataBaseType string
	Pool         *sql.DB
}

type RedisConfig struct {
	host     string
	password string
	prefix   string
}
