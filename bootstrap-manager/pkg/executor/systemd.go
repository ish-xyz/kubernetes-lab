package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	SYSTEMD_DONE = "done"
)

// Wait for the channel for 10 minutes trying every 5 seconds to read from it
func waitForChannel(ch chan string) (string, error) {
	var res string
	for retry := 0; retry <= 120; retry++ {

		select {
		case res = <-ch:
			return res, nil
		default:
			time.Sleep(5 * time.Second)
		}

	}

	return "", fmt.Errorf("operation took too long")
}

func (e *Executor) StopServiceW(svc string) error {

	var err error
	var res string

	logrus.Infoln("stopping service", svc)
	for retry := 0; retry <= 10; retry++ {

		logrus.Infof("trying to send dbus message to systemd to stop %s ...", svc)

		ch := make(chan string)
		ctx, cancelCtx := context.WithCancel(context.Background())

		_, err = e.SystemdConn.StopUnitContext(ctx, svc, "replace", ch)
		if err == nil {
			res, err = waitForChannel(ch)
			if err == nil {
				if res == SYSTEMD_DONE {
					cancelCtx()
					return nil
				} else {
					err = fmt.Errorf("systemd did not complete operation in time => %s", res)
				}
			}
		}

		cancelCtx()
		time.Sleep(5 * time.Second)
	}

	return err
}

func (e *Executor) StopService(svc string) error {

	logrus.Infof("trying to stop service %s", svc)

	// TODO: add retry mechanism
	ch := make(chan string)
	_, err := e.SystemdConn.StopUnitContext(context.TODO(), svc, "replace", ch)
	if err != nil {
		return fmt.Errorf("error sending stop signal to service %s => %v", svc, err)
	}

	res := <-ch // stopping here and waiting for systemd to reply //TODO: add timeout
	if res != SYSTEMD_DONE {
		return fmt.Errorf("systemd couldn't stop service %s => %v", svc, err)
	}

	return nil
}

func (e *Executor) DisableServices(units []string) error {
	_, err := e.SystemdConn.DisableUnitFilesContext(context.TODO(), units, false)
	return err
}

// func (e *Executor) StartServices(failmode int, services ...string) error {

// 	for _, svc := range services {
// 		logrus.Infoln("trying to stop service %s", svc)

// 		ch := make(chan string)
// 		_, err := e.SystemdConn.StartUnitContext(context.TODO(), svc, "replace", ch)
// 		if err != nil && failmode != MODE_FAIL_FAST {
// 			logrus.Warningf("error while stopping service %s => %v", svc, err)
// 			continue
// 		}

// 		res := <-ch // stopping here and waiting for systemd to reply
// 		if res != SYSTEMD_DONE && failmode != MODE_FAIL_FAST {
// 			logrus.Warningf("error while stopping service %s => %v", svc, err)
// 		}

// 		if err != nil || res != SYSTEMD_DONE {
// 			return fmt.Errorf("failed to stop service %s => %v", svc, err)
// 		}
// 	}

// 	return nil
// }
