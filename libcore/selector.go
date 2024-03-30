package libcore

import (
	"context"
	"time"

	"github.com/sagernet/sing-box/log"
)

type selectorCallback func(selectorTag, tag string)

/*
listenSelectorChange check the change of the main selector once in a while and use callback.
It will block the thread, so run it in a new goroutine.
*/
func (b *BoxInstance) listenSelectorChange(ctx context.Context, callback selectorCallback) {
	if b.selector == nil || callback == nil {
		return
	}

	defer catchPanic("listenSelectorChange", func(panicErr error) { log.Error(panicErr) })

	const (
		duration0 = 500 * time.Millisecond
		duration1 = 700 * time.Millisecond
		duration2 = 1000 * time.Millisecond
		duration3 = 2000 * time.Millisecond
	)

	nowDuration := duration0
	ticker := time.NewTicker(nowDuration)
	defer ticker.Stop()

	// updateTicker 动态调整 ticker 间隔。如果选择的标签已更改，则重置为默认的间隔，如果没有则逐级延长检查间隔。
	updateTicker := func(changed bool) {
		if changed {
			nowDuration = duration0
		} else {
			switch nowDuration {
			case duration0:
				nowDuration = duration1
			case duration1:
				nowDuration = duration2
			case duration2:
				nowDuration = duration3
			case duration3:
				nowDuration = duration0
			default:
				nowDuration = duration0
			}
		}

		ticker.Reset(nowDuration)
	}

	selectorTag := b.selector.Tag()
	oldTag := b.selector.Now()
	log.Trace("Started selector change listener")
	for {
		select {
		case <-ctx.Done():
			log.Trace("Selector change listener close: ", ctx.Err())
			return
		case <-ticker.C:
		}

		newTag := b.selector.Now()
		changed := oldTag != newTag
		if changed {
			callback(selectorTag, newTag)
			oldTag = newTag
		}
		updateTicker(changed)
	}
}
