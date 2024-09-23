package executor

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	SYSTEMD_DONE     = "done"
	MODE_FAIL_FAST   = 1
	CONTINUE_ON_FAIL = 2
)

func (e *Executor) StopServices(failmode int, services ...string) error {

	for _, svc := range services {
		logrus.Infoln("trying to stop service %s", svc)

		ch := make(chan string)
		_, err := e.SystemdConn.StopUnit(svc, "replace", ch)
		if err != nil && failmode != MODE_FAIL_FAST {
			logrus.Warningf("error while stopping service %s => %v", svc, err)
			continue
		}

		res := <-ch // stopping here and waiting for systemd to reply
		if res != SYSTEMD_DONE && failmode != MODE_FAIL_FAST {
			logrus.Warningf("error while stopping service %s => %v", svc, err)
		}

		if err != nil || res != SYSTEMD_DONE {
			return fmt.Errorf("failed to stop service %s => %v", svc, err)
		}
	}

	return nil
}

func (e *Executor) DisableServices(units []string) error {
	_, err := e.SystemdConn.DisableUnitFiles(units, false)
	return err
}

func (e *Executor) StartServices(failmode int, services ...string) error {

	for _, svc := range services {
		logrus.Infoln("trying to stop service %s", svc)

		ch := make(chan string)
		_, err := e.SystemdConn.StartUnit(svc, "replace", ch)
		if err != nil && failmode != MODE_FAIL_FAST {
			logrus.Warningf("error while stopping service %s => %v", svc, err)
			continue
		}

		res := <-ch // stopping here and waiting for systemd to reply
		if res != SYSTEMD_DONE && failmode != MODE_FAIL_FAST {
			logrus.Warningf("error while stopping service %s => %v", svc, err)
		}

		if err != nil || res != SYSTEMD_DONE {
			return fmt.Errorf("failed to stop service %s => %v", svc, err)
		}
	}

	return nil
}
