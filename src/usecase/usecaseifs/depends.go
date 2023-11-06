package usecaseifs

import (
	"context"
	"time"
)

// Adapters are plugins converting to and from the usecases, they are usually found in the "adapters" folder,
// implementing the following interfaces.
type (
	// IConfiguration determines the used configuration values & methods by the use-cases.
	IConfiguration interface {
		ShouldDebugOSCConditions() bool
	}

	// ILogger specifies an interface for general logging.
	ILogger interface {
		Debugf(ctx context.Context, message string, args ...interface{})
		Debug(ctx context.Context, message string)
		Infof(ctx context.Context, message string, args ...interface{})
		Info(ctx context.Context, message string)
		Warnf(ctx context.Context, message string, args ...interface{})
		Warn(ctx context.Context, message string)
		Fatalf(ctx context.Context, message error, args ...interface{})
		Fatal(ctx context.Context, message error)
		Errorf(ctx context.Context, message string, args ...interface{})
		Error(ctx context.Context, message string)
		Err(ctx context.Context, err error)
		Messsage(ctx context.Context, level string, message string)
	}

	IOSCConnection interface {
		Start(context.Context) error
		Stop(context.Context)
		Notify() <-chan error
		GetEventChan(ctx context.Context) <-chan IOSCMessage
		SendMessage(ctx context.Context, msg IOSCMessage) error
	}

	IOSCMessage interface {
		Equal(msg IOSCMessage) bool
		GetAddress() string
		GetArguments() []IOSCMessageArgument
		String() string
	}

	IOSCMessageArgument interface {
		GetType() string
		GetValue() string
		String() string
	}

	IEventPublisher interface {
		PublishEvent(ctx context.Context, e IOSCMessage) error
	}
)

type (
	IOBSRemote interface {
		ListScenes(ctx context.Context) ([]string, error)
		SwitchPreviewScene(ctx context.Context, sceneName string) error
		SwitchProgramScene(ctx context.Context, sceneName string) error
		GetCurrentProgramScene(ctx context.Context) (string, error)
		GetCurrentPreviewScene(ctx context.Context) (string, error)
		IsStreaming(ctx context.Context) (bool, error)
		IsRecording(ctx context.Context) (bool, error)
		VendorRequest(ctx context.Context, vendorName string, requestType string, requestData interface{}) (responseData interface{}, err error)
	}
)

type (
	IMessageStore interface {
		Clone() IMessageStore
		GetAll() map[string]IMessageStoreRecord
		GetRecord(key string, trackAccess bool) (IMessageStoreRecord, bool)
		GetOneRecordByRegexp(re string, trackAccess bool) (IMessageStoreRecord, error)
		GetRecordsByRegexp(re string, trackAccess bool) ([]IMessageStoreRecord, error)
		GetRecordsByPrefix(prefix string, trackAccess bool) []IMessageStoreRecord
		SetRecord(msg IOSCMessage) (updated bool)
		WatchRecordAccess(msg *IOSCMessage)
		GetWatchedRecordAccesses() int64
	}

	IMessageStoreRecord interface {
		GetMessage() IOSCMessage
		GetArrivedAt() time.Time
	}

	ActionTaskFactory func() IActionTask

	IActionTask interface {
		Execute(ctx context.Context, store IMessageStore) error
		SetParameters(map[string]interface{})
		Validate() error
	}

	IAction interface {
		GetName() string
		Evaluate(ctx context.Context, store IMessageStore) (bool, error)
		Execute(ctx context.Context, store IMessageStore) error
		GetDebounceMillis() int64
	}

	ActionConditionFactory func(path string) IActionCondition

	IActionCondition interface {
		// @TODO notsure if these are right here
		SetParameters(map[string]interface{})
		AddChild(condition IActionCondition)

		Evaluate(ctx context.Context, store IMessageStore) (bool, error)
		Validate() error
	}
)
