package obstasks

import (
	"context"
	"fmt"

	"net.kopias.oscbridge/app/drivers/obsremote"

	"net.kopias.oscbridge/app/drivers/paramsanitizer"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IActionTask = &VendorRequest{}

const (
	ParamVendorName  = "vendorName"
	ParamRequestType = "requestType"
	ParamRequestData = "requestData"
)

// VendorRequest allows sending arbitrary "vendor" messages.
type VendorRequest struct {
	obsConnections map[string]*obsremote.OBSRemote

	debug          bool
	log            usecaseifs.ILogger
	configError    error
	vendorName     string
	requestType    string
	requestData    map[string]interface{}
	connectionName string
}

func NewVendorRequest(obsConnections map[string]*obsremote.OBSRemote, log usecaseifs.ILogger, debug bool) usecaseifs.IActionTask {
	return &VendorRequest{obsConnections: obsConnections, log: log, debug: debug}
}

func (o *VendorRequest) Validate() error {
	return o.configError
}

func (o *VendorRequest) Execute(ctx context.Context, store usecaseifs.IMessageStore) error {
	o.log.Infof(ctx, "\tExecuting task: OBS vendor request")

	if _, err := o.obsConnections[o.connectionName].VendorRequest(ctx, o.vendorName, o.requestType, o.requestData); err != nil {
		return fmt.Errorf("failed to execute vendor request: %w", err)
	}
	return nil
}

func (o *VendorRequest) SetParameters(m map[string]interface{}) {
	sanitized, err := paramsanitizer.SanitizeParams(m, []paramsanitizer.ParameterDefinition{
		{
			Name:     ParamVendorName,
			Optional: false,
			Type:     []string{"string"},
		}, {
			Name:     ParamRequestType,
			Optional: false,
			Type:     []string{"string"},
		}, {
			Name:     ParamRequestData,
			Optional: false,
			Type:     []string{"map[string]interface {}"},
		}, {
			Name:     ParamConnectionKey,
			Optional: false,
			Type:     []string{"string"},
		},
	})
	if err != nil {
		o.configError = fmt.Errorf("failed to verify parameters: %w", err)
		return
	}

	// nolint:forcetypeassert
	o.vendorName = sanitized[ParamVendorName].(string)

	// nolint:forcetypeassert
	o.requestType = sanitized[ParamRequestType].(string)

	// nolint:forcetypeassert
	o.requestData = sanitized[ParamRequestData].(map[string]interface{})

	// nolint:forcetypeassert
	o.connectionName = sanitized[ParamConnectionKey].(string)
}
