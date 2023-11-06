package obstasks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"net.kopias.oscbridge/app/drivers/obsremote"

	"net.kopias.oscbridge/app/drivers/paramsanitizer"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IActionTask = &SceneChanger{}

const (
	ParamSceneTargetKey     = "target"
	ParamSceneTargetProgram = "program"
	ParamSceneTargetPreview = "preview"
	ParamSceneKey           = "scene"
	ParamSceneMatchTypeKey  = "scene_match_type"
	ParamSceneMatchExact    = "exact"
	ParamSceneMatchRegexp   = "regexp"
)

// SceneChanger allows for changing preview/program scenes.
type SceneChanger struct {
	obsConnections map[string]*obsremote.OBSRemote

	cachedSceneList []string

	sceneName      string
	scenePattern   *string
	debug          bool
	log            usecaseifs.ILogger
	configError    error
	sceneTarget    string
	connectionName string
}

func NewSceneChanger(obsConnections map[string]*obsremote.OBSRemote, log usecaseifs.ILogger, debug bool) usecaseifs.IActionTask {
	return &SceneChanger{obsConnections: obsConnections, log: log, debug: debug, cachedSceneList: []string{}}
}

func (o *SceneChanger) Validate() error {
	return o.configError
}

func (o *SceneChanger) Execute(ctx context.Context, store usecaseifs.IMessageStore) error {
	o.log.Infof(ctx, "\tExecuting task: obs scene change")

	var err error
	sceneName := o.sceneName

	if len(o.cachedSceneList) == 0 {
		o.cachedSceneList, err = o.obsConnections[o.connectionName].ListScenes(ctx)
		if err != nil {
			return fmt.Errorf("failed to list scenes: %w", err)
		}
	}

	if o.scenePattern != nil {
		sceneName, err = o.getSceneNameByRegexp(*o.scenePattern)
		if err != nil {
			return err
		}
	}

	if o.sceneTarget == ParamSceneTargetPreview {
		err := o.obsConnections[o.connectionName].SwitchPreviewScene(ctx, sceneName)
		if err != nil {
			return fmt.Errorf("failed to change preview scene to %s: %w", sceneName, err)
		}
	}

	if o.sceneTarget == ParamSceneTargetProgram {
		err := o.obsConnections[o.connectionName].SwitchProgramScene(ctx, sceneName)
		if err != nil {
			return fmt.Errorf("failed to change program scene to %s: %w", sceneName, err)
		}
	}

	return nil
}

func (o *SceneChanger) SetParameters(m map[string]interface{}) {
	sanitized, err := paramsanitizer.SanitizeParams(m, []paramsanitizer.ParameterDefinition{
		{
			Name:     ParamSceneKey,
			Optional: false,
			Type:     []string{"string"},
		}, {
			Name:     ParamConnectionKey,
			Optional: false,
			Type:     []string{"string"},
		}, {
			Name:         ParamSceneMatchTypeKey,
			Optional:     true,
			DefaultValue: ParamSceneMatchExact,
			ValuePattern: fmt.Sprintf("^%s|%s$", ParamSceneMatchExact, ParamSceneMatchRegexp),
			Type:         []string{"string"},
		}, {
			Name:         ParamSceneTargetKey,
			Optional:     false,
			ValuePattern: fmt.Sprintf("^%s|%s$", ParamSceneTargetProgram, ParamSceneTargetPreview),
			Type:         []string{"string"},
		},
	})
	if err != nil {
		o.configError = fmt.Errorf("failed to verify parameters: %w", err)
		return
	}

	// nolint:forcetypeassert
	o.sceneName = sanitized[ParamSceneKey].(string)

	// nolint:forcetypeassert
	o.connectionName = sanitized[ParamConnectionKey].(string)

	// nolint:forcetypeassert
	o.sceneTarget = sanitized[ParamSceneTargetKey].(string)

	if sanitized[ParamSceneMatchTypeKey] == ParamSceneMatchRegexp {
		// nolint:forcetypeassert
		scp := sanitized[ParamSceneKey].(string)
		_, err = regexp.Compile(scp)
		if err != nil {
			o.configError = fmt.Errorf("%s is not a valid regexp: %w", ParamSceneMatchTypeKey, err)
			return
		}
		o.scenePattern = &scp
	}
}

func (o *SceneChanger) getSceneNameByRegexp(p string) (string, error) {
	r, err := regexp.Compile(p)
	if err != nil {
		return "", err
	}
	for _, scene := range o.cachedSceneList {
		if r.MatchString(scene) {
			return scene, nil
		}
	}

	return "", fmt.Errorf("no scene matched '%s' (current scenes: %s)", p, strings.Join(o.cachedSceneList, ", "))
}
