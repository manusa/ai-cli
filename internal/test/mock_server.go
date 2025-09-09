package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

type MockHandlerFunc func(http.ResponseWriter, *http.Request) (handled bool)

type MockServer struct {
	server       *httptest.Server
	restHandlers []MockHandlerFunc
}

func NewMockServer() *MockServer {
	ms := &MockServer{}
	ms.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, handler := range ms.restHandlers {
			if handler(w, req) {
				return
			}
		}
		http.NotFound(w, req)
	}))
	ms.restHandlers = make([]MockHandlerFunc, 0)
	return ms
}

func (m *MockServer) Close() {
	m.server.Close()
}

func (m *MockServer) URL() string {
	return m.server.URL
}

func (m *MockServer) Handle(handler MockHandlerFunc) {
	m.restHandlers = append(m.restHandlers, handler)
}

func WriteObject(w http.ResponseWriter, obj any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
