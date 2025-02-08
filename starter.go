package openFiegnStarter

import (
	"github.com/dennesshen/photon-core-starter/core"
	"github.com/dennesshen/photon-openfeign-starter/openfeign"
)

func init() {
	core.RegisterCoreDependency(openfeign.StartOpenFeign)
}
