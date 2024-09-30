// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package xcweaver provides the interface for building
// [single-image distributed programs].
//
// A program is composed of a set of Go interfaces called
// components. Components are recognized by "xcweaver generate" (typically invoked
// via "go generate"). "xcweaver generate" generates code that allows a component
// to be invoked over the network. This flexibility allows Service Weaver
// to decompose the program execution across many processes and machines.
//
// [single-image distributed programs]: https://serviceweaver.dev
package xcweaver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/XCWeaver/xcweaver/internal/antipode"
	"github.com/XCWeaver/xcweaver/internal/reflection"
	"github.com/XCWeaver/xcweaver/internal/xcweaver"
	"github.com/XCWeaver/xcweaver/runtime"
	"github.com/XCWeaver/xcweaver/runtime/codegen"
	"go.opentelemetry.io/otel/trace"
)

//go:generate ./dev/protoc.sh internal/status/status.proto
//go:generate ./dev/protoc.sh internal/tool/single/single.proto
//go:generate ./dev/protoc.sh internal/tool/ssh/impl/ssh.proto
//go:generate ./dev/protoc.sh runtime/protos/runtime.proto
//go:generate ./dev/protoc.sh runtime/protos/config.proto
//go:generate ./cmd/xcweaver/xcweaver generate . ./internal/tool/multi
//go:generate ./dev/writedeps.sh

// RemoteCallError indicates that a remote component method call failed to
// execute properly. This can happen, for example, because of a failed machine
// or a network partition. Here's an illustrative example:
//
//	// Call the foo.Foo method.
//	err := foo.Foo(ctx)
//	if errors.Is(err, xcweaver.RemoteCallError) {
//	    // foo.Foo did not execute properly.
//	} else if err != nil {
//	    // foo.Foo executed properly, but returned an error.
//	} else {
//	    // foo.Foo executed properly and did not return an error.
//	}
//
// Note that if a method call returns an error with an embedded
// RemoteCallError, it does NOT mean that the method never executed. The method
// may have executed partially or fully. Thus, you must be careful retrying
// method calls that result in a RemoteCallError. Ensuring that all methods are
// either read-only or idempotent is one way to ensure safe retries, for
// example.
var RemoteCallError = errors.New("Service Weaver remote call error")

// HealthzHandler is a health-check handler that returns an OK status for all
// incoming HTTP requests.
var HealthzHandler = func(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "OK")
}

// HealthzURL is the URL path on which Service Weaver performs health checks.
// Every application HTTP server must register a handler for this URL path,
// e.g.:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc(xcweaver.HealthzURL, func(http.ResponseWriter, *http.Request) {
//	    // ...
//	})
//
// As a convenience, Service Weaver registers HealthzHandler under
// this URL path in the default ServerMux, i.e.:
//
//	http.HandleFunc(xcweaver.HealthzURL, xcweaver.HealthzHandler)
const HealthzURL = "/debug/xcweaver/healthz"

var healthzInit sync.Once

// Main is the interface implemented by an application's main component.
type Main interface{}

// PointerToMain is a type constraint that asserts *T is an instance of Main
// (i.e. T is a struct that embeds xcweaver.Implements[xcweaver.Main]).
type PointerToMain[T any] interface {
	*T
	InstanceOf[Main]
}

// Run runs app as a Service Weaver application.
//
// The application is composed of a set of components that include xcweaver.Main
// as well as any components transitively needed by xcweaver.Main. An instance
// that implement xcweaver.Main is automatically created by xcweaver.Run and passed
// to app. Note: other replicas in which xcweaver.Run is called may also create
// instances of xcweaver.Main.
//
// The type T must be a struct type that contains an embedded
// `xcweaver.Implements[xcweaver.Main]` field. A value of type T is created,
// initialized (by calling its Init method if any), and a pointer to the value
// is passed to app. app contains the main body of the application; it will
// typically run HTTP servers, etc.
//
// If this process is hosting the `xcweaver.Main` component, Run will call app
// and will return when app returns. If this process is hosting other
// components, Run will start those components and never return. Most callers
// of Run will not do anything (other than possibly logging any returned error)
// after Run returns.
//
//	func main() {
//	    if err := xcweaver.Run(context.Background(), app); err != nil {
//	        log.Fatal(err)
//	    }
//	}
func Run[T any, P PointerToMain[T]](ctx context.Context, app func(context.Context, *T) error) error {
	// Register HealthzHandler in the default ServerMux.
	healthzInit.Do(func() {
		http.HandleFunc(HealthzURL, HealthzHandler)
	})

	//add empty lineage to context
	ctx = antipode.InitCtx(ctx)

	bootstrap, err := runtime.GetBootstrap(ctx)
	if err != nil {
		return err
	}
	if !bootstrap.Exists() {
		return runLocal[T, P](ctx, app)
	}
	return runRemote[T, P](ctx, app, bootstrap)
}

