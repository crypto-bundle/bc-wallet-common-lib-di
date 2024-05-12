package di

import (
	"sync"
)

type ConainerProvider[T any] struct {
	registrySvc dependencyRegistryService
	e           errorFormatterService
}

func (r *ConainerProvider[T]) Resolve(unitType ContainerUnitType) (*T, error) {
	unit, err := r.registrySvc.Get(unitType)
	if err != nil {
		return nil, r.e.ErrNoWrap(err)
	}

	if unit == nil {
		return nil, r.e.ErrNoWrap(ErrDependencyUnitMissing)
	}

	castedUnitWrapper, isCasted := unit.(*dependencyUnitWrapper[T])
	if !isCasted {
		return nil, r.e.ErrNoWrap(ErrUnableToCastDependencyUnit)
	}

	resolvedObjectInst, err := castedUnitWrapper.Resolve()
	if err != nil {
		return nil, r.e.ErrNoWrap(err)
	}

	if resolvedObjectInst == nil {
		return nil, r.e.ErrNoWrap(ErrDependencyObjectMissing)
	}

	return resolvedObjectInst, nil
}

func (r *ConainerProvider[T]) RegisterLazy(unitType ContainerUnitType, factory func() *T) error {
	err := r.registrySvc.Register(unitType, &dependencyUnitWrapper[T]{
		once:    sync.Once{},
		e:       r.e,
		factory: factory,
		object:  nil,
	})
	if err != nil {
		return r.e.ErrNoWrap(err)
	}

	return nil
}

func newContainerProvider[T any](errFmtSvc errorFormatterService,
	registrySvc dependencyRegistryService,
) *ConainerProvider[T] {
	return &ConainerProvider[T]{
		e:           errFmtSvc,
		registrySvc: registrySvc,
	}
}
