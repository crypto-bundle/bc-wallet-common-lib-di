package di

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"
)

// testWriter routes writes directly into testing.TB
type testWriter struct {
	t testing.TB
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(bytes.TrimSpace(p)))
	return len(p), nil
}

func TestConainerProvider_Resolving(t *testing.T) {
	t.Run("test logger dependency resolving one unit", func(t *testing.T) {
		const LoggerDepUnitType ContainerUnitType = 1
		const SomeDepUnitType ContainerUnitType = 2

		registry := NewDependencyRegistry()
		errFmtSvc := newErrFormatter()
		defaultDepList := []any{registry, errFmtSvc}

		err := DIMustWith[slog.Logger](defaultDepList...).RegisterLazy(LoggerDepUnitType, func() *slog.Logger {
			return slog.New(slog.NewTextHandler(testWriter{t: t}, nil))
		})
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		logger, err := DIMustWith[slog.Logger](defaultDepList...).Resolve(LoggerDepUnitType)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		if logger == nil {
			t.Error(errors.New("nil deps (logger) returned from DI contaner"))
			t.FailNow()
		}

		logger.Info("test success")
	})
}

func TestConainerProvider_ResolvingMultiple(t *testing.T) {
	t.Run("test logger dependency resolving multiple units", func(t *testing.T) {
		const LoggerDepUnitType ContainerUnitType = 1
		const SomeDepUnitType ContainerUnitType = 2

		registry := NewDependencyRegistry()
		errFmtSvc := newErrFormatter()
		defaultDepList := []any{registry, errFmtSvc}

		type someDepStruct struct {
			log *slog.Logger
		}

		err := DIMustWith[slog.Logger](defaultDepList...).RegisterLazy(LoggerDepUnitType, func() *slog.Logger {
			return slog.New(slog.NewTextHandler(testWriter{t: t}, nil))
		})
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		err = DIMustWith[someDepStruct](defaultDepList...).RegisterLazy(SomeDepUnitType, func() *someDepStruct {
			logger, err := DIMustWith[slog.Logger](defaultDepList...).Resolve(LoggerDepUnitType)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			if logger == nil {
				t.Error(errors.New("nil deps (logger) returned from DI contaner"))
				t.FailNow()
			}

			return &someDepStruct{
				log: logger,
			}
		})
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		someDep, err := DIMustWith[someDepStruct](defaultDepList...).Resolve(SomeDepUnitType)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		if someDep == nil {
			t.Error(errors.New("nil deps returned from DI contaner"))
			t.FailNow()
		}

		someDep.log.Info("some logger message")
	})
}
