package urlhandler

import "github.com/JOTPOT-UK/JOTPOT-Server/jps"

import "strings"

func StripPort(f *jps.ServerFrame) (bool, error) {
	u := f.Request.URL()
	u.Host = u.Hostname()
	return false, nil
}

var PortStripper = jps.ServerFrameHandler{
	F: StripPort,
}

func StripHTTPPort(f *jps.ServerFrame) (bool, error) {
	u := f.Request.URL()
	port := u.Port()
	if port == "80" || port == "443" {
		u.Host = u.Hostname()
	}
	return false, nil
}

var HTTPPortStripper = jps.ServerFrameHandler{
	F: StripHTTPPort,
}

type HostSplitter map[string]*jps.ServerFrameHandler

func (hs HostSplitter) Handle(f *jps.ServerFrame) (bool, error) {
	handler := hs[f.Request.URL().Host]
	//if len(handler.Protocol) == 0 || f.Protocol().Protocol() == handler.Protocol { REMOVED PROTOCOL CHECK
	return handler.F(f)
	/*}
	return false, nil*/
}

type HostnameSplitter map[string]*jps.ServerFrameHandler

func (hns HostnameSplitter) Handle(f *jps.ServerFrame) (bool, error) {
	handler := hns[f.Request.URL().Hostname()]
	if len(handler.Protocol) == 0 || f.Protocol().Protocol() == handler.Protocol {
		return handler.F(f)
	}
	return false, nil
}

type PathPrefixSplitter struct {
	Prefix string
	F      func(*jps.ServerFrame) (bool, error)
}

func (pps PathPrefixSplitter) Handle(f *jps.ServerFrame) (bool, error) {
	if strings.HasPrefix(f.Request.URL().Path, pps.Prefix) {
		return pps.F(f)
	}
	return false, nil
}

type PathPrefixStripSplitter struct {
	Prefix string
	F      func(*jps.ServerFrame) (bool, error)
}

func (ppss PathPrefixStripSplitter) Handle(f *jps.ServerFrame) (bool, error) {
	u := f.Request.URL()
	if strings.HasPrefix(u.Path, ppss.Prefix) {
		u.Path = u.Path[len(ppss.Prefix):]
		return ppss.F(f)
	}
	return false, nil
}

func PathPrefixHandler(pathPrefix string, handler jps.ServerFrameHandler, strip bool) jps.ServerFrameHandler {
	if strip {
		return jps.ServerFrameHandler{
			F: PathPrefixStripSplitter{
				Prefix: pathPrefix,
				F:      handler.F,
			}.Handle,
			Protocol: handler.Protocol,
		}
	}
	return jps.ServerFrameHandler{
		F: PathPrefixSplitter{
			Prefix: pathPrefix,
			F:      handler.F,
		}.Handle,
		Protocol: handler.Protocol,
	}
}

type PathHandler struct {
	Path string
	F    func(*jps.ServerFrame) (bool, error)
}

func (ph PathHandler) Handle(f *jps.ServerFrame) (bool, error) {
	if f.Request.URL().Path == ph.Path {
		return ph.F(f)
	}
	return false, nil
}

func NewPathHandler(path string, handler jps.ServerFrameHandler) jps.ServerFrameHandler {
	return jps.ServerFrameHandler{
		Protocol: handler.Protocol,
		F: PathHandler{
			Path: path,
			F:    handler.F,
		}.Handle,
	}
}
