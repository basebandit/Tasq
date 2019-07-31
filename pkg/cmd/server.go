package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	"bitbucket.org/liamstask/goose/lib/goose"
	"github.com/basebandit/go-grpc/pkg/protocol/grpc"
	"github.com/basebandit/go-grpc/pkg/protocol/rest"
	v1 "github.com/basebandit/go-grpc/pkg/service/v1"
)

//Config is our configuration for our server
type Config struct {

	//gRPC is the TCP port to listen by gRPC server
	GRPCPort string
	//HTTPPort is the TCP port to listen for HTTP/REST gateway connections
	HTTPPort string
	//DBHost is the host of database
	DBHost string

	//DBUser is the username to connect to  database
	DBUser string

	//DBPassword is the password to connect to database
	DBPassword string

	//DBName is the name of the database
	DBName string

	//DBMigrations is the migrations path of the db schema migrations
	DBMigrations string
}

//RunServer runs gRPC server and HTTP gateway
func RunServer() error {
	ctx := context.Background()

	//get configuration

	var cfg Config
	flag.StringVar(&cfg.GRPCPort, "port", "", "gRPC port to bind")
	flag.StringVar(&cfg.DBHost, "host", "", "Database host")
	flag.StringVar(&cfg.DBUser, "user", "", "Database user")
	flag.StringVar(&cfg.DBPassword, "password", "", "Database password")
	flag.StringVar(&cfg.DBName, "db", "", "Database name")
	flag.StringVar(&cfg.DBMigrations, "migrations", "", "Database schema migrations")
	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: '%s'", cfg.GRPCPort)
	}

	if len(cfg.HTTPPort) == 0 {
		return fmt.Errorf("invalid TCP port for HTTP gateway: '%s'", cfg.HTTPPort)
	}

	if len(cfg.DBMigrations) == 0 {
		return fmt.Errorf("invalid database migrations path: '%s'", cfg.DBMigrations)
	}

	//Lets chek if migrations Directory path exists
	if _, err := os.Stat(cfg.DBMigrations); os.IsNotExist(err) {
		return fmt.Errorf("you need to provide the path to directory where your migrations are stored:  -migrations <migrations_path>")
	}
	//add MySQL driver specific parameter to parse date/time
	//Drop it for another database
	param := "parseTime=true"

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBName, param)

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	//Lets run our migrations first
	if err := runMigrations(db, cfg.DBMigrations); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	v1API := v1.NewToDoServiceServer(db)

	//run HTTP/REST gateway
	go func() {
		_ = rest.RunServer(ctx, cfg.GRPCPort, cfg.HTTPPort)
	}()

	return grpc.RunServer(ctx, v1API, cfg.GRPCPort)
}

func runMigrations(db *sql.DB, migrationsPath string) error {

	// Setup the goose configuration
	migrateConf := &goose.DBConf{
		MigrationsDir: migrationsPath,
		Env:           "production",
		Driver: goose.DBDriver{
			Name:    "mysql",
			OpenStr: "mars:mars@/ToDo?parseTime=true,charset=utf8",
			Import:  "database/sql",
			Dialect: &goose.MySqlDialect{},
		},
	}
	// Get the latest possible migration
	latest, err := goose.GetMostRecentDBVersion(migrateConf.MigrationsDir)
	if err != nil {
		return err
	}

	// Migrate up to the latest version
	err = goose.RunMigrationsOnDb(migrateConf, migrateConf.MigrationsDir, latest, db)
	if err != nil {
		return err
	}
	return err
}
