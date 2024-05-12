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

func NewDependencyRegistryWith(errFmtSvc errorFormatterService) *conainerRegistry {
	return &conainerRegistry{
		mu: &sync.Mutex{},

		l: nil,
		e: errFmtSvc,

		list: make([]any, 0, DefaultRegistryBufferSize),
		reg:  make(map[ContainerUnitType]uint32, DefaultRegistryBufferSize),
	}
}
