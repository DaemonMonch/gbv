package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
)

type detail struct {
	meterName string

	g *gocui.Gui
	client   *Client
	eventBus eventBus
	mutex    *sync.Mutex
}

func newDetail(g *gocui.Gui,client *Client, eventBus eventBus) *detail {

	summary := &detail{g:g,
		client:   client,
		eventBus: eventBus,
		mutex:    &sync.Mutex{}}
	eventBus.regist(summary)
	return summary
}

func (s *detail) subscribe(evt event) {
	if evt.t == METER_NAME_CHANGED {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		meterName := evt.e
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		meter, err := s.client.Meter(ctx, meterName)
		if err != nil {
			log.Println(err)
			log.Fatalln(err)
			return
		}
		//	log.Println(*meter)
		s.updateSummary(meter)
		s.updateTag(meterName, meter)
	}
}
func (s *detail) updateTag(meterName string, meter *Meter) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelFunc()
	tags := meter.flattenTags()
		c := len(tags)
		rc := make(chan interface{}, c)
	for _, tag := range tags {
	log.Println(tag)
	go func(tag string) {
		meter, err := s.client.Meter(ctx, meterName, tag)
		if err != nil {
			rc <- err
			return
		}
		rc <- meter
	}(tag)
	}

	timer := time.NewTimer(2 * time.Second)
	sb := bytes.NewBufferString("")
	for {
		select {
		case <-timer.C:
			cancelFunc()
			close(rc)
			// log.Println(sb.String())
			s.g.Update(func(g *gocui.Gui) error {
				v, err := g.View("tags")
				if err != nil {
					// handle error
				}
				v.Clear()
				fmt.Fprint(v, sb.String())
				return nil
			})
			return
		case v := <-rc:
			if _, ok := v.(error); ok {
				//sb.WriteString(fmt.Sprintf("%s %s\n", m.Statistic, err.Error()))

			} else {
				m := v.(*Meter)
				for _, me := range m.Measurements {
					sb.WriteString(fmt.Sprintf("%s %f\n", me.Statistic, me.Value))
				}
			}
		}
	}

}

func (s *detail) updateSummary(meter *Meter) {
	sb := bytes.NewBufferString("")
	for _, m := range meter.Measurements {
		sb.WriteString(fmt.Sprintf("%s %f\n", m.Statistic, m.Value))
	}
	// s.summaryView.Title = meter.Description
	
	s.g.Update(func(g *gocui.Gui) error {
		v, err := g.View("summary")
		if err != nil {
			// handle error
		}
		v.Clear()
		fmt.Fprint(v, sb.String())
		return nil
	})
}

func (s *detail) Layout(g *gocui.Gui) error {
	_, _, meterNamesBottomX, _, _ := g.ViewPosition("side")
	maxX, maxY := g.Size()
	if _, err := g.SetView("summary", meterNamesBottomX, 1, maxX-1, 10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

	}
	if _, err := g.SetView("tags", meterNamesBottomX, 10, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	return nil
}
