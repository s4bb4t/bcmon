package appcloser

import (
	"sync"

	"go.uber.org/zap"
)

type (
	// AppCloser -
	AppCloser struct {
		mu         *sync.Mutex
		components []component
		logger     *zap.Logger
	}

	component struct {
		closer  func()
		appName string
	}
)

// InitCloser - ##
func InitCloser(l *zap.Logger) *AppCloser {
	return &AppCloser{
		mu:         &sync.Mutex{},
		components: []component{},
		logger:     l,
	}
}

// AddCloser -
func (s *AppCloser) AddCloser(closer func(), appName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.components = append(s.components, component{closer: closer, appName: appName})
}

// CloseAll -
func (s *AppCloser) CloseAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	wg := &sync.WaitGroup{}
	wg.Add(len(s.components))

	for _, c := range s.components {
		wg.Add(1)
		go func(wg *sync.WaitGroup, c component) {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Info("panic recovered: CloseAll", zap.String("app name", c.appName))
				}
				wg.Done()
			}()
			c.closer()
		}(wg, c)
	}

	wg.Wait()
}
