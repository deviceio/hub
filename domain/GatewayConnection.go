package domain

import (
	"quantum/hub/infra"
	"quantum/shared/logging"
	"quantum/shared/protocol_v1"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// GatewayConnection represents the underlying connection of a connected device to
// the gateway component
type GatewayConnection struct {
	// gwservice used to interact with the parent GatewayService domain-level api
	gwservice *GatewayService

	// gwstreams map of GatewayStreams this GatewayConnection is currently managing
	gwstreams map[string]*GatewayStream

	// handshake reference to the device's initial connection handshake containing
	// important device information to orchestrate the protocol
	handshake *protocol_v1.Handshake

	// logger provides logging for this type
	logger logging.Logger

	// wsclosed channel to coordinate shutdown of this GatewayConnection
	wsclosed chan bool

	// wsconn the underlying websocket connection this GatewayConnection represents
	wsconn *websocket.Conn

	// wssend channel used to send new protocol envelopes to the underlying websocket connection
	wssend chan *protocol_v1.Envelope

	// mutex is used for syncronization
	mutex *sync.Mutex
}

// NewGatewayConnection instantiates a new instance of the GatewayConnection type
func NewGatewayConnection(wsconn *websocket.Conn, logger logging.Logger, gateway *GatewayService) *GatewayConnection {
	gc := &GatewayConnection{
		gwservice: gateway,
		gwstreams: make(map[string]*GatewayStream),
		logger:    logger,
		wsclosed:  make(chan bool),
		wsconn:    wsconn,
		wssend:    make(chan *protocol_v1.Envelope),
		mutex:     &sync.Mutex{},
	}

	return gc
}

// OpenStream ...
func (t *GatewayConnection) OpenStream() *GatewayStream {
	uuid, err := uuid.NewRandom()

	if err != nil {
		t.logger.Error(err.Error())
		return nil
	}

	stream := &GatewayStream{
		context:  uuid.String(),
		gwconntx: make(chan *protocol_v1.Envelope, 500),
		gwconn:   t,
		logger:   &logging.DefaultLogger{},
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.gwstreams[stream.context] = stream

	return stream
}

// CloseStream ...
func (t *GatewayConnection) CloseStream(s *GatewayStream) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	delete(t.gwstreams, s.context)
	close(s.gwconntx)
}

func (t *GatewayConnection) sendStreamEnvelope(e *protocol_v1.Envelope) {
	t.mutex.Lock()
	v, ok := t.gwstreams[e.Context]
	t.mutex.Unlock()

	if !ok {
		t.logger.Error("No such context found")
		close(t.wsclosed)
		return
	}

	go func() {}()

	select {
	case v.gwconntx <- e:
		return
	}
}

func (t *GatewayConnection) start() {
	go t.wsReadLoop()
	go t.wsWriteLoop()
	go t.wsCloseLoop()
}

func (t *GatewayConnection) wsReadLoop() {
	for {
		select {
		case <-t.wsclosed:
			return
		default:
			break
		}

		mt, mb, err := t.wsconn.ReadMessage()

		if err != nil {
			t.logger.Error(err.Error())
			close(t.wsclosed)
			return
		}

		if mt != websocket.BinaryMessage {
			t.logger.Error("Agent Connection attempted to send plain-text data")
			close(t.wsclosed)
			return
		}

		envelope := &protocol_v1.Envelope{}

		err = proto.Unmarshal(mb, envelope)

		if err != nil {
			t.logger.Error("Agent Connection failed to unmarshal protocol envelope")
			close(t.wsclosed)
			return
		}

		if envelope.Type == protocol_v1.Envelope_Handshake && t.handshake != nil {
			t.logger.Error("Agent duplicate handshake messages recieved")
			close(t.wsclosed)
			return
		}

		if envelope.Type == protocol_v1.Envelope_Handshake && t.handshake == nil {
			handshake := &protocol_v1.Handshake{}

			err := proto.Unmarshal(envelope.Data, handshake)

			if err != nil {
				t.logger.Error("Failed to unmarshal device handshake from Agent")
				close(t.wsclosed)
				return
			}

			device := infra.NewClusterDeviceModelFromHandshake(handshake, true)

			if !t.gwservice.hub.Cluster.AuthDevice(device) {
				t.logger.Error("Device failed handshake")
				close(t.wsclosed)
				return
			}

			t.handshake = handshake
			t.gwservice.addConnection(t)
			t.gwservice.hub.Cluster.AddOrUpdateDevice(device)
			continue
		}

		if envelope.Type != protocol_v1.Envelope_Handshake && t.handshake == nil {
			t.logger.Error("Agent message before handshake")
			close(t.wsclosed)
			return
		}

		switch envelope.Type {
		case protocol_v1.Envelope_CallMember:
			t.logger.Error("Agent sent CallMember message")
			close(t.wsclosed)
			return

		case protocol_v1.Envelope_GetMember:
			t.logger.Error("Agent sent GetMember message")
			close(t.wsclosed)
			return

		case protocol_v1.Envelope_SetMember:
			t.logger.Error("Agent sent SetMember message")
			close(t.wsclosed)
			return

		case protocol_v1.Envelope_CreateObject:
			t.logger.Error("Agent sent CreateObject message")
			close(t.wsclosed)
			return

		case protocol_v1.Envelope_DescribeObject:
			t.logger.Error("Agent sent DescribeObject message")
			close(t.wsclosed)
			return

		case protocol_v1.Envelope_DeleteObject:
			t.logger.Error("Agent sent DeleteObject message")
			close(t.wsclosed)
			return

		case protocol_v1.Envelope_Bool:
			t.sendStreamEnvelope(envelope)

		case protocol_v1.Envelope_Bytes:
			t.sendStreamEnvelope(envelope)

		case protocol_v1.Envelope_DateTimeISO8601:
			t.sendStreamEnvelope(envelope)

		case protocol_v1.Envelope_Error:
			t.sendStreamEnvelope(envelope)

		case protocol_v1.Envelope_Float32:
			t.sendStreamEnvelope(envelope)

		case protocol_v1.Envelope_Float64:
			t.sendStreamEnvelope(envelope)

		case protocol_v1.Envelope_Int32:
			t.sendStreamEnvelope(envelope)

		case protocol_v1.Envelope_Int64:
			t.sendStreamEnvelope(envelope)

		case protocol_v1.Envelope_StringUTF8:
			t.sendStreamEnvelope(envelope)

		case protocol_v1.Envelope_Close:
			t.sendStreamEnvelope(envelope)

		default:
			t.logger.Error("Agent sent unknown message", envelope.Type, envelope.Context, envelope.Data)
			close(t.wsclosed)
			return
		}
	}
}

func (t *GatewayConnection) wsWriteLoop() {
	for {
		select {
		case <-t.wsclosed:
			return
		case msg := <-t.wssend:
			msgb, err := proto.Marshal(msg)

			if err != nil {
				t.logger.Error(err.Error())
			}

			if err := t.wsconn.WriteMessage(websocket.BinaryMessage, msgb); err != nil {
				t.logger.Error(err.Error())
			}
		}
	}
}

func (t *GatewayConnection) wsCloseLoop() {
	<-t.wsclosed
	t.wsconn.Close()

	t.gwservice.delConnection(t)
	t.gwservice.hub.Cluster.AddOrUpdateDevice(
		infra.NewClusterDeviceModelFromHandshake(t.handshake, false),
	)
}
