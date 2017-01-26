package domain

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/deviceio/shared/protocol_v1"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
)

// APIService ...
type APIService struct {
	hub  *Hub
	opts *APIOptions
}

// NewAPIService ...
func NewAPIService(hub *Hub, opts *APIOptions) *APIService {
	return &APIService{
		hub:  hub,
		opts: opts,
	}
}

// Start ...
func (t *APIService) Start() {
	var err error

	server := http.NewServeMux()
	router := mux.NewRouter()

	router.HandleFunc("/status", t.status).Methods("GET")
	router.HandleFunc("/device/{deviceid}/resource/{objectid}/{member}", t.callDeviceResourceMember).Methods("POST")

	server.Handle("/", t.auth(router))

	err = http.ListenAndServeTLS(
		t.opts.BindAddr,
		t.opts.TLSCertPath,
		t.opts.TLSKeyPath,
		server,
	)

	if err != nil {
		log.Fatal(err, string(debug.Stack()))
	}
}

func (t *APIService) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		next.ServeHTTP(w, r)
	})
}

// httpGetStatus GET /v1/status
func (t *APIService) status(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("OK"))
}

// paramsToByteMap ...
func (t *APIService) paramsToByteMap(p map[string]interface{}) map[string][]byte {
	bm := map[string][]byte{}

	for k, v := range p {
		switch tv := v.(type) {
		case string:
			bm[k] = []byte(tv)
		}
	}

	return bm
}

// httpPostDeviceResourceMember POST /v1/device/:deviceid/resource/:objectid
func (t *APIService) callDeviceResourceMember(rw http.ResponseWriter, r *http.Request) {
	httpclose := rw.(http.CloseNotifier).CloseNotify()
	httpflusher := rw.(http.Flusher)
	vars := mux.Vars(r)

	deviceid := vars["deviceid"]
	objectid := vars["objectid"]
	member := vars["member"]

	if !t.hub.Cluster.DeviceExists(deviceid) {
		rw.WriteHeader(404)
		rw.Write([]byte("No such device exists"))
		return
	}

	if !t.hub.Cluster.IsDeviceConnected(deviceid) {
		rw.WriteHeader(400)
		rw.Write([]byte("Device is not connected"))
		return
	}

	if t.hub.Cluster.ShouldProxyRequest(deviceid) {
		t.hub.Cluster.ProxyRequest(deviceid, rw, r)
		return
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		rw.WriteHeader(400)
		rw.Write([]byte("Invalid param body"))
		return
	}

	var callparams map[string]interface{}

	if err = json.Unmarshal(body, &callparams); err != nil {
		rw.WriteHeader(400)
		rw.Write([]byte("Invalid param body"))
		return
	}

	conn := t.hub.Gateway.FindConnectionForDevice(deviceid)

	stream := conn.OpenStream()
	defer conn.CloseStream(stream)

	stream.CallMember(&protocol_v1.CallMember{
		Reference: objectid,
		Name:      member,
		Params:    t.paramsToByteMap(callparams),
	})

	rw.Header().Set("Trailer", "Error")
	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(200)

	var envelope *protocol_v1.Envelope

	for {
		select {
		case <-httpclose:
			log.Println("Lost api connetion")
			return
		case envelope = <-stream.gwconntx:
			log.Println("Received env", envelope.Type)
			break
		}

		switch envelope.Type {
		case protocol_v1.Envelope_Close:
			log.Println("Recieving Close message")

			rw.Write([]byte(""))
			httpflusher.Flush()
			rw.Header().Set("Error", "")
			httpflusher.Flush()
			return

		case protocol_v1.Envelope_Error:
			log.Println("Recieving Error Message")
			e := &protocol_v1.Error{}

			if err := proto.Unmarshal(envelope.Data, e); err != nil {
				rw.Write([]byte(""))
				rw.Header().Set("Error", err.Error())
				return
			}

			rw.Write([]byte(""))
			httpflusher.Flush()
			rw.Header().Set("Error", e.Message)
			httpflusher.Flush()
			return

		case protocol_v1.Envelope_Bytes:
			log.Println("Recieving Bytes message")
			b := &protocol_v1.Bytes{}

			if err := proto.Unmarshal(envelope.Data, b); err != nil {
				rw.Write([]byte(""))
				rw.Header().Set("Error", err.Error())
				return
			}

			rw.Write(b.Value)
			httpflusher.Flush()
		}
	}
}
