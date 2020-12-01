package main

import (
	// "context"
	"flag"
	"log"
	"os"

	"github.com/jroimartin/gocui"
)

var metricsAddress = flag.String("a", "localhost:8080", "metrics address")
var metricsPath = flag.String("p", "/actuator/metrics", "metrics path")
var debug = flag.Bool("vv", false, "verbose output")

// type viewUpdater struct {
// 	cancelFunc context.CancelFunc

// 	q <-chan int8
// }

// func newViewUpdater(g *gocui.Gui) *viewUpdater {
// 	vu := &viewUpdater{}
// 	ctx,cancelFunc := context.WithCancel(context.Background())
// 	vu.cancelFunc = cancelFunc
// 	vu.q = make(<-chan int8,0)
// 	go func() {
// 		for {
// 			select {
// 			case <- ctx.Done():
// 				return
// 			case <- vu.q:
// 				g.UP
// 			}
// 		}
// 	}

// }

// func (vu *viewUpdater) subscribe(evt event) {

// }

func main() {
	flag.Parse()
	client := NewClient(*metricsAddress, *metricsPath)
	eventBus := newEventBus()
	g, err := gocui.NewGui(gocui.OutputNormal)
	if *debug {
		f, err := os.OpenFile(".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
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

	meternames, err := client.Names()
	if err != nil {
		log.Panicln(err)
	}
	g.SetManagerFunc(layoutFunc(client, eventBus, meternames))
	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
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

func layoutFunc(client *Client, eventBus eventBus, meterNames *MeterNames) func(g *gocui.Gui) error {
	meterNamesBottomX := meterNamesMaxLen(meterNames.Names)
	if meterNamesBottomX == 0 {
		meterNamesBottomX = 30
	}

	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		if v, err := g.SetView("side", 1, 1, meterNamesBottomX, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			sideBar := newSideBar(v, eventBus, meterNames.Names)
			sideBar.init(g)
			if _, err := g.SetCurrentView("side"); err != nil {
				return err
			}
		}
		if v, err := g.SetView("summary", meterNamesBottomX, 1, maxX-1, 10); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			_ = newSummary(v, nil, client, eventBus)

		}
		return nil
	}
}
