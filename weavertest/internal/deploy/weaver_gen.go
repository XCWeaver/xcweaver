// Code generated by "weaver generate". DO NOT EDIT.
//go:build !ignoreWeaverGen

package deploy

import (
	"context"
	"errors"
	"github.com/ServiceWeaver/weaver"
	"github.com/ServiceWeaver/weaver/runtime/codegen"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"reflect"
)

func init() {
	codegen.Register(codegen.Registration{
		Name:  "github.com/ServiceWeaver/weaver/weavertest/internal/deploy/Started",
		Iface: reflect.TypeOf((*Started)(nil)).Elem(),
		Impl:  reflect.TypeOf(started{}),
		LocalStubFn: func(impl any, caller string, tracer trace.Tracer) any {
			return started_local_stub{impl: impl.(Started), tracer: tracer, markStartedMetrics: codegen.MethodMetricsFor(codegen.MethodLabels{Caller: caller, Component: "github.com/ServiceWeaver/weaver/weavertest/internal/deploy/Started", Method: "MarkStarted", Remote: false, Generated: true})}
		},
		ClientStubFn: func(stub codegen.Stub, caller string) any {
			return started_client_stub{stub: stub, markStartedMetrics: codegen.MethodMetricsFor(codegen.MethodLabels{Caller: caller, Component: "github.com/ServiceWeaver/weaver/weavertest/internal/deploy/Started", Method: "MarkStarted", Remote: true, Generated: true})}
		},
		ServerStubFn: func(impl any, addLoad func(uint64, float64)) codegen.Server {
			return started_server_stub{impl: impl.(Started), addLoad: addLoad}
		},
		ReflectStubFn: func(caller func(string, context.Context, []any, []any) error) any {
			return started_reflect_stub{caller: caller}
		},
		RefData: "",
	})
	codegen.Register(codegen.Registration{
		Name:  "github.com/ServiceWeaver/weaver/weavertest/internal/deploy/Widget",
		Iface: reflect.TypeOf((*Widget)(nil)).Elem(),
		Impl:  reflect.TypeOf(widget{}),
		LocalStubFn: func(impl any, caller string, tracer trace.Tracer) any {
			return widget_local_stub{impl: impl.(Widget), tracer: tracer, useMetrics: codegen.MethodMetricsFor(codegen.MethodLabels{Caller: caller, Component: "github.com/ServiceWeaver/weaver/weavertest/internal/deploy/Widget", Method: "Use", Remote: false, Generated: true})}
		},
		ClientStubFn: func(stub codegen.Stub, caller string) any {
			return widget_client_stub{stub: stub, useMetrics: codegen.MethodMetricsFor(codegen.MethodLabels{Caller: caller, Component: "github.com/ServiceWeaver/weaver/weavertest/internal/deploy/Widget", Method: "Use", Remote: true, Generated: true})}
		},
		ServerStubFn: func(impl any, addLoad func(uint64, float64)) codegen.Server {
			return widget_server_stub{impl: impl.(Widget), addLoad: addLoad}
		},
		ReflectStubFn: func(caller func(string, context.Context, []any, []any) error) any {
			return widget_reflect_stub{caller: caller}
		},
		RefData: "⟦f3fa3c18:wEaVeReDgE:github.com/ServiceWeaver/weaver/weavertest/internal/deploy/Widget→github.com/ServiceWeaver/weaver/weavertest/internal/deploy/Started⟧\n",
	})
}

// weaver.InstanceOf checks.
var _ weaver.InstanceOf[Started] = (*started)(nil)
var _ weaver.InstanceOf[Widget] = (*widget)(nil)

// weaver.Router checks.
var _ weaver.Unrouted = (*started)(nil)
var _ weaver.Unrouted = (*widget)(nil)

// Local stub implementations.

type started_local_stub struct {
	impl               Started
	tracer             trace.Tracer
	markStartedMetrics *codegen.MethodMetrics
}

// Check that started_local_stub implements the Started interface.
var _ Started = (*started_local_stub)(nil)

