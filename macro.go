package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alash3al/go-color"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
)

// Macro - a macro configuration
type Macro struct {
	Methods     []string
	Include     []string
	Validators  map[string]string
	Authorizer  string
	Bind        map[string]string
	Exec        string
	Aggregate   []string
	Transformer string

	Trigger struct {
		Webhook string
		Macro   string
	}

	Cron string

	name    string
	manager *Manager
}

// Call - executes the macro
func (m *Macro) Call(input map[string]interface{}) (interface{}, error) {
	ok, err := m.authorize(input)
	if err != nil {
		return err.Error(), err
	}

	if !ok {
		return errAuthorizationError.Error(), errAuthorizationError
	}

	invalid, err := m.validate(input)
	if err != nil {
		return err.Error(), err
	} else if len(invalid) > 0 {
		return invalid, errValidationError
	}

	if err := m.runIncludes(input); err != nil {
		return err.Error(), err
	}

	var out interface{}

	if len(m.Aggregate) > 0 {
		out, err = m.aggregate(input)
		if err != nil {
			return err.Error(), err
		}
	} else {
		out, err = m.execSQLQuery(strings.Split(strings.TrimSpace(m.Exec), *flagSQLSeparator), input)
		if err != nil {
			return err.Error(), err
		}
	}

	out, err = m.transform(out)
	if err != nil {
		return err.Error(), err
	}

	go (func() {
		if m.Trigger.Webhook != "" {
			_, err := resty.New().R().SetDoNotParseResponse(true).SetHeader("Content-Type", "application/json").SetBody(map[string]interface{}{
				"payload": out,
			}).Post(m.Trigger.Webhook)

			if err != nil {
				color.Red("[X]- " + err.Error())
			}
		}

		if subm := m.manager.Get(m.Trigger.Macro); subm != nil {
			_, err := subm.Call(map[string]interface{}{
				"payload": out,
			})

			if err != nil {
				color.Red("[X]- " + err.Error())
			}
		}
	})()

	return out, nil
}

// execSQLQuery - execute the specified sql query
func (m *Macro) execSQLQuery(sqls []string, input map[string]interface{}) (interface{}, error) {
	args, err := m.buildBind(input)
	if err != nil {
		return nil, err
	}

	conn, err := sqlx.Open(*flagDBDriver, *flagDBDSN)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	for i, sql := range sqls {
		if strings.TrimSpace(sql) == "" {
			sqls = append(sqls[0:i], sqls[i+1:]...)
		}
	}

	for _, sql := range sqls[0 : len(sqls)-1] {
		sql = strings.TrimSpace(sql)
		if "" == sql {
			continue
		}
		if _, err := conn.NamedExec(sql, args); err != nil {
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

// transform - run the transformer function
func (m *Macro) transform(data interface{}) (interface{}, error) {
	transformer := strings.TrimSpace(m.Transformer)
	if transformer == "" {
		return data, nil
	}

	vm := initJSVM(map[string]interface{}{"$result": data})

	v, err := vm.RunString(transformer)
	if err != nil {
		return nil, err
	}

	return v.Export(), nil
}

// authorize - run the authorizer function
func (m *Macro) authorize(input map[string]interface{}) (bool, error) {
	authorizer := strings.TrimSpace(m.Authorizer)
	if authorizer == "" {
		return true, nil
	}

	var execError error

	vm := initJSVM(map[string]interface{}{"$input": input})

	val, err := vm.RunString(m.Authorizer)
	if err != nil {
		return false, err
	}

	if execError != nil {
		return false, execError
	}

	return val.ToBoolean(), nil
}

// aggregate - run the aggregators
func (m *Macro) aggregate(input map[string]interface{}) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	for _, k := range m.Aggregate {
		macro := m.manager.Get(k)
		if nil == macro {
			err := fmt.Errorf("unknown macro %s", k)
			return nil, err
		}
		out, err := macro.Call(input)
		if err != nil {
			return nil, err
		}
		ret[k] = out
	}
	return ret, nil
}

// validate - validate the input aginst the rules
func (m *Macro) validate(input map[string]interface{}) (ret []string, err error) {
	if len(m.Validators) < 1 {
		return nil, nil
	}

	vm := initJSVM(map[string]interface{}{"$input": input})

	for k, src := range m.Validators {
		val, err := vm.RunString(src)
		if err != nil {
			return nil, err
		}

		if !val.ToBoolean() {
			ret = append(ret, k)
		}
	}

	return ret, err
}

// buildBind - build the bind vars
func (m *Macro) buildBind(input map[string]interface{}) (map[string]interface{}, error) {
	if len(m.Bind) < 1 {
		return nil, nil
	}

	vm := initJSVM(map[string]interface{}{"$input": input})
	ret := map[string]interface{}{}

	for k, src := range m.Bind {
		val, err := vm.RunString(src)
		if err != nil {
			return nil, err
		}

		ret[k] = val.Export()
	}

	return ret, nil
}

// runIncludes - run the include function
func (m *Macro) runIncludes(input map[string]interface{}) error {
	for _, name := range m.Include {
		macro := m.manager.Get(name)
		if nil == macro {
			return fmt.Errorf("macro %s not found", name)
		}
		_, err := macro.Call(input)
		if err != nil {
			return err
		}
	}
	return nil
}
