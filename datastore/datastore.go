package datastore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pygrum/karmine/config"
)

type Kdb struct {
	DB       *sql.DB
	Database string
}

func New(database string) (*Kdb, error) {
	username, password, err := config.GetSqlCreds()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("mysql", username+":"+password+"@/"+database)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(20)
	return &Kdb{
		DB:       db,
		Database: database,
	}, nil
}

func (db *Kdb) UUIDExists(UUID string) bool {
	row := db.DB.QueryRow("SELECT COUNT(1) FROM karmine WHERE uuid = ?", UUID)
	var temp string
	row.Scan(&temp)
	return len(temp) != 0
}

func (db *Kdb) GetKeysByUUID(UUID string) (string, string, string) {
	row := db.DB.QueryRow("SELECT aeskey, xorkey1, xorkey2 FROM karmine WHERE uuid = ?", UUID)
	var aeskey, x1, x2 string
	row.Scan(&aeskey, &x1, &x2)
	return aeskey, x1, x2
}

func (db *Kdb) AddCmdToStack(uuid string, cmd string) error {
	query := "INSERT INTO kmdstack(staged_cmd, uuid) VALUES (?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error when preparing SQL statement: %v", err)
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, cmd, uuid)
	if err != nil {
		return fmt.Errorf("error when inserting row into products table: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error when finding rows affected: %v", err)
	}
	if rows < 1 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (db *Kdb) DeleteCmdByID(id int) error {
	query := "DELETE FROM kmdstack WHERE cmd_id = ?"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error when preparing SQL statement: %v", err)
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("error when deleting row from table: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error when finding rows affected: %v", err)
	}
	if rows < 1 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (db *Kdb) GetCmdByID(ID int) (cmd string) {
	row := db.DB.QueryRow("SELECT staged_cmd FROM kmdstack WHERE cmd_id = ?", ID)
	row.Scan(&cmd)
	return
}

func (db *Kdb) GetAddrByUUID(UUID string) string {
	row := db.DB.QueryRow("SELECT rhostname FROM karmine WHERE uuid = ?", UUID)
	var rhost string
	row.Scan(&rhost)
	return rhost
}

func (db *Kdb) CreateNewInstance(UUID, aesKey, x1, x2 string) error {
	query := "INSERT INTO karmine(uuid, aeskey, xorkey1, xorkey2) VALUES (?, ?, ?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error when preparing SQL statement: %v", err)
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, UUID, aesKey, x1, x2)
	if err != nil {
		return fmt.Errorf("error when inserting row into table: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error when finding rows affected: %v", err)
	}
	if rows < 1 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (db *Kdb) GetCmdID() int {
	r := db.DB.QueryRow("SELECT max(cmd_id) FROM kmdstack")
	var id int
	if err := r.Scan(&id); err != nil {
		id = -1
	}
	return id
}

func GetUUIDByName(name string) (string, error) {
	conf, err := config.GetFullConfig()
	if err != nil {
		return "", err
	}
	for _, h := range conf.Hosts {
		if h.Name == name {
			return h.UUID, nil
		}
	}
	return "", fmt.Errorf("name not found in config")
}

func GetNameByUUID(uuid string) (string, error) {
	conf, err := config.GetFullConfig()
	if err != nil {
		return "", err
	}
	for _, h := range conf.Hosts {
		if h.UUID == uuid {
			return h.Name, nil
		}
	}
	return "", fmt.Errorf("uuid not found in config")
}

func AddProfile(uuid, name, strain string) error {
	conf, err := config.GetFullConfig()
	if err != nil {
		return err
	}
	conf.Hosts = append(conf.Hosts, config.Profile{
		UUID:   uuid,
		Name:   name,
		Strain: strain,
	})
	b, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.ConfigPath(), b, 0644)
}

func (db *Kdb) RemoveProfile(name string) error {
	conf, err := config.GetFullConfig()
	if err != nil {
		return err
	}
	var newHosts []config.Profile
	var uuid string
	for _, h := range conf.Hosts {
		if h.Name != name {
			newHosts = append(newHosts, h)
		} else {
			uuid = h.UUID
		}
	}
	conf.Hosts = newHosts
	b, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return err
	}
	if len(uuid) == 0 {
		return fmt.Errorf("could not find uuid for %s", name)
	}
	query := "DELETE FROM karmine WHERE uuid = ?"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error when preparing SQL statement: %v", err)
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, uuid)
	if err != nil {
		return fmt.Errorf("error when deleting row from table: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error when finding rows affected: %v", err)
	}
	if rows < 1 {
		return fmt.Errorf("no rows affected")
	}
	return os.WriteFile(config.ConfigPath(), b, 0644)
}

func ShowStage(stage string) (string, error) {
	const filestage, cmdstage = "/tmp/filestage.tmp", "/tmp/cmdstage.tmp"
	conf, err := config.GetFullConfig()
	if err != nil {
		return "", err
	}
	if stage == filestage {
		if len(conf.Stages.File) == 0 {
			if _, err = os.Stat(filestage); err == nil {
				return "nothing staged in configuration, but stagefile exists. delete it", nil
			}
			return "nothing staged", nil
		} else {
			return fmt.Sprintf("%s is staged", conf.Stages.File), nil
		}
	} else {
		if len(conf.Stages.Cmd) == 0 {
			if _, err = os.Stat(cmdstage); err == nil {
				return "nothing staged in configuration, but stagefile exists. delete it", nil
			}
			return "nothing staged", nil
		} else {
			return conf.Stages.Cmd, nil
		}
	}
}

func ClearStage(stage string) error {
	const filestage = "/tmp/filestage.tmp"
	conf, err := config.GetFullConfig()
	if err != nil {
		return err
	}
	if stage == filestage {
		conf.Stages.File = ""
	} else {
		conf.Stages.Cmd = ""
	}
	bytes, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.ConfigPath(), bytes, 0644)
}
