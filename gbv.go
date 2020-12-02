package main

import (
	"flag"
	"log"
	"os"

	"github.com/jroimartin/gocui"
)

var metricsAddress = flag.String("a", "localhost:8080", "metrics address")
var metricsPath = flag.String("p", "/actuator/metrics", "metrics path")
var debug = flag.Bool("vv", false, "verbose output")
var eb = newEventBus()


func main() {
	flag.Parse()
	g, err := gocui.NewGui(gocui.OutputNormal)
	client := NewClient(*metricsAddress, *metricsPath)
	// vu := newViewUpdater(g)
	// eb.regist(vu)

	log.SetFlags(log.Llongfile)
	if *debug {
		f, err := os.OpenFile(".log", os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
		log.SetOutput(f)
	}
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true
	sideBar := newSideBar(g, client, eb)
	detail := newDetail(g, client, eb)
	sideBar.init()
	g.SetManager(sideBar, detail)

	if err := keybindings(g, sideBar); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui,s *SideBar) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowDown, gocui.ModNone, s.curDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowUp, gocui.ModNone, curUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyEnter, gocui.ModNone, s.onKeyEnter); err != nil {
		return err
	}

	return nil
}
