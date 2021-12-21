package main

import (
	"errors"
	"io/ioutil"
	"os"
	"unicode/utf8"

	lua "github.com/Shopify/go-lua"
)

func rtQuitEditor(l *lua.State) int {
	screen.Fini()
	os.Exit(0)
	return 0
}

func rtFatal(l *lua.State) int {
	fatal(errors.New(lua.CheckString(l, 1)))
	return 0
}

func rtPadLeft(l *lua.State) int {
	str := lua.CheckString(l, 1)
	length := lua.CheckInteger(l, 2)
	padding := lua.CheckString(l, 3)[0]
	for utf8.RuneCountInString(str) < length {
		str = string(padding) + str
	}
	l.PushString(str)
	return 1
}

func rtPadRight(l *lua.State) int {
	str := lua.CheckString(l, 1)
	length := lua.CheckInteger(l, 2)
	padding := lua.CheckString(l, 3)[0]
	for utf8.RuneCountInString(str) < length {
		str = str + string(padding)
	}
	l.PushString(str)
	return 1
}

func rtFileReadAll(l *lua.State) int {
	path := lua.CheckString(l, 1)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		l.PushString("Error reading file \"" + path + "\"")
		l.Error()
		panic("unreachable")
	}
	l.PushString(string(contents))
	return 1
}

func rtFileWriteAll(l *lua.State) int {
	path := lua.CheckString(l, 1)
	contents := lua.CheckString(l, 2)
	err := ioutil.WriteFile(path, []byte(contents), 0666)
	if err != nil {
		l.PushString("Error writing file \"" + path + "\"")
		l.Error()
		panic("unreachable")
	}
	return 0
}
