package kabarda

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

func (k *Kabarda) MigrateUp(dsn string) error {
	mgr, err := migrate.New("file://"+k.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer mgr.Close()
	if err := mgr.Up(); err != nil {
		log.Println("error run migration")
		return err
	}
	return nil
}

func (k *Kabarda) MigrateDownAll(dsn string) error {
	mgr, err := migrate.New("file://"+k.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer mgr.Close()
	if err := mgr.Down(); err != nil {
		return err
	}
	return nil

}

func (k *Kabarda) MigrateSteps(n int, dsn string) error {
	mgr, err := migrate.New("file://"+k.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer mgr.Close()
	if err := mgr.Steps(n); err != nil {
		return err
	}
	return nil
}

func (k *Kabarda) MigrateForce(dsn string) error {
	mgr, err := migrate.New("file://"+k.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer mgr.Close()
	if err := mgr.Force(-1); err != nil {
		return err
	}
	return nil
}
