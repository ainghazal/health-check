package health

// TODO can move to minivpn/ extras https://github.com/ooni/minivpn/issues/24

type nullLogger struct{}

func (n *nullLogger) Info(string) {}

func (n *nullLogger) Infof(string, ...interface{}) {}

func (n *nullLogger) Debug(string) {}

func (n *nullLogger) Debugf(string, ...interface{}) {}

func (n *nullLogger) Warn(string) {}

func (n *nullLogger) Warnf(string, ...interface{}) {}

func (n *nullLogger) Error(string) {}

func (n *nullLogger) Errorf(string, ...interface{}) {}
