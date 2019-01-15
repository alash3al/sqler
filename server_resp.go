package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/redcon"
)

func initRESPServer() error {
	return redcon.ListenAndServe(
		*flagRESPListenAddr,
		func(conn redcon.Conn, cmd redcon.Command) {
			// handles any panic
			defer (func() {
				if err := recover(); err != nil {
					conn.WriteError(fmt.Sprintf("fatal error: %s", (err.(error)).Error()))
				}
			})()

			// normalize the todo action "command"
			// normalize the command arguments
			todo := strings.TrimSpace(string(cmd.Args[0]))
			todoNormalized := strings.ToLower(todo)
			args := []string{}
			for _, v := range cmd.Args[1:] {
				v := strings.TrimSpace(string(v))
				args = append(args, v)
			}

			// internal command to pick a database
			if todoNormalized == "select" {
				conn.WriteString("OK")
				return
			}

			// internal ping-pong
			if todoNormalized == "ping" {
				conn.WriteString("PONG")
				return
			}

			// ECHO <args ...>
			if todoNormalized == "echo" {
				conn.WriteString(strings.Join(args, " "))
				return
			}

			// HELP|INFO|LIST
			if todoNormalized == "list" || todoNormalized == "help" || todoNormalized == "info" {
				conn.WriteArray(macrosManager.Size())
				for _, v := range macrosManager.List() {
					conn.WriteBulkString(v)
				}
				return
			}

			// close the connection
			if todoNormalized == "quit" {
				conn.WriteString("OK")
				conn.Close()
				return
			}

			macro := macrosManager.Get(todo)
			if nil == macro {
				conn.WriteError("not found")
				conn.Close()
				return
			}

			var input map[string]interface{}
			if len(args) > 0 {
				json.Unmarshal([]byte(args[0]), &input)
			}

			// handle our command
			commandExecMacro(conn, macro, input)
		},
		func(conn redcon.Conn) bool {
			conn.SetContext(map[string]interface{}{})
			return true
		},
		nil,
	)
}

// commandExecMacro - resp command handler
func commandExecMacro(conn redcon.Conn, macro *Macro, input map[string]interface{}) {
	out, err := macro.Call(input)
	if err != nil {
		conn.WriteArray(2)
		conn.WriteInt(0)

		j, _ := json.Marshal(err.Error())

		conn.WriteBulk(j)

		return
	}

	jsonOUT, _ := json.Marshal(out)

	conn.WriteArray(2)
	conn.WriteInt(1)
	conn.WriteBulk(jsonOUT)
}
