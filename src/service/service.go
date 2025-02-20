package service

import (
	"fmt"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	ServiceName = "Dark Mode Theme Autochanger"
	ServiceDesc = "Automatically changes Windows theme between light and dark mode based on sunrise/sunset"
)

type ThemeService struct {
	cancel chan bool
	elog   *eventlog.Log
}

func New() (*ThemeService, error) {
	elog, err := eventlog.Open(ServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open eventlog: %v", err)
	}

	return &ThemeService{
		cancel: make(chan bool),
		elog:   elog,
	}, nil
}

func (s *ThemeService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	// Start the theme daemon in a goroutine
	go s.runThemeDaemon()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	// Service loop using for range
	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			close(s.cancel)
			changes <- svc.Status{State: svc.StopPending}
			return
		default:
			s.elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
		}
	}

	return
}
