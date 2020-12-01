package main

import (
	"log"
	"fmt"
	"github.com/jroimartin/gocui"
)

type SideBar struct {
	view       *gocui.View
	meterNames []string

	eventBus eventBus
}

func newSideBar(view *gocui.View, eventBus eventBus ,meterNames []string) *SideBar {
	
	sideBar := &SideBar{view, meterNames,eventBus}
	eventBus.regist(sideBar)
	return sideBar
}

func (sideBar *SideBar) init(g *gocui.Gui) error {
	v := sideBar.view
	v.Title = "meters"
	v.Highlight = true
	v.SelBgColor = gocui.ColorGreen
	v.SelFgColor = gocui.ColorBlack
	for _, name := range sideBar.meterNames {
		fmt.Fprintln(v, name)
	}

	if err := g.SetKeybinding(sideBar.view.Name(), gocui.KeyArrowDown, gocui.ModNone, sideBar.curDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(sideBar.view.Name(), gocui.KeyArrowUp, gocui.ModNone, curUp); err != nil {
		return err
	}
	if err := g.SetKeybinding(sideBar.view.Name(), gocui.KeyEnter, gocui.ModNone, sideBar.onKeyEnter); err != nil {
		return err
	}

	return nil
}

func (s *SideBar) onKeyEnter(g *gocui.Gui, v *gocui.View) error {
	l := getLine(g,v)
	if l != "" {
		log.Println(l)
		s.eventBus.pub(event{e:l,t:METER_NAME_CHANGED})
	}
	return nil
}

func (s *SideBar) curDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		l := getLine(g,v)
		if l == s.meterNames[len(s.meterNames) - 1] {
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

func  getLine(g *gocui.Gui, v *gocui.View) string {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	return l
}

func  (s *SideBar) subscribe(evt event) {

}