func runLocal[T any, _ PointerToMain[T]](ctx context.Context, app func(context.Context, *T) error) error {
	// Read config from SERVICEWEAVER_CONFIG env variable, if non-empty.
	opts := xcweaver.SingleWeaveletOptions{}
	if filename := os.Getenv("SERVICEWEAVER_CONFIG"); filename != "" {
		contents, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("config file: %w", err)
		}
		opts.ConfigFilename = filename
		opts.Config = string(contents)
	}

	regs := codegen.Registered()
	if err := validateRegistrations(regs); err != nil {
		return err
	}

	wlet, err := xcweaver.NewSingleWeavelet(ctx, regs, opts)
	if err != nil {
		return err
	}

	go func() {
		if err := wlet.ServeStatus(ctx); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	main, err := wlet.GetImpl(reflection.Type[T]())
	if err != nil {
		return err
	}
	return app(ctx, main.(*T))
}

func runRemote[T any, _ PointerToMain[T]](ctx context.Context, app func(context.Context, *T) error, bootstrap runtime.Bootstrap) error {
	regs := codegen.Registered()
	if err := validateRegistrations(regs); err != nil {
		return err
	}

	opts := xcweaver.RemoteWeaveletOptions{}
	wlet, err := xcweaver.NewRemoteWeavelet(ctx, regs, bootstrap, opts)
	if err != nil {
		return err
	}

	// Return when either (1) the remote weavelet exits, or (2) the user
	// provided app function returns, whichever happens first.
	errs := make(chan error, 2)
	if wlet.Info().RunMain {
		main, err := wlet.GetImpl(reflection.Type[T]())
		if err != nil {
			return err
		}
		go func() {
			errs <- app(ctx, main.(*T))
		}()
	}
	go func() {
		errs <- wlet.Wait()
	}()
	return <-errs
}

var (
	// Equivalence checks with the struct in internal/xcweaver/types.go.
	_ = WeaverInfo(xcweaver.WeaverInfo{})
	_ = xcweaver.WeaverInfo(WeaverInfo{})
)

// WeaverInfo holds runtime information about a deployed application.
type WeaverInfo struct {
	// Unique identifier for the application deployment.
	DeploymentID string

	// TODO(spetrovic): Add other runtime fields here (e.g., application start
	// time, name of the deployer).
}

// Implements[T] is a type that is be embedded inside a component
// implementation struct to indicate that the struct implements a component of
// type T. For example, consider a Cache component.
//
//	type Cache interface {
//	    Get(ctx context.Context, key string) (string, error)
//	    Put(ctx context.Context, key, value string) error
//	}
//
// A concrete type that implements the Cache component is written as follows:
//
//	type lruCache struct {
//	    xcweaver.Implements[Cache]
//	    ...
//	}
//
// Because Implements is embedded inside the component implementation, methods
// of Implements are available as methods of the component implementation type
// and can be invoked directly. For example, given an instance c of type
// lruCache, we can call c.Logger().
type Implements[T any] struct {
	// Component logger.
	logger *slog.Logger

	weaverInfo *xcweaver.WeaverInfo

	// Given a component implementation type, there is currently no nice way,
	// using reflection, to get the corresponding component interface type [1].
	// The component_interface_type field exists to make it possible.
	//
	// [1]: https://github.com/golang/go/issues/54393.
	//
	//lint:ignore U1000 See comment above.
	component_interface_type T

	// We embed implementsImpl so that component implementation structs
	// implement the Unrouted interface by default but implement the
	// RoutedBy[T] interface when they embed WithRouter[T].
	implementsImpl
}

