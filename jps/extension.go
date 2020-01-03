package jps

type Extension interface {
	ServerExtensionInit(*Server) error
}

type ExtensionList []Extension

func (exts ExtensionList) ServerExtensionInit(s *Server) error {
	for _, ext := range exts {
		err := ext.ServerExtensionInit(s)
		if err != nil {
			return err
		}
	}
	return nil
}
