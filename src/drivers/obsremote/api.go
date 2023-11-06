package obsremote

import (
	"context"
	"fmt"

	"github.com/andreykaipov/goobs/api/requests/general"

	"github.com/andreykaipov/goobs/api/requests/ui"

	"github.com/andreykaipov/goobs/api/requests/record"
	"github.com/andreykaipov/goobs/api/requests/stream"

	"github.com/andreykaipov/goobs/api/requests/scenes"

	"github.com/andreykaipov/goobs/api/typedefs"
	"net.kopias.oscbridge/app/pkg/slicetools"
)

func (or *OBSRemote) ListScenes(ctx context.Context) ([]string, error) {
	if or.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	or.m.Lock()
	defer or.m.Unlock()

	list, err := or.client.Scenes.GetSceneList()
	if err != nil {
		return nil, fmt.Errorf("failed to list scenes: %w", err)
	}
	sceneNames := slicetools.Map(list.Scenes, func(t *typedefs.Scene) string {
		return t.SceneName
	})

	return sceneNames, nil
}

func (or *OBSRemote) SwitchPreviewScene(ctx context.Context, sceneName string) error {
	if or.client == nil {
		return fmt.Errorf("not connected")
	}

	or.m.Lock()
	defer or.m.Unlock()

	params := &scenes.SetCurrentPreviewSceneParams{
		SceneName: sceneName,
	}

	_, err := or.client.Scenes.SetCurrentPreviewScene(params)
	if err != nil {
		return fmt.Errorf("failed to switch scene: %w", err)
	}
	return nil
}

func (or *OBSRemote) SwitchProgramScene(ctx context.Context, sceneName string) error {
	if or.client == nil {
		return fmt.Errorf("not connected")
	}

	or.m.Lock()
	defer or.m.Unlock()

	params := &scenes.SetCurrentProgramSceneParams{
		SceneName: sceneName,
	}
	_, err := or.client.Scenes.SetCurrentProgramScene(params)
	if err != nil {
		return fmt.Errorf("failed to switch scene: %w", err)
	}
	return nil
}

func (or *OBSRemote) GetCurrentProgramScene(ctx context.Context) (string, error) {
	if or.client == nil {
		return "", fmt.Errorf("not connected")
	}

	or.m.Lock()
	defer or.m.Unlock()

	sme, err := or.client.Ui.GetStudioModeEnabled(&ui.GetStudioModeEnabledParams{})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve studio mode state: %w", err)
	}
	if !sme.StudioModeEnabled {
		return "", nil
	}

	params := &scenes.GetCurrentProgramSceneParams{}
	r, err := or.client.Scenes.GetCurrentProgramScene(params)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve current program scene: %w", err)
	}
	return r.CurrentProgramSceneName, nil
}

func (or *OBSRemote) GetCurrentPreviewScene(ctx context.Context) (string, error) {
	if or.client == nil {
		return "", fmt.Errorf("not connected")
	}
	or.m.Lock()
	defer or.m.Unlock()

	params := &scenes.GetCurrentPreviewSceneParams{}
	r, err := or.client.Scenes.GetCurrentPreviewScene(params)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve current preview scene: %w", err)
	}
	return r.CurrentPreviewSceneName, nil
}

func (or *OBSRemote) IsStreaming(ctx context.Context) (bool, error) {
	if or.client == nil {
		return false, fmt.Errorf("not connected")
	}

	or.m.Lock()
	defer or.m.Unlock()

	params := &stream.GetStreamStatusParams{}
	r, err := or.client.Stream.GetStreamStatus(params)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve stream status: %w", err)
	}
	return r.OutputActive, nil
}

func (or *OBSRemote) IsRecording(ctx context.Context) (bool, error) {
	if or.client == nil {
		return false, fmt.Errorf("not connected")
	}

	or.m.Lock()
	defer or.m.Unlock()

	params := &record.GetRecordStatusParams{}
	r, err := or.client.Record.GetRecordStatus(params)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve recording status: %w", err)
	}

	return r.OutputActive, nil
}

func (or *OBSRemote) VendorRequest(ctx context.Context, vendorName string, requestType string, requestData interface{}) (responseData interface{}, err error) {
	if or.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	or.m.Lock()
	defer or.m.Unlock()

	params := &general.CallVendorRequestParams{
		RequestData: requestData,
		RequestType: requestType,
		VendorName:  vendorName,
	}
	r, err := or.client.General.CallVendorRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to send vendor request: %w", err)
	}

	return r.ResponseData, nil
}
