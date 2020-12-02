package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	// "strconv"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/jroimartin/gocui"
)

type detail struct {
	meterName string

	width    int
	height   int
	g        *gocui.Gui
	client   *Client
	eventBus eventBus
	mutex    *sync.Mutex
}

type TagMeter struct {
	meter *Meter
	tag   string
}

func newDetail(g *gocui.Gui, client *Client, eventBus eventBus) *detail {

	summary := &detail{g: g,
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
		s.eventBus.pub(event{t: CLIENT_REQUEST_DONE})
	}
}
func (s *detail) updateTag(meterName string, meter *Meter) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFunc()
	tags := meter.flattenTags()
	c := len(tags)
	rc := make(chan interface{}, c)
	for _, tag := range tags {
		go func(tag string) {
			meter, err := s.client.Meter(ctx, meterName, tag)
			if err != nil {
				rc <- err
				return
			}
			rc <- TagMeter{meter, tag}
		}(tag)
	}

	timer := time.NewTimer(2 * time.Second)
	// sb := bytes.NewBufferString("")
	lines := make([]string, 0)
	for {
		select {
		case <-timer.C:
			cancelFunc()
			close(rc)
			// log.Println(sb.String())
			sort.Strings(lines)
			s.g.Update(func(g *gocui.Gui) error {
				v, err := g.View("tags")
				if err != nil {
					// handle error
				}
				v.Clear()
				tw := tabwriter.NewWriter(v, 0, 0, 1, '\t', tabwriter.AlignRight)
				for _, line := range lines {
					fmt.Fprintln(tw, line)
				}
				tw.Flush()

				return nil
			})
			return
		case v := <-rc:
			c--
			if _, ok := v.(error); ok {
				//sb.WriteString(fmt.Sprintf("%s %s\n", m.Statistic, err.Error()))

			} else {
				m := v.(TagMeter)
				for _, me := range m.meter.Measurements {
					// ss := s.expandStrings(m.tag, me.Statistic, strconv.FormatFloat(me.Value, 'f', 3, 64))
					// log.Println(ss)
					// sb.WriteString(ss + "\n")
					lines = append(lines, fmt.Sprintf("%s\t%s\t%f\t", m.tag, me.Statistic, me.Value))
				}
			}

			if c == 0 {
				sort.Strings(lines)
				s.g.Update(func(g *gocui.Gui) error {
					v, err := g.View("tags")
					if err != nil {
						// handle error
					}
					v.Clear()
					tw := tabwriter.NewWriter(v, 10, 5, 5, ' ', 0)
					for _, line := range lines {
						fmt.Fprintln(tw, line)
					}
					tw.Flush()
					return nil
				})
				return
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
	if sView, err := g.SetView("summary", meterNamesBottomX, 1, maxX-1, 10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		sView.Title = "summary"

	}
	if tagsView, err := g.SetView("tags", meterNamesBottomX, 10, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		tagsView.Title = "statics"
	}

	s.width = maxX - 1 - meterNamesBottomX
	return nil
}

func (s *detail) expandStrings(strs ...string) string {
	var ssize int
	for _, str := range strs {
		ssize += len(str)
	}

	space := s.width - ssize
	spacePerStr := int(space / (len(strs) - 1))
	sb := bytes.NewBufferString(strs[0])
	for i := 1; i < len(strs); i++ {
		sb.WriteString(strings.Repeat(" ", spacePerStr) + strs[i])
	}
	return sb.String()
}
