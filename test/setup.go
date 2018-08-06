package test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Setup struct {
	conn *pop.Connection
	args []string
}

func NewSetup(args []string) (Setup, error) {
	conn, err := pop.Connect("test")
	if err != nil {
		return Setup{}, err
	}

	return Setup{
		conn: conn,
		args: args,
	}, nil
}

func (s Setup) Run() error {
	if s.skipDBSetup() {
		s.args = removeFlag("--skip-db-setup", s.args)
		return nil
	}

	logrus.Info("Setting up database")
	s.resetDatabase()

	if s.forceMigrations() {
		logrus.Info("Running migrations")
		fm, err := pop.NewFileMigrator("./migrations", s.conn)

		if err != nil {
			return err
		}

		if err := fm.Up(); err != nil {
			return err
		}
	} else if schema := s.findSchemaFile(); schema != nil {
		logrus.Info("Loading schema file")
		err := s.conn.Dialect.LoadSchema(schema)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (s Setup) skipDBSetup() bool {
	logrus.Info(strings.Join(s.args, ""))
	return strings.Contains(strings.Join(s.args, ""), "--skip-db-setup")
}

func (s Setup) forceMigrations() bool {
	return strings.Contains(strings.Join(s.args, ""), "--force-migrations")
}

func (s Setup) findSchemaFile() io.Reader {
	if f, err := os.Open(filepath.Join("migrations", "schema.sql")); err == nil {
		return f
	}
	if dev, err := pop.Connect("development"); err == nil {
		schema := &bytes.Buffer{}
		if err = dev.Dialect.DumpSchema(schema); err == nil {
			return schema
		}
	}

	if test, err := pop.Connect("test"); err == nil {
		fm, err := pop.NewFileMigrator("./migrations", test)
		if err != nil {
			return nil
		}

		if err := fm.Up(); err == nil {
			if f, err := os.Open(filepath.Join("migrations", "schema.sql")); err == nil {
				return f
			}
		}
	}
	return nil
}

func (s Setup) resetDatabase() error {
	conn, err := pop.Connect("test")
	if err != nil {
		return err
	}

	err = conn.Dialect.DropDB()
	if err != nil {
		return err
	}

	return conn.Dialect.CreateDB()
}
