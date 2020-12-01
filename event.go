package main

import (
	"sync"
)

type evtType int8

const (
	METER_NAME_CHANGED evtType = 1
	REQUEST_UPDATE_VIEW evtType = 2
)

type event struct {
	e string
	t evtType //type
}

type subscriber interface {
	subscribe(evt event)
}

type eventBus interface {
	pub(evt event)
	regist(subscriber subscriber)
}

type defaultEventBus struct {
	subs   []subscriber
	rwLock *sync.RWMutex
}

func newEventBus() eventBus {
	return &defaultEventBus{make([]subscriber,0),&sync.RWMutex{}}
}

func (e *defaultEventBus) pub(evt event) {
	e.rwLock.RLock()
	defer e.rwLock.RUnlock()
	for _,sub := range e.subs {
		sub.subscribe(evt)
	}
}

func (e *defaultEventBus) regist(subscriber subscriber) {
	e.rwLock.Lock()
	defer e.rwLock.Unlock()
	e.subs = append(e.subs,subscriber)
}
