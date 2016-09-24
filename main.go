package main

import (
	"bufio"
	"fmt"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"os"
	"os/exec"
	"strings"
)

type Target struct {
	Host      string
	Protocol  string
	IpAddress string
}

func (t Target) remote_login() {
	fmt.Printf("Connect to %s(%s).", t.Host, t.IpAddress)
	cmd := exec.Command(t.Protocol, t.IpAddress)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	defer os.Exit(0)
}

func read_file(path string) (lines []string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}

	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if s.Err() != nil {
		return []string{}, err
	}
	return lines, nil
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}

func redraw_all(target_slice []Target, hl_line int) {
	const coldef = termbox.ColorDefault
	const SelectBgColor = 8
	bgcolor := coldef
	termbox.Clear(coldef, coldef)
	tbprint(0, 0, coldef, bgcolor, "fumu")
	tbprint(0, 1, coldef, bgcolor, "[Enter]:login\t[j,Ctrl-N]:down\t[k,Ctrl-P]:up\t[ESC]:quit")

	for i, t := range target_slice {
		if hl_line == i {
			bgcolor = SelectBgColor
		} else {
			bgcolor = coldef
		}
		msg := t.Host + "\t(" + t.IpAddress + ")"
		tbprint(0, i+3, coldef, bgcolor, msg)
	}
	termbox.Flush()
}

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Exec like..\n\t./step-go your.conf")
		os.Exit(0)
	}
	fp := os.Args[1]
	lines, err := read_file(fp)
	if err != nil {
		panic(err)
	}

	var target_slice []Target
	for _, line := range lines {
		elements := strings.Fields(line)
		if len(elements) != 3 {
			fmt.Println("configration file is invalid\n")
			os.Exit(0)
		}
		target_slice = append(target_slice, Target{Host: elements[0], Protocol: elements[1], IpAddress: elements[2]})
	}

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	const KeyJ = 106
	const KeyK = 107
	var hl_line int
	max_hl_line := len(lines) - 1

	redraw_all(target_slice, hl_line)

mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			case termbox.KeyCtrlP:
				hl_line = hl_line - 1
			case termbox.KeyCtrlN:
				hl_line = hl_line + 1
			case termbox.KeyEnter:
				termbox.Close()
				target_slice[hl_line].remote_login()
			default:
				if ev.Ch == KeyK {
					hl_line = hl_line - 1
				} else if ev.Ch == KeyJ {
					hl_line = hl_line + 1
				} else {
					// do nothing
					// fmt.Println(ev.Ch)
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
		if hl_line < 0 {
			hl_line = 0
		} else if hl_line > max_hl_line {
			hl_line = max_hl_line
		}
		redraw_all(target_slice, hl_line)
	}
}