// Logger returns a logger that associates its log entries with this component.
// Log entries are labeled with any OpenTelemetry trace id and span id in the
// provided context.
func (i Implements[T]) Logger(ctx context.Context) *slog.Logger {
	logger := i.logger
	s := trace.SpanContextFromContext(ctx)
	if s.HasTraceID() {
		logger = logger.With("traceid", s.TraceID().String())
	}
	if s.HasSpanID() {
		logger = logger.With("spanid", s.SpanID().String())
	}
	return logger
}

func (i *Implements[T]) setLogger(logger *slog.Logger) {
	i.logger = logger
}

// Weaver returns runtime information about the deployed application.
func (i Implements[T]) Weaver() WeaverInfo {
	return WeaverInfo(*i.weaverInfo)
}

func (i *Implements[T]) setWeaverInfo(info *xcweaver.WeaverInfo) {
	i.weaverInfo = info
}

// implements is a method that can only be implemented inside the xcweaver
// package. It exists so that a component struct that embeds Implements[T]
// implements the InstanceOf[T] interface.
//
//lint:ignore U1000 implements is used by InstanceOf.
func (Implements[T]) implements(T) {}

// InstanceOf[T] is the interface implemented by a struct that embeds
// xcweaver.Implements[T].
type InstanceOf[T any] interface {
	implements(T)
}

// Ref[T] is a field that can be placed inside a component implementation
// struct. T must be a component type. Service Weaver will automatically
// fill such a field with a handle to the corresponding component.
type Ref[T any] struct {
	value T
}

// Get returns a handle to the component of type T.
func (r Ref[T]) Get() T { return r.value }

// isRef is an internal method that is only implemented by Ref[T] and is
// used internally to check that a value is of type Ref[T].
func (r Ref[T]) isRef() {}

// setRef sets the underlying value of a Ref.
func (r *Ref[T]) setRef(value any) {
	r.value = value.(T)
}

// Listener is a network listener that can be placed as a field inside a
// component implementation struct. Once placed, Service Weaver automatically
// initializes the Listener and makes it suitable for receiving network
// traffic. For example:
//
//	type myComponentImpl struct {
//	    xcweaver.Implements[MyComponent]
//	    myListener      xcweaver.Listener
//	    myOtherListener xcweaver.Listener
//	}
//
// By default, all listeners listen on address ":0". This behavior can be
// modified by passing options for individual listeners in the application
// config. For example, to specify local addresses for the above two listeners,
// the user can add the following lines to the application config file:
//
//	[listeners]
//	myListener      = {local_address = "localhost:9000"}
//	myOtherListener = {local_address = "localhost:9001"}
//
// Listeners are identified by their field names in the component
// implementation structs (e.g., myListener and myOtherListener). If the user
// wishes to assign different names to their listeners, they may do so by
// adding a `xcweaver:"name"` struct tag to their listener fields, e.g.:
//
//	type myComponentImpl struct {
//	    xcweaver.Implements[MyComponent]
//	    myListener      xcweaver.Listener
//	    myOtherListener xcweaver.Listener `xcweaver:"mylistener2"`
//	}
//
// Listener names must be valid Go identifier names. Listener names must be
// unique inside a given application binary, regardless of which components
// they are specified in. For example, it is illegal to declare a Listener
// field "foo" in two different component implementation structs, unless one is
// renamed using the `xcweaver:"name"` struct tag.
//
// HTTP servers constructed using this listener are expected to perform health
// checks on the reserved HealthzURL path. (Note that this URL path is
// configured to never receive any user traffic.)
type Listener struct {
	net.Listener        // underlying listener
	proxyAddr    string // address of proxy that forwards to the listener
}

// String returns the address clients should dial to connect to the listener;
// this will be the proxy address if available, otherwise the <host>:<port> for
// this listener.
func (l Listener) String() string {
	if l.proxyAddr != "" {
		return l.proxyAddr
	}
	return l.Addr().String()
}

// ProxyAddr returns the dialable address of the proxy that forwards traffic to
// this listener, or returns the empty string if there is no such proxy.
func (l *Listener) ProxyAddr() string {
	return l.proxyAddr
}