func (s started_local_stub) MarkStarted(ctx context.Context, a0 string) (err error) {
	// Update metrics.
	begin := s.markStartedMetrics.Begin()
	defer func() { s.markStartedMetrics.End(begin, err != nil, 0, 0) }()
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		// Create a child span for this method.
		ctx, span = s.tracer.Start(ctx, "deploy.Started.MarkStarted", trace.WithSpanKind(trace.SpanKindInternal))
		defer func() {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.End()
		}()
	}

	return s.impl.MarkStarted(ctx, a0)
}

type widget_local_stub struct {
	impl       Widget
	tracer     trace.Tracer
	useMetrics *codegen.MethodMetrics
}

// Check that widget_local_stub implements the Widget interface.
var _ Widget = (*widget_local_stub)(nil)

func (s widget_local_stub) Use(ctx context.Context, a0 string) (err error) {
	// Update metrics.
	begin := s.useMetrics.Begin()
	defer func() { s.useMetrics.End(begin, err != nil, 0, 0) }()
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		// Create a child span for this method.
		ctx, span = s.tracer.Start(ctx, "deploy.Widget.Use", trace.WithSpanKind(trace.SpanKindInternal))
		defer func() {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.End()
		}()
	}

	return s.impl.Use(ctx, a0)
}

// Client stub implementations.

type started_client_stub struct {
	stub               codegen.Stub
	markStartedMetrics *codegen.MethodMetrics
}

// Check that started_client_stub implements the Started interface.
var _ Started = (*started_client_stub)(nil)

func (s started_client_stub) MarkStarted(ctx context.Context, a0 string) (err error) {
	// Update metrics.
	var requestBytes, replyBytes int
	begin := s.markStartedMetrics.Begin()
	defer func() { s.markStartedMetrics.End(begin, err != nil, requestBytes, replyBytes) }()

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		// Create a child span for this method.
		ctx, span = s.stub.Tracer().Start(ctx, "deploy.Started.MarkStarted", trace.WithSpanKind(trace.SpanKindClient))
	}

	defer func() {
		// Catch and return any panics detected during encoding/decoding/rpc.
		if err == nil {
			err = codegen.CatchPanics(recover())
			if err != nil {
				err = errors.Join(weaver.RemoteCallError, err)
			}
		}

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()

	}()

	// Preallocate a buffer of the right size.
	size := 0
	size += (4 + len(a0))
	enc := codegen.NewEncoder()
	enc.Reset(size)

	// Encode arguments.
	enc.String(a0)
	var shardKey uint64

	// Call the remote method.
	requestBytes = len(enc.Data())
	var results []byte
	results, err = s.stub.Run(ctx, 0, enc.Data(), shardKey)
	replyBytes = len(results)
	if err != nil {
		err = errors.Join(weaver.RemoteCallError, err)
		return
	}

	// Decode the results.
	dec := codegen.NewDecoder(results)
	err = dec.Error()
	return
}

type widget_client_stub struct {
	stub       codegen.Stub
	useMetrics *codegen.MethodMetrics
}

// Check that widget_client_stub implements the Widget interface.
var _ Widget = (*widget_client_stub)(nil)

func (s widget_client_stub) Use(ctx context.Context, a0 string) (err error) {
	// Update metrics.
	var requestBytes, replyBytes int
	begin := s.useMetrics.Begin()
	defer func() { s.useMetrics.End(begin, err != nil, requestBytes, replyBytes) }()

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		// Create a child span for this method.
		ctx, span = s.stub.Tracer().Start(ctx, "deploy.Widget.Use", trace.WithSpanKind(trace.SpanKindClient))
	}

	defer func() {
		// Catch and return any panics detected during encoding/decoding/rpc.
		if err == nil {
			err = codegen.CatchPanics(recover())
			if err != nil {
				err = errors.Join(weaver.RemoteCallError, err)
			}
		}

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()

	}()

	// Preallocate a buffer of the right size.
	size := 0
	size += (4 + len(a0))
	enc := codegen.NewEncoder()
	enc.Reset(size)

	// Encode arguments.
	enc.String(a0)
	var shardKey uint64

	// Call the remote method.
	requestBytes = len(enc.Data())
	var results []byte
	results, err = s.stub.Run(ctx, 0, enc.Data(), shardKey)
	replyBytes = len(results)
	if err != nil {
		err = errors.Join(weaver.RemoteCallError, err)
		return
	}

	// Decode the results.
	dec := codegen.NewDecoder(results)
	err = dec.Error()
	return
}

