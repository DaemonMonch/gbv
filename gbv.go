package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/jroimartin/gocui"
)

var metricsAddress = flag.String("a", "localhost:8080", "metrics address")
var metricsPath = flag.String("p", "/actuator/metrics", "metrics path")
var debug = flag.Bool("vv", false, "verbose output")
var eb = newEventBus()

type Refresher struct {
	lastMeterName string
	duration      int
	done          chan interface{}
	started       bool
	lastRequestDone bool
	eb            eventBus
}

func newRefresher(duration int, eb eventBus) *Refresher {
	r := &Refresher{}
	r.eb = eb
	r.done = make(chan interface{}, 1)
	r.duration = duration
	eb.regist(r)
	return r
}

func (r *Refresher) subscribe(evt event) {
	if evt.t == METER_NAME_CHANGED {
		r.lastMeterName = evt.e
	}

	if evt.t == CLIENT_REQUEST_DONE {
		r.lastRequestDone = true
	}
}

func (r *Refresher) start() {
	r.started = true
	go func() {
		ticker := time.NewTicker(time.Duration(r.duration) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-r.done:
				return
			case t := <-ticker.C:
				if r.started && r.lastMeterName != "" && r.lastRequestDone {
					log.Println("on time: ", t)
					eb.pub(event{t: METER_NAME_CHANGED, e: r.lastMeterName})
					r.lastRequestDone = false
				}
			}
		}
	}()
}

func main() {
	flag.Parse()
	g, err := gocui.NewGui(gocui.OutputNormal)
	client := NewClient(*metricsAddress, *metricsPath)
	// vu := newViewUpdater(g)
	// eb.regist(vu)

	// log.SetFlags(log.L)
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
	refresher := newRefresher(1,eb)

	g.Cursor = true
	sideBar := newSideBar(g, client, eb)
	detail := newDetail(g, client, eb)
	sideBar.init()
	refresher.start()
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

func keybindings(g *gocui.Gui, s *SideBar) error {
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
