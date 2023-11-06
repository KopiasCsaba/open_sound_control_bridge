package httpreq

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"net.kopias.oscbridge/app/drivers/paramsanitizer"

	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ usecaseifs.IActionTask = &HTTPRequest{}

// HTTPRequest makes a HTTP request as configured.
type HTTPRequest struct {
	log         usecaseifs.ILogger
	debug       bool
	configError error
	url         string
	body        string
	method      string
	headers     []string
	timeoutSecs int
}

const (
	ParamURLKey         = "url"
	ParamBodyKey        = "body"
	ParamMethodKey      = "method"
	ParamMethodPost     = "POST"
	ParamMethodGet      = "GET"
	ParamHeadersKey     = "headers"
	ParamTimeoutSecsKey = "timeout_secs"
)

func NewFactory(log usecaseifs.ILogger, debug bool) usecaseifs.ActionTaskFactory {
	return func() usecaseifs.IActionTask { return &HTTPRequest{log: log, debug: debug} }
}

func (o *HTTPRequest) SetParameters(m map[string]interface{}) {
	sanitized, err := paramsanitizer.SanitizeParams(m, []paramsanitizer.ParameterDefinition{
		{
			Name:     ParamURLKey,
			Optional: false,
			Type:     []string{"string"},
		}, {
			Name:         ParamBodyKey,
			Optional:     true,
			DefaultValue: "",
			Type:         []string{"string"},
		}, {
			Name:         ParamTimeoutSecsKey,
			Optional:     true,
			DefaultValue: 30,
			Type:         []string{"int"},
		}, {
			Name:         ParamMethodKey,
			Optional:     true,
			DefaultValue: ParamMethodGet,
			ValuePattern: fmt.Sprintf("(?i)^%s|%s$", ParamMethodPost, ParamMethodGet),
			Type:         []string{"string"},
		}, {
			Name:         ParamHeadersKey,
			Optional:     true,
			DefaultValue: []interface{}{},
			Type:         []string{"[]interface {}"},
		},
	})
	if err != nil {
		o.configError = fmt.Errorf("failed to verify parameters: %w", err)
		return
	}

	// nolint:forcetypeassert
	o.url = sanitized[ParamURLKey].(string)

	// nolint:forcetypeassert
	o.timeoutSecs = sanitized[ParamTimeoutSecsKey].(int)

	// nolint:forcetypeassert
	o.body = sanitized[ParamBodyKey].(string)

	// nolint:forcetypeassert
	o.method = sanitized[ParamMethodKey].(string)

	// nolint:forcetypeassert
	headerSlice := sanitized[ParamHeadersKey].([]interface{})

	o.headers = []string{}
	for i, v := range headerSlice {
		vString, ok := v.(string)
		if !ok {
			o.configError = fmt.Errorf("failed to convert header[%d] to string (from %T)", i, v)
			return
		}
		o.headers = append(o.headers, vString)
	}
}

func (o *HTTPRequest) Validate() error {
	return o.configError
}

func (o *HTTPRequest) Execute(ctx context.Context, store usecaseifs.IMessageStore) error {
	o.log.Infof(ctx, "\tExecuting task: HTTP request")

	// Configure request options
	headersToSend := map[string]string{}
	var parts []string
	for i, header := range o.headers {
		parts = strings.SplitN(header, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("failed to parse header[%d]: '%s' invalid header definition", i, header)
		}
		headersToSend[parts[0]] = parts[1]
	}

	reqCtx, cncl := context.WithTimeout(ctx, time.Second*time.Duration(o.timeoutSecs))
	defer cncl()

	// Create a request
	req, err := http.NewRequestWithContext(reqCtx, strings.ToUpper(o.method), o.url, bytes.NewBufferString(o.body))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Add custom headers
	for key, value := range headersToSend {
		req.Header.Set(key, value)
	}

	if o.debug {
		o.log.Debugf(ctx, "Initiating HTTP %s request on %s", o.method, o.url)
	}

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read and print the response
	respBody := new(bytes.Buffer)
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	if o.debug {
		o.log.Infof(ctx, "Response Status: %s", resp.Status)
		o.log.Infof(ctx, "Response Body: %s", respBody.String())
	}
	return nil
}