// WithConfig[T] is a type that can be embedded inside a component
// implementation. The Service Weaver runtime will take per-component
// configuration information found in the application config file and use it to
// initialize the contents of T.
//
// # Example
//
// Consider a cache component where the cache size should be configurable.
// Define a struct that includes the size, associate it with the component
// implementation, and use it inside the component methods.
//
//	type cacheConfig struct
//	    Size int
//	}
//
//	type cache struct {
//	    xcweaver.Implements[Cache]
//	    xcweaver.WithConfig[cacheConfig]
//	    // ...
//	}
//
//	func (c *cache) Init(context.Context) error {
//	    // Use c.Config().Size...
//	    return nil
//	}
//
// The application config file can specify these values as keys under the
// full component path.
//
//	["example.com/mypkg/Cache"]
//	Size = 1000
//
// # Field Names
//
// You can use `toml` struct tags to specify the name that should be used for a
// field in a config file. For example, we can change the cacheConfig struct to
// the following:
//
//	type cacheConfig struct
//	    Size int `toml:"my_custom_name"`
//	}
//
// And change the config file accordingly:
//
//	["example.com/mypkg/Cache"]
//	my_custom_name = 1000
type WithConfig[T any] struct {
	config T
}

// Config returns the configuration information for the component that embeds
// this [xcweaver.WithConfig].
//
// Any fields in T that were not present in the application config file will
// have their default values.
//
// Any fields in the application config file that are not present in T will be
// flagged as an error at application startup.
func (wc *WithConfig[T]) Config() *T {
	return &wc.config
}

// getConfig returns the underlying config.
func (wc *WithConfig[T]) getConfig() any {
	return &wc.config
}

// WithRouter[T] is a type that can be embedded inside a component
// implementation struct to indicate that calls to a method M on the component
// must be routed according to the the value returned by T.M().
//
// # Example
//
// For example, consider a Cache component that maintains an in-memory cache.
//
//	type Cache interface {
//	    Get(ctx context.Context, key string) (string, error)
//	    Put(ctx context.Context, key, value string) error
//	}
//
// We can create a router for the Cache component like this.
//
//	type cacheRouter struct{}
//	func (cacheRouter) Get(_ context.Context, key string) string { return key }
//	func (cacheRouter) Put(_ context.Context, key, value string) string { return key }
//
// To associate a router with its component, embed [xcweaver.WithRouter] in the
// component implementation.
//
//	type lruCache struct {
//		xcweaver.Implements[Cache]
//		xcweaver.WithRouter[cacheRouter]
//	}
//
// For every component method that needs to be routed (e.g., Get and Put), the
// associated router should implement an equivalent method (i.e., same name and
// argument types) whose return type is the routing key. When a component's
// routed method is invoked, its corresponding router method is invoked to
// produce a routing key. Method invocations that produce the same key are
// routed to the same replica.
//
// # Routing Keys
//
// A routing key can be any integer (e.g., int, int32), float (i.e. float32,
// float64), or string; or a struct where all fields are integers, floats, or
// strings. A struct may also embed [AutoMarshal]. For example, the following
// are valid routing keys.
//
//	int
//	int32
//	float32
//	float63
//	string
//	struct{}
//	struct{x int}
//	struct{x int; y string}
//	struct{xcweaver.AutoMarshal; x int; y string}
//
// Every router method must return the same routing key type. The following,
// for example, is invalid:
//
//	// ERROR: Get returns a string, but Put returns an int.
//	func (cacheRouter) Get(_ context.Context, key string) string { return key }
//	func (cacheRouter) Put(_ context.Context, key, value string) int { return 42 }
//
// # Semantics
//
// NOTE that routing is done on a best-effort basis. Service Weaver will try to
// route method invocations with the same key to the same replica, but this is
// not guaranteed. As a corollary, you should never depend on routing for
// correctness. Only use routing to increase performance in the common case.
type WithRouter[T any] struct{}

// routedBy(T) implements the RoutedBy[T] interface.
//
//lint:ignore U1000 routedBy is used by RoutedBy and Unrouted.
func (WithRouter[T]) routedBy(T) {}

// RoutedBy[T] is the interface implemented by a struct that embeds
// xcweaver.RoutedBy[T].
type RoutedBy[T any] interface {
	routedBy(T)
}

// See Implements.implementsImpl.
type implementsImpl struct{}

// See [Unrouted].
type if_youre_seeing_this_you_probably_forgot_to_run_weaver_generate struct{}

// See [Unrouted].
func (implementsImpl) routedBy(if_youre_seeing_this_you_probably_forgot_to_run_weaver_generate) {}

// Unrouted is the interface implemented by instances that don't embed
// xcweaver.WithRouter[T].
type Unrouted interface {
	routedBy(if_youre_seeing_this_you_probably_forgot_to_run_weaver_generate)
}

