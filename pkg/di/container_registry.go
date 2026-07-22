package di

import (
	"errors"
	"log/slog"
	"sync"
)

var defaultContainerRegistry *conainerRegistry = NewDependencyRegistry()

var (
	ErrDependencyUnitAlreadyExists = errors.New("dependency unit already exists")
	ErrDependencyUnitMissing       = errors.New("dependency unit is not exists")
	ErrDependencyObjectMissing     = errors.New("dependency object is not exists")
)

// SubOpt configures options for subscribing to JetStream consumers.
type RegistryOptions interface {
	configureParams(opts *registyOptionsParams) error
}

type registyOptionsParams struct {
	errorFmtSvc errorFormatterService
	loggerSvc   *slog.Logger
}

// sdiOptFn is a function option used to configure a DI container instance
type registryOptFn func(opts *registyOptionsParams) error

func (opt registryOptFn) configureParams(opts *registyOptionsParams) error {
	return opt(opts)
}

func RegistryErrFmtOpt(errFmtSvc errorFormatterService) RegistryOptions {
	return registryOptFn(func(opts *registyOptionsParams) error {
		var svc errorFormatterService = errFmtSvc

		if svc == nil {
			svc = defaultErrorsFmtSvc
		}

		opts.errorFmtSvc = svc

		return nil
	})
}

func RegistryLoggerOpt(loggerSvc *slog.Logger) RegistryOptions {
	return registryOptFn(func(opts *registyOptionsParams) error {
		var slogSvc *slog.Logger = loggerSvc

		if slogSvc == nil {
			slogSvc = slog.Default()
		}

		opts.loggerSvc = slogSvc

		return nil
	})
}

type conainerRegistry struct {
	mu *sync.Mutex

	l *slog.Logger
	e errorFormatterService

	list []any
	reg  map[ContainerUnitType]uint32
}

func (m *conainerRegistry) Register(unitType ContainerUnitType, unitService any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.register(unitType, unitService)
}

func (m *conainerRegistry) register(unitType ContainerUnitType, unitService any) error {
	_, isExists := m.reg[unitType]
	if isExists {
		return m.e.Errorf(ErrDependencyUnitAlreadyExists, "unit type: %d", unitType)
	}

	m.list = append(m.list, unitService)
	m.reg[unitType] = uint32(len(m.list) - 1)

	return nil
}

func (m *conainerRegistry) Get(unitType ContainerUnitType) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.get(unitType)
}

func (m *conainerRegistry) get(unitType ContainerUnitType) (any, error) {
	unitPosition, isExists := m.reg[unitType]
	if !isExists {
		return nil, m.e.Errorf(ErrDependencyUnitMissing, "unit type: %d", unitType)
	}

	return m.list[unitPosition], nil
}

func Default() *conainerRegistry {
	return defaultContainerRegistry
}

func NewDependencyRegistry() *conainerRegistry {
	return &conainerRegistry{
		mu: &sync.Mutex{},

		l: slog.Default(),
		e: defaultErrorsFmtSvc,

		list: make([]any, 0, DefaultRegistryBufferSize),
		reg:  make(map[ContainerUnitType]uint32, DefaultRegistryBufferSize),
	}
}

func NewDependencyRegistryWith(containerOptions ...RegistryOptions) *conainerRegistry {
	cfg := &registyOptionsParams{
		errorFmtSvc: nil,
		loggerSvc:   nil,
	}

	for _, opt := range containerOptions {
		if opt == nil {
			continue
		}

		if loopErr := opt.configureParams(cfg); loopErr != nil {
			panic(defaultErrorsFmtSvc.ErrNoWrap(ErrUnableToConfigureParams))
		}
	}

	return &conainerRegistry{
		mu: &sync.Mutex{},

		l: cfg.loggerSvc,
		e: cfg.errorFmtSvc,

		list: make([]any, 0, DefaultRegistryBufferSize),
		reg:  make(map[ContainerUnitType]uint32, DefaultRegistryBufferSize),
	}
}
