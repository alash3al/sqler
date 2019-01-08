// Copyright 2018 The SQLer Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Context ...
type Context struct {
	SQLArgs map[string]interface{}
	Input   map[string]interface{}
}

// NewContext - initialize a context
// ref: https://gist.github.com/siddontang/8875771
func NewContext() *Context {
	c := new(Context)
	c.Input = make(map[string]interface{})

	return c
}

// Hash - hash the specified input using the specified method [md5, sha1, sha256, sha512]
func (c Context) Hash(method string, input string) string {
	result := ""

	switch strings.ToLower(method) {
	case "md5":
		hash := md5.Sum([]byte(input))
		result = hex.EncodeToString(hash[:])
	case "sha1":
		hash := sha1.Sum([]byte(input))
		result = hex.EncodeToString(hash[:])
	case "sha256":
		hash := sha256.Sum256([]byte(input))
		result = hex.EncodeToString(hash[:])
	case "sha512":
		hash := sha512.Sum512([]byte(input))
		result = hex.EncodeToString(hash[:])
	case "bcrypt":
		hash, err := bcrypt.GenerateFromPassword([]byte(input), bcrypt.DefaultCost)
		if err == nil {
			result = string(hash)
		}
	}

	return result
}

// UnixTime - returns the unix time in seconds
func (c Context) UnixTime() int64 {
	return time.Now().Unix()
}

// UnixNanoTime - returns the unix time in nano seconds
func (c Context) UnixNanoTime() int64 {
	return time.Now().UnixNano()
}

// Uniqid - returns a unique string
func (c Context) Uniqid() string {
	return snow.Generate().String()
}

func (c *Context) BindVar(name string, value interface{}) string {
	c.SQLArgs[name] = value
	return ""
}
