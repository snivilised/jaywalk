package agenor

import (
	"github.com/snivilised/pants"
)

// Walk requests a sequential traversal of a directory tree.
func Walk() NavigatorFactory {
	return &walkerFac{}
}

// Sprint requests a concurrent traversal of a directory tree.
func Sprint(wg pants.WaitGroup) NavigatorFactory {
	return &runnerFac{
		wg: wg,
	}
}

type walkerFac struct {
}

// Configure builds a driver with a session that is configured for sequential
// traversal. It uses the Builders to construct the necessary artefacts for
// the session, including the kernel controller, options, extent, and plugins.
// The resulting driver is returned as a Navigator.
func (f *walkerFac) Configure(addons ...Addon) Director {
	return director(func(bs *Builders) Navigator {
		artefacts, err := bs.buildAll(addons...)

		return &driver{
			session{
				sync: &sequential{
					trunk: trunk{
						kc:  artefacts.kc,
						o:   artefacts.o,
						ext: artefacts.ext,
						err: err,
					},
				},
				plugins: artefacts.plugins,
			},
		}
	})
}

type runnerFac struct {
	wg pants.WaitGroup
}

// Configure builds a driver with a session that is configured for concurrent traversal.
// It uses the Builders to construct the necessary artefacts for
// the session, including the kernel controller, options, extent, and plugins.
func (f *runnerFac) Configure(addons ...Addon) Director {
	return director(func(bs *Builders) Navigator {
		artefacts, err := bs.buildAll(addons...)

		return &driver{
			session{
				sync: &concurrent{
					trunk: trunk{
						kc:  artefacts.kc,
						o:   artefacts.o,
						ext: artefacts.ext,
						err: err,
					},
					wg: f.wg,
				},
				plugins: artefacts.plugins,
			},
		}
	})
}

type (
	// Addon is a type that can be added to a session.
	Addon any
)
