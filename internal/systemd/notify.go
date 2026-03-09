package systemd

import "github.com/coreos/go-systemd/v22/daemon"

type Notifier interface {
	Ready() error
}

type notifier struct{}
type noopNotifier struct{}

func New() Notifier {
	return notifier{}
}

func NewNoop() Notifier {
	return noopNotifier{}
}

func (notifier) Ready() error {
	_, err := daemon.SdNotify(false, daemon.SdNotifyReady)
	return err
}

func (noopNotifier) Ready() error {
	return nil
}
