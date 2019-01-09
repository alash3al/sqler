package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/dop251/goja"
	"github.com/jmoiron/sqlx"
)

// Macro - a macro configuration
type Macro struct {
	Authorizers []string            `json:"authorizers"`
	Methods     []string            `json:"method"`
	Rules       map[string][]string `json:"rules"`
	Exec        string              `json:"exec"`
	Transformer string              `json:"transformer"`
	name        string
	compiled    *template.Template
}

// Call - executes the macro
func (m *Macro) Call(input map[string]interface{}) (interface{}, error) {
	ctx := NewContext()
	ctx.SQLArgs = make(map[string]interface{})
	ctx.Input = input

	errs := Validate(input, m.Rules)
	if len(errs) > 0 {
		return errs, errors.New("validation errors")
	}

	src, err := m.compileMacro(ctx)
	if err != nil {
		return err.Error(), err
	}

	out, err := m.execSQLQuery(strings.Split(src, ";"), ctx.SQLArgs)
	if err != nil {
		return err.Error(), err
	}

	return m.execTransformer(out, m.Transformer)
}

// compileMacro - compile the specified macro and pass the specified ctx
func (m *Macro) compileMacro(ctx *Context) (string, error) {
	if m.compiled.Lookup(m.name) == nil {
		return "resource not found", errors.New("resource not found")
	}

	var buf bytes.Buffer

	rw := io.ReadWriter(&buf)
	if err := m.compiled.ExecuteTemplate(rw, m.name, ctx); err != nil {
		return "", err
	}

	src, err := ioutil.ReadAll(rw)
	if err != nil {
		return "", err
	}

	if len(src) < 1 {
		return "", errors.New("empty resource")
	}

	return strings.Trim(strings.TrimSpace(string(src)), ";"), nil
}

// execSQLQuery - execute the specified sql query
func (m *Macro) execSQLQuery(sqls []string, args map[string]interface{}) (interface{}, error) {
	conn, err := sqlx.Open(*flagDBDriver, *flagDBDSN)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	for _, sql := range sqls[0 : len(sqls)-1] {
		sql = strings.TrimSpace(sql)
		if "" == sql {
			continue
		}
		if _, err := conn.NamedExec(sql, args); err != nil {
			fmt.Println("....")
			return nil, err
		}
	}

	rows, err := conn.NamedQuery(sqls[len(sqls)-1], args)
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
func (m *Macro) scanSQLRow(rows *sqlx.Rows) (map[string]interface{}, error) {
	row := make(map[string]interface{})
	if err := rows.MapScan(row); err != nil {
		return nil, err
	}

	for k, v := range row {
		if nil == v {
			continue
		}

		switch v.(type) {
		case []uint8:
			v = []byte(v.([]uint8))
		default:
			v, _ = json.Marshal(v)
		}

		var d interface{}
		if nil == json.Unmarshal(v.([]byte), &d) {
			row[k] = d
		} else {
			row[k] = string(v.([]byte))
		}
	}

	return row, nil
}

// execTransformer - run the transformer function
func (m *Macro) execTransformer(data interface{}, transformer string) (interface{}, error) {
	vm := goja.New()

	vm.Set("$result", data)

	v, err := vm.RunString(transformer)
	if err != nil {
		return nil, err
	}

	return v, nil
}
