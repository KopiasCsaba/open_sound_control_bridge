package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"net.kopias.oscbridge/app/drivers/osc_message"

	"net.kopias.oscbridge/app/pkg/filetools"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ iusecase = &oscMessageStoreManager{}

type serializedMessage struct {
	Address   string
	Arguments []serializedArgument
}
type serializedArgument struct {
	Type  string
	Value string
}

// oscMessageStoreManager is resposnible for managing store updates and state.
type oscMessageStoreManager struct {
	ucs *UseCases
	log usecaseifs.ILogger
	cfg usecaseifs.IConfiguration

	store            usecaseifs.IMessageStore
	storeVersion     int
	actions          []usecaseifs.IAction
	storePersistPath string
	notify           chan error
	quit             chan interface{}
}

func newOscMessageStoreManager(
	log usecaseifs.ILogger,
	cfg usecaseifs.IConfiguration,
	store usecaseifs.IMessageStore,
	actions []usecaseifs.IAction,
	storePersistPath string,
) *oscMessageStoreManager {
	return &oscMessageStoreManager{
		log:              log,
		cfg:              cfg,
		actions:          actions,
		store:            store,
		storeVersion:     0,
		storePersistPath: storePersistPath,
		notify:           make(chan error, 1),
		quit:             make(chan interface{}, 1),
	}
}

func (e *oscMessageStoreManager) Start(ctx context.Context) error {
	if e.storePersistPath != "" {
		if err := e.loadJSONDump(ctx); err != nil {
			return fmt.Errorf("failed to load persistence file: %w", err)
		}
	}

	go e.jsonSync(ctx)
	return nil
}

// jsonSync dumps the store into a json file every sync.
func (e *oscMessageStoreManager) jsonSync(ctx context.Context) {
	if e.storePersistPath == "" {
		return
	}

	lastStoreVersion := e.storeVersion
	for {
		select {
		case <-e.quit:
			e.log.Info(ctx, "Message store JSON syncing quiting...")
			return
		default:
		}

		if lastStoreVersion != e.storeVersion {
			e.dumpStoreToJSON(ctx)
			lastStoreVersion = e.storeVersion
		}
		time.Sleep(1 * time.Second)
	}
}

func (e *oscMessageStoreManager) setUseCases(ucs *UseCases) {
	e.ucs = ucs
}

func (e *oscMessageStoreManager) updateRecord(ctx context.Context, msg usecaseifs.IOSCMessage) {
	if e.store.SetRecord(msg) {
		e.storeVersion++

		e.log.Infof(ctx, "Store updated with: %v", msg)
		go e.evaluateActions(ctx, msg)
	}
}

func (e *oscMessageStoreManager) evaluateActions(ctx context.Context, latestUpdatedMessage usecaseifs.IOSCMessage) {
	ctx = getTaskExecutionSessionContext(ctx)
	currentStore := e.store.Clone()
	currentStore.WatchRecordAccess(&latestUpdatedMessage)

	if e.cfg.ShouldDebugOSCConditions() {
		e.log.Info(ctx, "Evaluating actions because of a change in the osc message store.")
	}
	for _, action := range e.actions {
		e.evaluateAction(ctx, action, currentStore)
	}

	e.log.Info(ctx, "finished.")
}

func (e *oscMessageStoreManager) evaluateAction(ctx context.Context, action usecaseifs.IAction, currentStore usecaseifs.IMessageStore) {
	if e.cfg.ShouldDebugOSCConditions() {
		e.log.Infof(ctx, "Evaluating action: %s", action.GetName())
	}
	matched, err := action.Evaluate(ctx, currentStore)
	if err != nil {
		e.log.Err(ctx, fmt.Errorf("error during evaluation of %s: %w", action.GetName(), err))
	}

	if !matched {
		return
	}

	if currentStore.GetWatchedRecordAccesses() == 0 {
		e.log.Infof(ctx, "Although the triggers matched, none of them selected the newly changed record, therefore skipping execution.")
		return
	}

	if action.GetDebounceMillis() != 0 {
		time.Sleep(time.Duration(action.GetDebounceMillis()) * time.Millisecond)
		matched, err = action.Evaluate(ctx, currentStore)
		if err != nil {
			e.log.Err(ctx, fmt.Errorf("error during evaluation of %s: %w", action.GetName(), err))
		}

		if e.cfg.ShouldDebugOSCConditions() {
			e.log.Infof(ctx, "After debouncing for %d ms, the result is: %t", action.GetDebounceMillis(), matched)
		}

		if !matched {
			return
		}
	}

	e.log.Infof(ctx, "Executing action: %s", action.GetName())
	if err := action.Execute(ctx, currentStore); err != nil {
		e.log.Err(ctx, err)
	}
}

// Notify returns the notification channel that can be used to listen for the client's exit
func (e *oscMessageStoreManager) Notify() <-chan error {
	return e.notify
}

func (e *oscMessageStoreManager) Stop(ctx context.Context) {
	e.quit <- true
}

func (e *oscMessageStoreManager) dumpStoreToJSON(ctx context.Context) {
	serialized := []serializedMessage{}
	for _, record := range e.store.GetAll() {
		if record.GetMessage() == nil {
			continue
		}

		msg := serializedMessage{
			Address:   record.GetMessage().GetAddress(),
			Arguments: []serializedArgument{},
		}

		for _, argument := range record.GetMessage().GetArguments() {
			msg.Arguments = append(msg.Arguments, serializedArgument{
				Type:  argument.GetType(),
				Value: argument.GetValue(),
			})
		}
		serialized = append(serialized, msg)
	}

	jsonData, err := json.MarshalIndent(serialized, "", "  ")
	if err != nil {
		e.log.Err(ctx, fmt.Errorf("failed to marshal store: %w", err))
		return
	}
	// Create or open the file for writing
	file, err := os.Create(e.storePersistPath)
	if err != nil {
		e.log.Err(ctx, fmt.Errorf("failed to create json dump to %s: %w", e.storePersistPath, err))
		return
	}
	defer file.Close() // Close the file when we're done

	// Write the JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		e.log.Err(ctx, fmt.Errorf("failed to write json dump to %s: %w", e.storePersistPath, err))
		return
	}
}

func (e *oscMessageStoreManager) loadJSONDump(ctx context.Context) error {
	unserialized := []serializedMessage{}

	if !filetools.FileExists(e.storePersistPath) {
		e.log.Infof(ctx, "Not loading persistence file %s: it does not exist", e.storePersistPath)
		return nil
	}
	content, err := os.ReadFile(e.storePersistPath)
	if err != nil {
		return fmt.Errorf("error when opening file: %s %w", e.storePersistPath, err)
	}

	if err = json.Unmarshal(content, &unserialized); err != nil {
		return fmt.Errorf("error when parsing file: %s %w", e.storePersistPath, err)
	}

	for _, msg := range unserialized {
		args := []usecaseifs.IOSCMessageArgument{}
		for _, argument := range msg.Arguments {
			args = append(args, osc_message.NewMessageArgument(argument.Type, argument.Value))
		}
		e.store.SetRecord(osc_message.NewMessage(msg.Address, args))
	}
	e.log.Infof(ctx, "loaded persistence file %s.", e.storePersistPath)
	return nil
}
