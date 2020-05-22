package token

import (
	"context"
	"fmt"
	"github.com/docker/distribution/registry/auth"
	"log"
	"sync/atomic"
	"time"
)

type reloadingAccessController struct {
	accessController atomic.Value // type: auth.AccessController
}

func newReloadingAccessController(options map[string]interface{}) (auth.AccessController, error) {
	accessController, err := newAccessController(options)
	if err != nil {
		return nil, fmt.Errorf("delegated access controller initialization failed: %s", err)
	}
	r := &reloadingAccessController{}
	r.accessController.Store(accessController)

	go func() {
		for range time.Tick(6 * time.Hour) { // TODO add to config options
			accessController, err := newAccessController(options)
			if err != nil {
				log.Printf("reloading access controller failed to reload %s\n", err)
				continue
			}
			r.accessController.Store(accessController)
		}
	}()

	return r, nil
}

func (r *reloadingAccessController) Authorized(ctx context.Context, access ...auth.Access) (context.Context, error) {
	accessController := r.accessController.Load().(auth.AccessController)
	return accessController.Authorized(ctx, access...)
}

// init handles registering the token auth backend.
func init() {
	auth.Register("token", auth.InitFunc(newReloadingAccessController))
}
