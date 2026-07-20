/*
 * MIT NON-AI License
 *
 * Copyright (c) 2022-2025 Aleksei Kotelnikov(gudron2s@gmail.com)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of the software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions.
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
 *
 * In addition, the following restrictions apply:
 *
 * 1. The Software and any modifications made to it may not be used for the purpose of training or improving machine learning algorithms,
 * including but not limited to artificial intelligence, natural language processing, or data mining. This condition applies to any derivatives,
 * modifications, or updates based on the Software code. Any usage of the Software in an AI-training dataset is considered a breach of this License.
 *
 * 2. The Software may not be included in any dataset used for training or improving machine learning algorithms,
 * including but not limited to artificial intelligence, natural language processing, or data mining.
 *
 * 3. Any person or organization found to be in violation of these restrictions will be subject to legal action and may be held liable
 * for any damages resulting from such use.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
 * DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
 * OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 */

package di

import (
	"errors"
	"log/slog"
)

var (
	ErrUnableToCastDependencyUnit  = errors.New("ubable to cast dependency unit")
	ErrMissingRequiredDependencies = errors.New("missing required dependencies")
	ErrUnableToConfigureParams     = errors.New("unable to configure params")
)

// SubOpt configures options for subscribing to JetStream consumers.
type DIOptions interface {
	configureParams(opts *diOptionsParams) error
}

type diOptionsParams struct {
	errorFmtSvc errorFormatterService
	loggerSvc   *slog.Logger
	registrySvc dependencyRegistryService
}

// sdiOptFn is a function option used to configure a DI container instance
type diOptFn func(opts *diOptionsParams) error

func (opt diOptFn) configureParams(opts *diOptionsParams) error {
	return opt(opts)
}

func ErrorFormatterOpt(errFmtSvc errorFormatterService) DIOptions {
	return diOptFn(func(opts *diOptionsParams) error {
		var svc errorFormatterService = errFmtSvc

		if svc == nil {
			svc = defaultErrorsFmtSvc
		}

		opts.errorFmtSvc = svc

		return nil
	})
}

func LoggerOpt(loggerSvc *slog.Logger) DIOptions {
	return diOptFn(func(opts *diOptionsParams) error {
		var slogSvc *slog.Logger = loggerSvc

		if slogSvc == nil {
			slogSvc = slog.Default()
		}

		opts.loggerSvc = slogSvc

		return nil
	})
}

func RegistryOpt(registrySvc dependencyRegistryService) DIOptions {
	return diOptFn(func(opts *diOptionsParams) error {
		var svc dependencyRegistryService = registrySvc

		if svc == nil {
			svc = NewDependencyRegistry()
		}

		opts.registrySvc = svc

		return nil
	})
}

func DIMustWith[T any](containerOptions ...DIOptions) *ConainerProvider[T] {
	cfg := &diOptionsParams{
		errorFmtSvc: nil,
		loggerSvc:   nil,
		registrySvc: nil,
	}

	for _, opt := range containerOptions {
		if opt == nil {
			continue
		}

		if loopErr := opt.configureParams(cfg); loopErr != nil {
			panic(defaultErrorsFmtSvc.ErrNoWrap(ErrUnableToConfigureParams))
		}
	}

	if cfg.errorFmtSvc == nil {
		panic(defaultErrorsFmtSvc.Errorf(ErrMissingRequiredDependencies, "dep type: %s", "error formatter"))
	}

	if cfg.registrySvc == nil {
		panic(defaultErrorsFmtSvc.Errorf(ErrMissingRequiredDependencies, "dep type: %s", "dependencies registry"))
	}

	return newContainerProvider[T](cfg.errorFmtSvc, cfg.registrySvc)
}
