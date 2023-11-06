package obsremote

import (
	"context"
	"strings"

	"github.com/andreykaipov/goobs/api"
	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

var _ api.Logger = &obsRemoteLogger{}

type obsRemoteLogger struct {
	logger usecaseifs.ILogger
	ctx    context.Context
	debug  bool
}

func newOBSRemoteLogger(ctx context.Context, logger usecaseifs.ILogger, debug bool) *obsRemoteLogger {
	return &obsRemoteLogger{ctx: ctx, logger: logger, debug: debug}
}

func (orem *obsRemoteLogger) Printf(s string, i ...interface{}) {
	if !orem.debug {
		if strings.HasPrefix(s, "[DEBUG]") || strings.HasPrefix(s, "[INFO]") {
			return
		}
	}
	orem.logger.Infof(orem.ctx, s, i...)
}
