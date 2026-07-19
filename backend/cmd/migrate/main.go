// Command migrate is the migration CLI: create-db | up | down [n] | status.
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"text/tabwriter"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"example.com/interview-question-002/config"
	"example.com/interview-question-002/internal/infrastructure/db/migrate"
	_ "example.com/interview-question-002/internal/infrastructure/db/migrate/migrations"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "create-db":
		if err := cmdCreateDB(); err != nil {
			fail(err)
		}
	case "up":
		run(cmdUp, os.Args[2:])
	case "down":
		run(cmdDown, os.Args[2:])
	case "status":
		run(cmdStatus, os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Print(`usage: migrate <command> [args]

commands:
  create-db        create the target database (DB_NAME) if it does not exist
  up               apply all pending migrations
  down [n]         rollback the last n migrations (default 1)
  status           list applied and pending migrations

examples:
  go run ./cmd/migrate create-db
  go run ./cmd/migrate up
  go run ./cmd/migrate status
`)
}

var validDBName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString

// cmdCreateDB connects to the "postgres" maintenance database and creates DB_NAME
// if it does not already exist (CREATE DATABASE cannot run in a transaction, so it
// lives outside the migration engine).
func cmdCreateDB() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	target := cfg.DB.Name
	if !validDBName(target) {
		return fmt.Errorf("unsafe database name %q", target)
	}
	maint, err := gorm.Open(postgres.Open(cfg.DB.MaintenanceDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("connect maintenance db: %w", err)
	}
	defer closeDB(maint)

	var exists bool
	if err := maint.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", target).Scan(&exists).Error; err != nil {
		return err
	}
	if exists {
		fmt.Printf("database %q already exists\n", target)
		return nil
	}
	if err := maint.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, target)).Error; err != nil {
		return err
	}
	fmt.Printf("created database %q\n", target)
	return nil
}

func run(fn func(*gorm.DB, []string) error, args []string) {
	db, err := openDB()
	if err != nil {
		fail(err)
	}
	defer closeDB(db)
	if err := fn(db, args); err != nil {
		fail(err)
	}
}

func openDB() (*gorm.DB, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	gormDB, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return gormDB, nil
}

func closeDB(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func cmdUp(db *gorm.DB, _ []string) error {
	ran, err := migrate.Up(db)
	if err != nil {
		return err
	}
	if len(ran) == 0 {
		fmt.Println("nothing to apply (already up-to-date)")
		return nil
	}
	for _, v := range ran {
		fmt.Printf("applied %d\n", v)
	}
	return nil
}

func cmdDown(db *gorm.DB, args []string) error {
	fs := flag.NewFlagSet("down", flag.ExitOnError)
	_ = fs.Parse(args)
	n := 1
	if fs.NArg() > 0 {
		v, err := strconv.Atoi(fs.Arg(0))
		if err != nil {
			return fmt.Errorf("invalid n: %w", err)
		}
		n = v
	}
	rolled, err := migrate.Down(db, n)
	if err != nil {
		return err
	}
	if len(rolled) == 0 {
		fmt.Println("nothing to rollback")
		return nil
	}
	for _, v := range rolled {
		fmt.Printf("rolled back %d\n", v)
	}
	return nil
}

func cmdStatus(db *gorm.DB, _ []string) error {
	rows, err := migrate.Status(db)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VERSION\tNAME\tSTATUS\tAPPLIED_AT")
	for _, r := range rows {
		status, applied := "pending", ""
		if r.Applied {
			status = "applied"
			applied = r.AppliedAt.Format(time.RFC3339)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", r.Version, r.Name, status, applied)
	}
	return w.Flush()
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
