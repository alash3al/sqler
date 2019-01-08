package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/hcl"
	"github.com/jmoiron/sqlx"
)

// Macro - a macro configuration
type Macro struct {
	Methods []string            `json:"method"`
	Rules   map[string][]string `json:"rules"`
	Exec    string              `json:"exec"`
}

// Manager - a macros manager
type Manager struct {
	configs map[string]*Macro
	macros  *template.Template
}

// NewManager - initialize a new manager
func NewManager(configpath string) (*Manager, error) {
	manager := new(Manager)
	manager.configs = make(map[string]*Macro)
	files, _ := filepath.Glob(configpath)

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var config map[string]*Macro
		if err := hcl.Unmarshal(data, &config); err != nil {
			return nil, err
		}

		for k, v := range config {
			manager.configs[k] = v
		}
	}

	manager.macros = template.New("main")
	for k, v := range manager.configs {
		_, err := manager.macros.New(k).Parse(v.Exec)
		if err != nil {
			return nil, err
		}
	}

	return manager, nil
}

// Call - call the specified macro
func (m *Manager) Call(macro string, input map[string]interface{}) (interface{}, error) {
	ctx := NewContext()
	ctx.Input = input

	src, err := m.compileMacro(macro, ctx)
	if err != nil {
		return nil, err
	}

	return m.execSQLQuery(src)
}

// Get - fetches the required macro
func (m *Manager) Get(macro string) *Macro {
	return m.configs[macro]
}

// compileMacro - compile the specified macro and pass the specified ctx
func (m *Manager) compileMacro(macro string, ctx *Context) (string, error) {
	if m.macros.Lookup(macro) == nil {
		return "", errors.New("resource not found #1")
	}

	var buf bytes.Buffer
	rw := io.ReadWriter(&buf)
	if err := m.macros.ExecuteTemplate(rw, macro, ctx); err != nil {
		return "", err
	}

	src, err := ioutil.ReadAll(rw)
	if err != nil {
		return "", err
	}

	if len(src) < 1 {
		return "", errors.New("resource not found #2")
	}

	return string(src), nil
}

// execSQLQuery - execute the specified sql query
func (m *Manager) execSQLQuery(sql string) (interface{}, error) {
	conn, err := sqlx.Open(*flagDBDriver, *flagDBDSN)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Queryx(string(sql))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []map[string]interface{}{}

	for rows.Next() {
		row, err := m.scanSQLRow(rows)
		if err != nil {
			continue
		}
		ret = append(ret, row)
	}

	return interface{}(ret), nil
}

// scanSQLRow - scan a row from the specified rows
func (m *Manager) scanSQLRow(rows *sqlx.Rows) (map[string]interface{}, error) {
	row := make(map[string]interface{})
	if err := rows.MapScan(row); err != nil {
		return nil, err
	}

	for k, v := range row {
		if nil == v {
			continue
		}
		v = []byte(v.([]uint8))
		var d interface{}
		if nil == json.Unmarshal(v.([]byte), &d) {
			row[k] = d
		} else {
			row[k] = string(v.([]byte))
		}
	}

	return row, nil
}
