package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

type SideBar struct {
	meterNames []string

	g        *gocui.Gui
	eventBus eventBus
	client   *Client
}

func newSideBar(g *gocui.Gui, client *Client, eventBus eventBus) *SideBar {

	sideBar := &SideBar{g: g, client: client, eventBus: eventBus}
	eventBus.regist(sideBar)
	return sideBar
}

func (s *SideBar) init() error {

	meternames, err := s.client.Names()
	s.meterNames = meternames.Names
	if err != nil {
		log.Panicln(err)
	}

	return nil
}

func meterNamesMaxLen(names []string) int {
	l := 0
	for _, name := range names {
		nameLen := len(name)
		if nameLen > l {
			l = nameLen
		}
	}
	return l
}

func (s *SideBar) onKeyEnter(g *gocui.Gui, v *gocui.View) error {
	log.Println("on key ender")
	l := getLine(g, v)
	if l != "" {
		s.eventBus.pub(event{e: l, t: METER_NAME_CHANGED})
	}
	return nil
}

func (s *SideBar) curDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		l := getLine(g, v)
		if l == s.meterNames[len(s.meterNames)-1] {
			return nil
		}
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func curUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func getLine(g *gocui.Gui, v *gocui.View) string {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	return l
}

func (s *SideBar) Layout(g *gocui.Gui) error {
	meterNamesBottomX := meterNamesMaxLen(s.meterNames)
	if meterNamesBottomX == 0 {
		meterNamesBottomX = 30
	}
	_, maxY := g.Size()
	if v, err := g.SetView("side", 1, 1, meterNamesBottomX, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "meters"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		for _, name := range s.meterNames {
			fmt.Fprintln(v, name)
		}
		if _, err := g.SetCurrentView("side"); err != nil {
			return err
		}
	}



	return nil
}

func (s *SideBar) subscribe(evt event) {

}
