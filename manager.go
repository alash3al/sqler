// Copyright 2018 The SQLer Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/alash3al/go-color"
	"github.com/hashicorp/hcl"
	"github.com/robfig/cron/v3"
)

// Manager - a macros manager
type Manager struct {
	macros   map[string]*Macro
	compiled *template.Template
	cron     *cron.Cron
	sync.RWMutex
}

// NewManager - initialize a new manager
func NewManager(configpath string) (*Manager, error) {
	manager := new(Manager)
	manager.macros = make(map[string]*Macro)
	manager.compiled = template.New("main")
	manager.cron = cron.New()

	for _, p := range strings.Split(configpath, ",") {
		files, _ := filepath.Glob(p)

		if len(files) < 1 {
			return nil, fmt.Errorf("invalid path (%s)", p)
		}

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
				manager.macros[k] = v
				_, err := manager.compiled.New(k).Parse(v.Exec)
				if err != nil {
					return nil, err
				}
				v.manager = manager
				v.name = k
				v.Trigger.Webhook = strings.TrimSpace(v.Trigger.Webhook)
				v.Trigger.Macro = strings.TrimSpace(v.Trigger.Macro)

				if strings.TrimSpace(v.Cron) != "" {
					(func(v *Macro) {
						_, err := manager.cron.AddFunc(v.Cron, func() {
							fmt.Println(color.YellowString("=> Executing cron " + v.name))
							if _, err := v.Call(map[string]interface{}{}); err != nil {
								fmt.Println(color.RedString("=> Faild executing cron " + v.name + " due to an error: " + err.Error()))
							} else {
								fmt.Println(color.GreenString("=> Executing cron " + v.name + " succeeded!"))
							}
						})

						if err != nil {
							fmt.Println(color.RedString(err.Error()))
							os.Exit(1)
						}
					})(v)
				}
			}
		}
	}

	manager.cron.Start()

	return manager, nil
}

// Get - fetches the required macro
func (m *Manager) Get(macro string) *Macro {
	m.RLock()
	defer m.RUnlock()

	return m.macros[macro]
}

// Size - return the size of the currently loaded configs
func (m *Manager) Size() int {
	return len(m.macros)
}

// List - return a list of registered macros
func (m *Manager) List() (ret []string) {
	for k := range m.macros {
		ret = append(ret, k)
	}

	return ret
}