// Note that "weaver generate" will always generate the error message below.
// Everything is okay. The error message is only relevant if you see it when
// you run "go build" or "go run".
var _ codegen.LatestVersion = codegen.Version[[0][24]struct{}](`

ERROR: You generated this file with 'weaver generate' (devel) (codegen
version v0.24.0). The generated code is incompatible with the version of the
github.com/ServiceWeaver/weaver module that you're using. The weaver module
version can be found in your go.mod file or by running the following command.

    go list -m github.com/ServiceWeaver/weaver

We recommend updating the weaver module and the 'weaver generate' command by
running the following.

    go get github.com/ServiceWeaver/weaver@latest
    go install github.com/ServiceWeaver/weaver/cmd/weaver@latest

Then, re-run 'weaver generate' and re-build your code. If the problem persists,
please file an issue at https://github.com/ServiceWeaver/weaver/issues.

`)

// Server stub implementations.

type started_server_stub struct {
	impl    Started
	addLoad func(key uint64, load float64)
}

// Check that started_server_stub implements the codegen.Server interface.
var _ codegen.Server = (*started_server_stub)(nil)

// GetStubFn implements the codegen.Server interface.
func (s started_server_stub) GetStubFn(method string) func(ctx context.Context, args []byte) ([]byte, error) {
	switch method {
	case "MarkStarted":
		return s.markStarted
	default:
		return nil
	}
}

func (s started_server_stub) markStarted(ctx context.Context, args []byte) (res []byte, err error) {
	// Catch and return any panics detected during encoding/decoding/rpc.
	defer func() {
		if err == nil {
			err = codegen.CatchPanics(recover())
		}
	}()

	// Decode arguments.
	dec := codegen.NewDecoder(args)
	var a0 string
	a0 = dec.String()

	// TODO(rgrandl): The deferred function above will recover from panics in the
	// user code: fix this.
	// Call the local method.
	appErr := s.impl.MarkStarted(ctx, a0)

	// Encode the results.
	enc := codegen.NewEncoder()
	enc.Error(appErr)
	return enc.Data(), nil
}

type widget_server_stub struct {
	impl    Widget
	addLoad func(key uint64, load float64)
}

// Check that widget_server_stub implements the codegen.Server interface.
var _ codegen.Server = (*widget_server_stub)(nil)

// GetStubFn implements the codegen.Server interface.
func (s widget_server_stub) GetStubFn(method string) func(ctx context.Context, args []byte) ([]byte, error) {
	switch method {
	case "Use":
		return s.use
	default:
		return nil
	}
}

func (s widget_server_stub) use(ctx context.Context, args []byte) (res []byte, err error) {
	// Catch and return any panics detected during encoding/decoding/rpc.
	defer func() {
		if err == nil {
			err = codegen.CatchPanics(recover())
		}
	}()

	// Decode arguments.
	dec := codegen.NewDecoder(args)
	var a0 string
	a0 = dec.String()

	// TODO(rgrandl): The deferred function above will recover from panics in the
	// user code: fix this.
	// Call the local method.
	appErr := s.impl.Use(ctx, a0)

	// Encode the results.
	enc := codegen.NewEncoder()
	enc.Error(appErr)
	return enc.Data(), nil
}

// Reflect stub implementations.

type started_reflect_stub struct {
	caller func(string, context.Context, []any, []any) error
}

// Check that started_reflect_stub implements the Started interface.
var _ Started = (*started_reflect_stub)(nil)

func (s started_reflect_stub) MarkStarted(ctx context.Context, a0 string) (err error) {
	err = s.caller("MarkStarted", ctx, []any{a0}, []any{})
	return
}

type widget_reflect_stub struct {
	caller func(string, context.Context, []any, []any) error
}

// Check that widget_reflect_stub implements the Widget interface.
var _ Widget = (*widget_reflect_stub)(nil)

func (s widget_reflect_stub) Use(ctx context.Context, a0 string) (err error) {
	err = s.caller("Use", ctx, []any{a0}, []any{})
	return
}
