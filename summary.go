package main

import (
	"fmt"
	"context"
	"time"
	"log"
	"bytes"
	"github.com/jroimartin/gocui"
)

type summary struct {
	summaryView *gocui.View
	tagView     *gocui.View
	meterName   string

	client *Client
	eventBus eventBus
	onFlyCtxCancelFunc context.CancelFunc

}

func newSummary(summaryView, tagView *gocui.View, client *Client, eventBus eventBus) *summary {

	summary := &summary{summaryView: summaryView, tagView: tagView, client: client,eventBus:eventBus}
	eventBus.regist(summary)
	return summary
}

func (s *summary) subscribe(evt event) {
	if evt.t == METER_NAME_CHANGED {
		meterName := evt.e
		if s.onFlyCtxCancelFunc != nil {
			s.onFlyCtxCancelFunc()
		}
		ctx,cancelFunc := context.WithTimeout(context.Background(),3 * time.Second)
		s.onFlyCtxCancelFunc = cancelFunc
		meter,err := s.client.Meter(ctx,meterName)
		if err != nil {
			log.Println(err)
			log.Fatalln(err)
			return
		}
		log.Println(*meter)
		sb := bytes.NewBufferString("")
		for _,m := range meter.Measurements {
			sb.WriteString(fmt.Sprintf("%s %f\n",m.Statistic,m.Value))
		}
		s.summaryView.Title = meter.Description
		s.summaryView.Clear()
		fmt.Fprint(s.summaryView,sb.String())
		evt := event{t:REQUEST_UPDATE_VIEW}
		s.eventBus.pub(evt)
		
	}
}
