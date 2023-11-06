package obstasks

import (
	"net.kopias.oscbridge/app/drivers/obsremote"
	"net.kopias.oscbridge/app/usecase/usecaseifs"
)

const ParamConnectionKey = "connection"

func NewSceneChangerFactory(obsConnections map[string]*obsremote.OBSRemote, log usecaseifs.ILogger, debug bool) usecaseifs.ActionTaskFactory {
	return func() usecaseifs.IActionTask { return NewSceneChanger(obsConnections, log, debug) }
}

func NewVendorRequestFactory(obsConnections map[string]*obsremote.OBSRemote, log usecaseifs.ILogger, debug bool) usecaseifs.ActionTaskFactory {
	return func() usecaseifs.IActionTask { return NewVendorRequest(obsConnections, log, debug) }
}
