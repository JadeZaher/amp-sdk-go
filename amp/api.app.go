package amp

import (
	"github.com/art-media-platform/amp-sdk-go/stdlib/media"
	"github.com/art-media-platform/amp-sdk-go/stdlib/tag"
	"github.com/art-media-platform/amp-sdk-go/stdlib/task"
)

// App is how an app module is registered with an amp.Host so it can be invoked.~
//
// An App is invoked by a client or other app via the app's Tag or URI.
type App struct {

	// tag.Spec identifies this app with form "amp.app.{PublisherID}.{FamilyID}.{AppNameID}" -- e.g. "amp.app.filesys.posix"
	//   - PublisherID: typically the domain name of the publisher of this app -- e.g. "artspace.systems"
	//   - FamilyID:    encompassing namespace ID used to group related apps (no spaces or punctuation)
	//   - AppNameID:   identifies this app within its parent family and domain (no spaces or punctuation)
	//
	AppSpec      tag.Spec // Universally unique and persistent ID for this module (and the module's "home" space if present)
	Desc         string   // Human-readable description of this app
	Version      string   // "v{MajorVers}.{MinorID}.{RevID}"
	Dependencies []tag.ID // Module Tags this app may access
	Invocations  []string // Additional aliases that invoke this app

	// NewAppInstance is the entry point for an App.
	// Called when an App is first invoked on an active User session and is not yet running.
	// Blocks minimally and returns quickly.
	NewAppInstance func(ctx AppContext) (AppInstance, error)
}

// AppContext is provided by the amp runtime to an AppInstance for support and context.
type AppContext interface {
	task.Context      // Allows select{} for graceful handling of app shutdown
	media.Publisher   // Allows an app to publish assets for client consumption
	Session() Session // Access to underlying Session

	// Returns the absolute file system path of the app's local read-write directory.
	// This directory is scoped by App.AppSpec
	LocalDataPath() string

	// Gets the named attribute from the user's home space -- used high-level app settings.
	// The attr is scoped by both the app Tag so key collision with other users or apps is not possible.
	// This is how an app can store and retrieve its settings for the current user.
	GetAppAttr(attrSpec tag.ID, dst tag.Value) error

	// Write analog for GetAppAttr()
	PutAppAttr(attrSpec tag.ID, src tag.Value) error
}

// Pinner is characterized by the ability to emit Pins.
type Pinner interface {

	// Creates or finds Pin for the given request.
	ServeRequest(req Requester) (Pin, error)
}

// AppInstance is implemented by an App and invoked by amp.Host responding to a client pin request.
type AppInstance interface {
	AppContext
	Pinner

	// Validates a request and performs any needed setup.
	// This åis a chance for an app to perform operations such refreshing an auth token.
	// Following this call, ServeRequest() is called.
	MakeReady(req Requester) error

	// Called exactly once when this App closes
	OnClosing()
}

// Pin is a attribute state connection to an amp.App.
// The handling App is responsible for updating the Requester with state changes as requested.
type Pin interface {
	Pinner

	// Apps start a Pin as a child Context of amp.AppContext.Context or as a child of another Pin.
	// This means an AppContext contains all its Pins and thus Close() will close all Pins (and child requests).
	// This is used to know if a request is still being served and to close it if needed.
	Context() task.Context
}

// TxMsg is workhorse generic transport serialization sent between client and host.
type TxMsg struct {
	TxEnvelope        // public fields and routing tags
	Ops        []TxOp // operations to perform on the target
	OpsSorted  bool   // describes order of []Ops
	DataStore  []byte // stores serialized TxOp data
	refCount   int32  // see AddRef() / ReleaseRef()
}

// ElementID is a multi-part LSM key: CellID / AttrID / SI
type ElementID [3]tag.ID

// TxOpID is TxOp atomic edit entry ID, functioning as a multi-part LSM key: CellID / AttrID / SI / EditID.
type TxOpID struct {
	CellID tag.ID // target cell ID
	AttrID tag.ID // attribute specification ID that expresses or implies operation
	ItemID tag.ID // user-defined value / index / tag ID
	EditID tag.ID // references the previous revision -- see tag.ForkEdit()
}

// TxOp is an atomic operation and is a unit of change (or message).
type TxOp struct {
	TxOpID           // applicable cell, attribute, element, and edit IDs
	OpCode  TxOpCode // operation to perform
	DataLen uint64   // length of data in TxMsg.DataStore
	DataOfs uint64   // offset into TxMsg.DataStore
}

type AttrDef struct {
	tag.Spec
	Prototype tag.Value
}
