package server

import (
	iris "github.com/daemon-coder/idalloc/infrastructure/iris_infra"
	"github.com/daemon-coder/idalloc/transport"
)

func AddRoute(app *iris.IrisApp) {
	app.Handle("POST", "/alloc", iris.JsonWrapper(transport.Alloc))
}
