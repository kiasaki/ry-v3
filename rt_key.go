package main

import lua "github.com/Shopify/go-lua"

func rtKey(l *lua.State) int {
	representation := lua.CheckString(l, 1)
	l.PushUserData(NewKey(representation))
	return 1
}

func rtKeyStr(l *lua.State) int {
	key := l.ToValue(1).(*Key)
	l.PushString(key.String())
	return 1
}

func rtKeyLen(l *lua.State) int {
	key := l.ToValue(1).(*Key)
	l.PushInteger(key.Length())
	return 1
}

func rtKeyAppend(l *lua.State) int {
	key := l.ToValue(1).(*Key)
	key2 := l.ToValue(2).(*Key)
	key.AppendKey(key2)
	l.PushUserData(key)
	return 1
}

func rtKeyMatches(l *lua.State) int {
	key := l.ToValue(1).(*Key)
	key2 := l.ToValue(2).(*Key)
	l.PushBoolean(key.Matches(key2))
	return 1
}

func rtKeyMatchesPart(l *lua.State) int {
	key := l.ToValue(1).(*Key)
	key2 := l.ToValue(2).(*Key)
	l.PushBoolean(key.MatchesPart(key2))
	return 1
}
