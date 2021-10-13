package engine

type Notifier interface {
	NotifyTo(channel string, obj interface{}, args ...interface{})
	Notify(obj interface{}, args ...interface{})
}

type Notifiability struct {
	notifiers []Notifier
}

func (m *Notifiability) AddNotifier(notifier Notifier) {
	m.notifiers = append(m.notifiers, notifier)
}

func (m *Notifiability) Notify(obj interface{}, args ...interface{}) {
	for _, n := range m.notifiers {
		n.Notify(obj, args...)
	}
}

func (m *Notifiability) NotifyTo(channel string, obj interface{}, args ...interface{}) {
	for _, n := range m.notifiers {
		n.NotifyTo(channel, obj, args...)
	}
}