var _ Unrouted = (*implementsImpl)(nil)

// AutoMarshal is a type that can be embedded within a struct to indicate that
// "xcweaver generate" should generate serialization methods for the struct.
//
// Named struct types are not serializable by default. However, they can
// trivially be made serializable by embedding AutoMarshal. For example:
//
//	type Pair struct {
//	    xcweaver.AutoMarshal
//	    x, y int
//	}
//
// The AutoMarshal embedding instructs "xcweaver generate" to generate
// serialization methods for the struct, Pair in this example.
//
// Note, however, that AutoMarshal cannot magically make any type serializable.
// For example, "xcweaver generate" will raise an error for the following code
// because the NotSerializable struct is fundamentally not serializable.
//
//	// ERROR: NotSerializable cannot be made serializable.
//	type NotSerializable struct {
//	    xcweaver.AutoMarshal
//	    f func()   // functions are not serializable
//	    c chan int // chans are not serializable
//	}
type AutoMarshal struct{}

// TODO(mwhittaker): The following methods have AutoMarshal implement
// codegen.AutoMarshal. Alternatively, we could modify the code generator to
// ignore AutoMarshal during marshaling and unmarshaling.

func (AutoMarshal) WeaverMarshal(*codegen.Encoder)   {}
func (AutoMarshal) WeaverUnmarshal(*codegen.Decoder) {}

type NotRetriable interface{}

var (
	ErrNotFound = errors.New("key not found")
)

type Antipode struct {
	Datastore_type antipode.Datastore_type
	Datastore_ID   string
}

type AntipodeObject struct {
	Version string
	Lineage []byte
}

func (a Antipode) String() string {
	return a.Datastore_ID
}

func (a Antipode) Write(ctx context.Context, table string, key string, value string) (context.Context, error) {

	return antipode.Write(ctx, a.Datastore_type, a.Datastore_ID, table, key, value)
}

func (a Antipode) Read(ctx context.Context, table string, key string) (string, []byte, error) {

	value, line, err := antipode.Read(ctx, a.Datastore_type, table, key)
	if err == antipode.ErrNotFound {
		return "", []byte{}, ErrNotFound
	} else if err != nil {
		return "", []byte{}, err
	}

	lineageBytes, err := json.Marshal(line)
	if err != nil {
		return "", []byte{}, err
	}

	return value, lineageBytes, nil
}

func (a Antipode) Consume(ctx context.Context, exchange string, key string, stop chan struct{}) (<-chan AntipodeObject, error) {

	msgs, err := antipode.Consume(ctx, a.Datastore_type, exchange, key, stop)
	if err != nil {
		return nil, err
	}

	antipodeObjctsChan := make(chan AntipodeObject)

	go func() {
		defer close(antipodeObjctsChan)
		//requeue non-processed messages
		defer func(<-chan AntipodeObject) {
			for d := range antipodeObjctsChan {
				var lineage []antipode.WriteIdentifier
				err := json.Unmarshal(d.Lineage, &lineage)
				if err != nil {
					return
				}
				obj := antipode.AntiObj{d.Version, lineage}
				err = antipode.Requeue(ctx, a.Datastore_type, key, obj)
				if err != nil {
					return
				}
			}
		}(antipodeObjctsChan)
		select {
		case <-stop:
			return
		default:
			for d := range msgs {
				lineageBytes, err := json.Marshal(d.Lineage)
				if err != nil {
					fmt.Println(err.Error())
				}
				antiObj := AntipodeObject{d.Version, lineageBytes}
				antipodeObjctsChan <- antiObj
			}
		}

	}()

	return antipodeObjctsChan, nil
}

func (a Antipode) Barrier(ctx context.Context) error {

	return antipode.Barrier(ctx, a.Datastore_type, a.Datastore_ID)
}

func Transfer(ctx context.Context, lineage []byte) (context.Context, error) {

	var line []antipode.WriteIdentifier
	err := json.Unmarshal(lineage, &line)
	if err != nil {
		return nil, err
	}

	return antipode.Transfer(ctx, line)
}

func GetLineage(ctx context.Context) ([]byte, error) {

	line, err := antipode.GetLineage(ctx)
	if err != nil {
		return []byte{}, err
	}

	lineageBytes, err := json.Marshal(line)
	if err != nil {
		return []byte{}, err
	}

	return lineageBytes, nil
}
