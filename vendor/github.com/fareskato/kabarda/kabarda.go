package kabarda

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/fareskato/kabarda/cache"
	"github.com/fareskato/kabarda/mailer"
	"github.com/fareskato/kabarda/render"
	"github.com/fareskato/kabarda/session"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//////////////////////////
// Kabarda framework main
/////////////////////////

const version = "1.0.0"

var appRedisCache *cache.RedisCache

// Kabarda type: is the main application type
type Kabarda struct {
	AppName       string
	Debug         bool
	Version       string
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	RootPath      string
	config        config
	Routes        *chi.Mux
	Render        *render.Render
	JetViews      *jet.Set
	Session       *scs.SessionManager
	DB            DataBase
	EncryptionKey string
	Cache         cache.Cache
	Mail          mailer.Mail
	Server        Server
}

// New create new instance of Kabarda app
func (k *Kabarda) New(rootPath string) error {
	// application init paths
	pathConfig := initPaths{
		rootPath: rootPath,
		// all needed dirs for the application
		folderNames: []string{"handlers", "views", "data", "tmp", "public", "middlewares", "migrations", "mail", "logs"},
	}
	// call Init to create application needed folders
	err := k.Init(pathConfig)
	if err != nil {
		return err
	}
	// check if .env file
	err = k.CheckDotEnv(rootPath)
	if err != nil {
		return err
	}
	// load .env file
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	// loggers
	infoLog, errorLog := k.startLoggers()

	// connect to database
	if os.Getenv("DATABASE_TYPE") != "" {
		dbPool, err := k.OpenDB(os.Getenv("DATABASE_TYPE"), k.BuildDSN())
		if err != nil {
			k.ErrorLog.Println(err)
			os.Exit(1)
		}
		k.DB = DataBase{
			DataBaseType: os.Getenv("DATABASE_TYPE"),
			Pool:         dbPool,
		}
	}
	// check if redis needed or not
	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		appRedisCache = k.createClientRedisCache()
		k.Cache = appRedisCache
	}

	k.InfoLog = infoLog
	k.ErrorLog = errorLog

	// debug from .env
	k.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))

	// version
	k.Version = version

	// application root path
	k.RootPath = rootPath

	// mail
	k.Mail = k.createMailer()

	// roues: convert to *chi.Mux
	k.Routes = k.initRoutes().(*chi.Mux)

	// config
	k.config = config{
		port:          os.Getenv("PORT"),
		templateEngin: os.Getenv("RENDERER"),
		// scs cookie config
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSIST"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		dbConfig: databaseConfig{
			dsn:      k.BuildDSN(),
			database: os.Getenv("DATABASE_TYPE"),
		},
		redis: RedisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	// kabarda server
	secure := true
	if strings.ToLower(os.Getenv("SECURE")) == "false" {
		secure = false
	}
	k.Server = Server{
		ServerName: os.Getenv("SERVER_NAME"),
		Port:       os.Getenv("PORT"),
		Secure:     secure,
		URL:        os.Getenv("APP_URL"),
	}

	// jet templating engine: check if we are in development or production
	if k.Debug {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
			jet.InDevelopmentMode(),
		)
		// set views
		k.JetViews = views
	} else {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		)
		// set views
		k.JetViews = views
	}

	// create session
	kabardaSession := session.Session{
		CookieName:     k.config.cookie.name,
		CookieDomain:   k.config.cookie.domain,
		CookieLifeTime: k.config.cookie.lifetime,
		CookiePersist:  k.config.cookie.persist,
		CookieSecure:   k.config.cookie.secure,
		SessionType:    k.config.sessionType,
	}
	switch k.config.sessionType {
	case "redis":
		kabardaSession.RedisPool = appRedisCache.Conn
	case "mysql", "mariadb", "postgres", "postgresql":
		kabardaSession.DBPool = k.DB.Pool
	}
	k.Session = kabardaSession.InitSession()

	// set application key
	k.EncryptionKey = os.Getenv("KEY")

	// init the renderer
	k.createRenderer()

	// run the mail functionality in the background
	go k.Mail.ListenForEmail()

	// all good
	return nil
}

// Init :initialize Kabarda framework
func (k *Kabarda) Init(p initPaths) error {
	// application root path
	root := p.rootPath
	for _, path := range p.folderNames {
		// create folder if not exists
		err := k.CreateDirIfNotExists(root + "/" + path)
		if err != nil {
			return err
		}
	}
	return nil
}

// startLoggers create info and error loggers
func (k *Kabarda) startLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errLog *log.Logger
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	return infoLog, errLog
}

// StartServer starts the web serve
func (k *Kabarda) StartServer() {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler:      k.Routes,
		ErrorLog:     k.ErrorLog,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	// close the db connection
	defer k.DB.Pool.Close()

	k.InfoLog.Printf("Listening on port: %s", os.Getenv("PORT"))
	err := server.ListenAndServe()
	if err != nil {
		k.ErrorLog.Fatal(err)
	}
}

// createRenderer create Kabarda renderer
func (k *Kabarda) createRenderer() {
	theRenderer := render.Render{
		RenderingEngine: k.config.templateEngin,
		RootPath:        k.RootPath,
		Port:            k.config.port,
		JetViews:        k.JetViews,
		Session:         k.Session,
	}
	k.Render = &theRenderer
}

// createMailer attach mailer.Mail to Kabarda
func (k *Kabarda) createMailer() mailer.Mail {
	// port
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	ml := mailer.Mail{
		Domain:        os.Getenv("MAIL_DOMAIN"),
		TemplatesPath: k.RootPath + "/mail",
		Host:          os.Getenv("SMTP_HOST"),
		Port:          port,
		UserName:      os.Getenv("SMTP_USERNAME"),
		Password:      os.Getenv("SMTP_PASSWORD"),
		Encryption:    os.Getenv("SMTP_ENCRYPTION"),
		FromAddress:   os.Getenv("SMTP_FROM_ADDRESS"),
		FromName:      os.Getenv("SMTP_FROM_NAME"),
		Jobs:          make(chan mailer.Message, 20),
		Result:        make(chan mailer.Result, 20),
		Api:           os.Getenv("MAILER_API"),
		ApiKey:        os.Getenv("MAILER_KEY"),
		ApiUrl:        os.Getenv("MAILER_URL"),
	}
	return ml
}

// BuildDSN will generate the Database connection string(dsn) according to the database type
// supports postgres, Mariadb and mysql
func (k *Kabarda) BuildDSN() string {
	var dsn string
	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		// here we left password because of the default password on dev environment is empty
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"))
		// if the password set on .env file
		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}
	default:

	}
	return dsn
}

// connection to redis
func (k *Kabarda) createRedisPool() *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				k.config.redis.host,
				redis.DialPassword(k.config.redis.password),
			)
		},
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (k *Kabarda) createClientRedisCache() *cache.RedisCache {
	cac := cache.RedisCache{
		Conn:   k.createRedisPool(),
		Prefix: k.config.redis.prefix,
	}
	return &cac
}
