package amp

import (
	"net/url"

	"github.com/amp-space/amp-sdk-go/stdlib/task"
)

// Host allows app and transport services to be attached.
// Child processes attach as it responds to client requests to "pin" cells via URLs.
type Host interface {
	task.Context

	// Offers Go runtime and package level access to this Host's primary symbol and amp.App registry.
	// The amp.Registry interface bakes security and efficiently and tries to serve as effective package manager.
	HostRegistry() Registry

	// StartNewSession creates a new HostSession and binds its Msg transport to a stream.
	StartNewSession(parent HostService, via Transport) (HostSession, error)
}

// Transport wraps a Msg transport abstraction, allowing a Host to connect over any data transport layer.
// For example, a tcp-based transport as well as a dll-based transport are both implemented..
type Transport interface {

	// Describes this transport for logging and debugging.
	Label() string

	// Called when this stream should close because the associated parent host session is closing or has closed.
	Close() error

	// SendSTx sends a Msg to the remote client.
	// ErrStreamClosed is used to denote normal stream close.
	SendTx(tx *TxMsg) error

	// RecvTx blocks until it receives a Msg or the stream is done.
	// ErrStreamClosed is used to denote normal stream close.
	RecvTx() (*TxMsg, error)
}

// HostService attaches to a amp.Host as a child, extending host functionality.
type HostService interface {
	task.Context

	// StartService attaches a child task to a Host and starts this HostService.
	// This service may retain the amp.Host instance so that it can make calls to StartNewSession().
	StartService(on Host) error

	// GracefulStop initiates a polite stop of this extension and blocks until it's in a "soft" closed state,
	//    meaning that its service has effectively stopped but its Context is still open.
	// Note this could any amount of time (e.g. until all open requests are closed)
	// Typically, GracefulStop() is called (blocking) and then Context.Close().
	// To stop immediately, Context.Close() is always available.
	GracefulStop()
}

// HostSession in an open client session with a Host.
// Closing is initiated via task.Context.Close().
type HostSession interface {
	task.Context // Underlying task context
	Registry     // How symbols and types registered and resolved

	// Returns the running AssetPublisher instance for this session.
	AssetPublisher() AssetPublisher

	// Returns info about this user and session
	LoginInfo() Login

	// Sends a readied Msg to the client for handling.
	// If msg.ReqID == 0, the attr is sent to the client's session controller (for sending session meta messages).
	// On exit, the given msg should not be referenced further.
	SendTx(tx *TxMsg) error

	// PinCell resolves and pins a requested cell.
	PinCell(req PinOp) (PinContext, error)

	// Gets the currently running AppInstance for an AppID.
	// If the requested app is not running and autoCreate is set, a new instance is created and started.
	GetAppInstance(appID UID, autoCreate bool) (AppInstance, error)
}

// Registry is where apps and types are registered -- concurrency safe.
type Registry interface {

	// Registers an element value type (ElemVal) as a prototype under its pure scalar element type name (also a valid TagSpec type expression).
	// If an entry already exists (common for a type used by multiple apps), then this is a no-op.
	// if registerAs == "", then the prototype.ElemTypeName() is used.
	RegisterPrototype(registerAs string, prototype ElemVal) (AttrID, error)

	// Imports all the types and apps from another registry.
	// When a HostSession is created, its registry starts by importing the Host's registry.
	Import(other Registry) error

	// Registers an app by its UUID, URI, and schemas it supports.
	RegisterApp(app *App) error

	// Looks-up an app by UUID -- READ ONLY ACCESS
	GetAppByUID(appUID UID) (*App, error)

	// Selects the app that best matches an invocation string.
	GetAppForInvocation(invocation string) (*App, error)

	// Registers a block of symbol, attr, cell, and selector definitions for a client.
	RegisterDefs(defs *RegisterDefs) error

	// Instantiates an attr element value for a given attr UID -- typically followed by ElemVal.Unmarshal()
	NewAttrElem(attrID AttrID) (ElemVal, error)
}

// PinContext wraps a client request to receive a cell's state / updates.
type PinContext interface {
	task.Context // Started as a CHILD of the amp.PinnedCell returned by AppInstance.PinCell()

	Op() PinOp // Originating request info

	// Marshals a TxOp and optional value to the given Tx's data store.
	//
	// If the given attr is not enabled within this PinContext, this function is a no-op.
	MarshalTxOp(dst *TxMsg, op TxOp, val ElemVal)

	// PushTx pushes the given tx to this PinContext
	PushTx(tx *TxMsg) error

	// App returns the resolved AppContext that is servicing this PinContext
	App() AppContext
}

// PinOp is a client request to pin a cell.
type PinOp interface {
	RawRequest() PinRequest
	URL() *url.URL
	ContextID() TimeID
}
