package remote

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mongodb/grip"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/testutil"
	"github.com/pkg/errors"
)

func startRESTService(ctx context.Context, client *http.Client) (*Service, int, error) {
tryStartService:
	for {
		select {
		case <-ctx.Done():
			return nil, -1, errors.WithStack(ctx.Err())
		default:
			mngr, err := jasper.NewSynchronizedManager(false)
			if err != nil {
				return nil, -1, errors.WithStack(err)
			}
			srv := NewRESTService(mngr)
			app := srv.App(ctx)
			app.SetPrefix("jasper")

			if err := app.SetHost("localhost"); err != nil {
				continue tryStartService
			}

			port := testutil.GetPortNumber()
			if err := app.SetPort(port); err != nil {
				continue tryStartService
			}

			srvCtx, srvCancel := context.WithCancel(ctx)
			go func() {
				grip.Warning(app.Run(srvCtx))
			}()

			failedToConnect := make(chan struct{})
			go func() {
				defer srvCancel()
				select {
				case <-ctx.Done():
				case <-failedToConnect:
				}
			}()

			url := fmt.Sprintf("http://localhost:%d/jasper/v1/", port)
			if err := tryConnectToRESTService(ctx, url); err != nil {
				close(failedToConnect)
				continue
			}
			return srv, port, nil
		}
	}
}

func tryConnectToRESTService(ctx context.Context, url string) error {
	maxAttempts := 10
	for attempt := 0; attempt < 10; attempt++ {
		err := func() error {
			connCtx, connCancel := context.WithTimeout(ctx, time.Second)
			defer connCancel()
			if err := testutil.WaitForHTTPService(connCtx, url); err != nil {
				return errors.WithStack(err)
			}
			return nil
		}()
		if err != nil {
			continue
		}
		return nil
	}
	return errors.Errorf("failed to connect after %d attempts", maxAttempts)
}
