package subpub

import (
	"context"
	"errors"
	"sync"
)

// MessageHandler — функция-обработчик сообщений.
type MessageHandler func(msg interface{})

// Subscription — интерфейс для управления подпиской.
type Subscription interface {
	Unsubscribe() error
}

// SubPub — интерфейс шины событий.
type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

// subscriber — структура подписчика.
type subscriber struct {
	handler MessageHandler
	ch      chan interface{}
	done    chan struct{}
	once    sync.Once
	wg      *sync.WaitGroup
	unsub   func()
}

// Unsubscribe — отписка от событий.
func (s *subscriber) Unsubscribe() error {
	s.once.Do(func() {
		close(s.done)
		s.unsub()
	})
	return nil
}

// subject — структура для хранения подписчиков по теме.
type subject struct {
	mu          sync.RWMutex
	subscribers map[*subscriber]struct{}
}

// subPub — основная реализация SubPub.
type subPub struct {
	mu       sync.RWMutex
	subjects map[string]*subject
	closed   bool
	wg       sync.WaitGroup
}

// NewSubPub — конструктор SubPub.
func NewSubPub() SubPub {
	return &subPub{
		subjects: make(map[string]*subject),
	}
}

// Subscribe — подписка на события по теме.
func (sp *subPub) Subscribe(subjectName string, cb MessageHandler) (Subscription, error) {
	sp.mu.Lock()
	if sp.closed {
		sp.mu.Unlock()
		return nil, errors.New("subpub is closed")
	}
	subj, ok := sp.subjects[subjectName]
	if !ok {
		subj = &subject{
			subscribers: make(map[*subscriber]struct{}),
		}
		sp.subjects[subjectName] = subj
	}
	sp.mu.Unlock()

	sub := &subscriber{
		handler: cb,
		ch:      make(chan interface{}, 64), // буфер для медленных подписчиков
		done:    make(chan struct{}),
		wg:      &sp.wg,
	}

	// Функция отписки
	sub.unsub = func() {
		subj.mu.Lock()
		delete(subj.subscribers, sub)
		subj.mu.Unlock()
		close(sub.ch)
	}

	subj.mu.Lock()
	subj.subscribers[sub] = struct{}{}
	subj.mu.Unlock()

	sp.wg.Add(1)
	go func() {
		defer sp.wg.Done()
		for {
			select {
			case msg, ok := <-sub.ch:
				if !ok {
					return
				}
				sub.handler(msg)
			case <-sub.done:
				return
			}
		}
	}()

	return sub, nil
}

// Publish — публикация сообщения по теме.
func (sp *subPub) Publish(subjectName string, msg interface{}) error {
	sp.mu.RLock()
	if sp.closed {
		sp.mu.RUnlock()
		return errors.New("subpub is closed")
	}
	subj, ok := sp.subjects[subjectName]
	sp.mu.RUnlock()
	if !ok {
		return nil // Нет подписчиков — ничего не делаем
	}

	subj.mu.RLock()
	defer subj.mu.RUnlock()
	for sub := range subj.subscribers {
		select {
		case sub.ch <- msg:
		default:
			// Если канал переполнен — блокируемся (FIFO, не теряем порядок)
			sub.ch <- msg
		}
	}
	return nil
}

// Close — завершение работы шины.
func (sp *subPub) Close(ctx context.Context) error {
	sp.mu.Lock()
	if sp.closed {
		sp.mu.Unlock()
		return nil
	}
	sp.closed = true
	// Копируем subjects для дальнейшей работы вне блокировки
	subjects := make([]*subject, 0, len(sp.subjects))
	for _, subj := range sp.subjects {
		subjects = append(subjects, subj)
	}
	sp.mu.Unlock()

	// Закрываем все подписки
	var allSubs []*subscriber
	for _, subj := range subjects {
		subj.mu.RLock()
		for sub := range subj.subscribers {
			allSubs = append(allSubs, sub)
		}
		subj.mu.RUnlock()
	}

	for _, sub := range allSubs {
		sub.Unsubscribe()
	}

	// Ждём завершения всех горутин или отмены контекста
	done := make(chan struct{})
	go func() {
		sp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
