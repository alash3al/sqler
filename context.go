package main

// Context ...
type Context struct {
	Input map[string]interface{}
}

// NewContext - initialize a context
// ref: https://gist.github.com/siddontang/8875771
func NewContext() *Context {
	c := new(Context)
	c.Input = make(map[string]interface{})

	return c
}

// SQL - a sql escape function
func (c *Context) SQL(s interface{}) interface{} {
	if s == nil {
		return ""
	}

	sql, ok := s.(string)
	if !ok {
		return s
	}

	dest := []byte{}
	var escape byte
	for i := 0; i < len(sql); i++ {
		c := sql[i]

		escape = 0

		switch c {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
			break
		case '\n': /* Must be escaped for logs */
			escape = 'n'
			break
		case '\r':
			escape = 'r'
			break
		case '\\':
			escape = '\\'
			break
		case '\'':
			escape = '\''
			break
		case '"': /* Better safe than sorry */
			escape = '"'
			break
		case '\032': /* This gives problems on Win32 */
			escape = 'Z'
		}

		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, c)
		}
	}

	return interface{}(string(dest))
}
