package engine

import (
	"context"
	"fmt"
	"github.com/pymba86/bingo/pkg/service"
	"sync"
	"time"
)

var LoadedExchangeStrategies = make(map[string]SingleExchangeStrategy)
var LoadedCrossExchangeStrategies = make(map[string]CrossExchangeStrategy)

func RegisterStrategy(key string, s interface{}) {
	loaded := 0
	if d, ok := s.(SingleExchangeStrategy); ok {
		LoadedExchangeStrategies[key] = d
		loaded++
	}

	if d, ok := s.(CrossExchangeStrategy); ok {
		LoadedCrossExchangeStrategies[key] = d
		loaded++
	}

	if loaded == 0 {
		panic(fmt.Errorf("%T does not implement SingleExchangeStrategy or CrossExchangeStrategy", s))
	}
}

type SyncStatus int

const (
	SyncNotStarted SyncStatus = iota
	Syncing
	SyncDone
)

type Environment struct {
	Notifiability

	SyncService              *service.SyncService

	// startTime is the time of start point (which is used in the backtest)
	startTime time.Time

	// syncStartTime is the time point we want to start the sync (for trades and orders)
	syncStartTime time.Time
	syncMutex     sync.Mutex

	syncStatusMutex sync.Mutex
	syncStatus      SyncStatus

	sessions map[string]*ExchangeSession
}

func NewEnvironment() *Environment {
	return &Environment{
		sessions:  make(map[string]*ExchangeSession),
		startTime: time.Now(),
	}
}

func (e *Environment) Start(ctx context.Context) error {
	return nil
}

func (e *Environment) Connect(ctx context.Context) error {
	return nil
}

func (e *Environment) ConfigureExchangeSessions(userConfig *Config) error {
	return e.AddExchangesFromSessionConfig(userConfig.Sessions)
}

func (e *Environment) AddExchangesFromSessionConfig(sessions map[string]*ExchangeSession) error {
	for sessionName, session := range sessions {
		if err := InitExchangeSession(sessionName, session); err != nil {
			return err
		}

		e.AddExchangeSession(sessionName, session)
	}

	return nil
}

func (e *Environment) AddExchangeSession(name string, session *ExchangeSession) *ExchangeSession {

	session.Notifiability = e.Notifiability

	e.sessions[name] = session
	return session
}

func (environ *Environment) Init(ctx context.Context) (err error) {
	for n := range environ.sessions {
		var session = environ.sessions[n]
		if err = session.Init(ctx, environ); err != nil {
			// we can skip initialized sessions
			if err != ErrSessionAlreadyInitialized {
				return err
			}
		}
	}

	return
}



