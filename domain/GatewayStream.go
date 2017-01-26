package domain

import (
	"quantum/shared/logging"
	"quantum/shared/protocol_v1"

	"github.com/golang/protobuf/proto"
)

// GatewayStream ...
type GatewayStream struct {
	context  string
	gwconn   *GatewayConnection
	gwconntx chan *protocol_v1.Envelope
	logger   logging.Logger
}

// CallMember ...
func (t *GatewayStream) CallMember(msg *protocol_v1.CallMember) {
	msgb, err := proto.Marshal(msg)

	if err != nil {
		t.logger.Error(err.Error())
	}

	t.gwconn.wssend <- &protocol_v1.Envelope{
		Context: t.context,
		Type:    protocol_v1.Envelope_CallMember,
		Data:    msgb,
	}
}
