// Auto-generated by avdl-compiler v1.3.3 (https://github.com/keybase/node-avdl-compiler)
//   Input file: gregor1/index.avdl

package gregor1

import (
	rpc "github.com/keybase/go-framed-msgpack-rpc"
)

type IndexInterface interface {
}

func IndexProtocol(i IndexInterface) rpc.Protocol {
	return rpc.Protocol{
		Name:    "gregor.1.index",
		Methods: map[string]rpc.ServeHandlerDescription{},
	}
}

type IndexClient struct {
	Cli rpc.GenericClient
}
