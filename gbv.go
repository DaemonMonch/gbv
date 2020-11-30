package main

import (
	"flag"
	"log"

	"github.com/jroimartin/gocui"
)

var metricsAddress = flag.String("a", "localhost:8080", "metrics address")
var metricsPath = flag.String("p", "/actuator/metrics", "metrics path")
var debug = flag.Bool("vv", false, "verbose output")

func main() {
	flag.Parse()
	client := NewClient(*metricsAddress, *metricsPath)
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true

	meternames, err := client.Names()
	if err != nil {
		log.Panicln(err)
	}
	g.SetManagerFunc(layoutFunc(meternames))
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

func layoutFunc(meterNames *MeterNames) func(g *gocui.Gui) error {
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
			sideBar := newSideBar(v, meterNames.Names)
			sideBar.init(g)
			if _, err := g.SetCurrentView("side"); err != nil {
				return err
			}
		}
		_ = maxX
		// if v, err := g.SetView("main", 30, -1, maxX, maxY); err != nil {
		//         if err != gocui.ErrUnknownView {
		//                 return err
		//         }
		//         b, err := ioutil.ReadFile("Mark.Twain-Tom.Sawyer.txt")
		//         if err != nil {
		//                 panic(err)
		//         }
		//         fmt.Fprintf(v, "%s", b)
		//         v.Editable = true
		//         v.Wrap = true
		//         if _, err := g.SetCurrentView("main"); err != nil {
		//                 return err
		//         }
		// }
		return nil
	}
}
