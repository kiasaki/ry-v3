package main

import (
	"os"

	lua "github.com/Shopify/go-lua"
	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"
)

func rtScreenShow(l *lua.State) int {
	screen.Show()
	return 0
}

func rtScreenSync(l *lua.State) int {
	screen.Sync()
	return 0
}

func rtScreenClear(l *lua.State) int {
	screen.Clear()
	return 0
}

func rtScreenSize(l *lua.State) int {
	w, h := screen.Size()
	l.PushInteger(w)
	l.PushInteger(h)
	return 2
}

func rtScreenQuit(l *lua.State) int {
	closeScreen()
	return 0
}

func rtScreenNextKey(l *lua.State) int {
	ev := <-screenEvents
	switch ev := ev.(type) {
	case *tcell.EventKey:
		// TODO remove when we have higer confidence in runtime working
		if ev.Key() == tcell.KeyCtrlQ {
			screen.Fini()
			os.Exit(0)
		}
		pushKeyFromEvent(l, ev)
	default:
		l.PushNil()
	}
	return 1
}

func pushKeyFromEvent(l *lua.State, ev *tcell.EventKey) {
	ks := NewKeyStrokeFromKeyEvent(ev)
	k := NewKey("")
	k.AppendKeyStroke(ks)
	l.PushUserData(k)
}

func rtScreenWrite(l *lua.State) int {
	// style := lua.CheckUserData(l, 1, "style_mt").(tcell.Style)
	style := l.ToValue(1).(tcell.Style)
	x := lua.CheckInteger(l, 2)
	y := lua.CheckInteger(l, 3)
	str := lua.CheckString(l, 4)

	s := screen
	i := 0
	var deferred []rune
	dwidth := 0
	for _, r := range str {
		// Handle tabs
		if r == '\t' {
			// Print first tab char
			s.SetContent(x+i, y, '>', nil, style.Foreground(tcell.ColorAqua))
			i++

			// Add space till we reach tab column or tabWidth
			for j := 0; j < tabWidth-1 || i%tabWidth == 0; j++ {
				s.SetContent(x+i, y, ' ', nil, style)
				i++
			}

			deferred = nil
			continue
		}

		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}

	if len(deferred) != 0 {
		s.SetContent(x+i, y, deferred[0], deferred[1:], style)
		i += dwidth
	}

	// i is the real width of what we just outputed
	l.PushInteger(i)
	return 1
}
