// This file was generated by github.com/nelsam/hel.  Do not
// edit this code by hand unless you *really* know what you're
// doing.  Expect any changes made manually to be overwritten
// the next time hel regenerates this file.

package component_tests_test

import (
	"context"

	v2 "github.com/cloudfoundry/statsd-injector/internal/plumbing/v2"
	"google.golang.org/grpc/metadata"
)

type mockMetronIngressServer struct {
	SenderCalled chan bool
	SenderInput  struct {
		Arg0 chan v2.Ingress_SenderServer
	}
	SenderOutput struct {
		Ret0 chan error
	}
}

func newMockMetronIngressServer() *mockMetronIngressServer {
	m := &mockMetronIngressServer{}
	m.SenderCalled = make(chan bool, 100)
	m.SenderInput.Arg0 = make(chan v2.Ingress_SenderServer, 100)
	m.SenderOutput.Ret0 = make(chan error, 100)
	return m
}
func (m *mockMetronIngressServer) Sender(arg0 v2.Ingress_SenderServer) error {
	m.SenderCalled <- true
	m.SenderInput.Arg0 <- arg0
	return <-m.SenderOutput.Ret0
}

type mockIngress_SenderServer struct {
	SendAndCloseCalled chan bool
	SendAndCloseInput  struct {
		Arg0 chan *v2.IngressResponse
	}
	SendAndCloseOutput struct {
		Ret0 chan error
	}
	RecvCalled chan bool
	RecvOutput struct {
		Ret0 chan *v2.Envelope
		Ret1 chan error
	}
	SetHeaderCalled chan bool
	SetHeaderInput  struct {
		Arg0 chan metadata.MD
	}
	SetHeaderOutput struct {
		Ret0 chan error
	}
	SendHeaderCalled chan bool
	SendHeaderInput  struct {
		Arg0 chan metadata.MD
	}
	SendHeaderOutput struct {
		Ret0 chan error
	}
	SetTrailerCalled chan bool
	SetTrailerInput  struct {
		Arg0 chan metadata.MD
	}
	ContextCalled chan bool
	ContextOutput struct {
		Ret0 chan context.Context
	}
	SendMsgCalled chan bool
	SendMsgInput  struct {
		M chan interface{}
	}
	SendMsgOutput struct {
		Ret0 chan error
	}
	RecvMsgCalled chan bool
	RecvMsgInput  struct {
		M chan interface{}
	}
	RecvMsgOutput struct {
		Ret0 chan error
	}
}

func newMockIngress_SenderServer() *mockIngress_SenderServer {
	m := &mockIngress_SenderServer{}
	m.SendAndCloseCalled = make(chan bool, 100)
	m.SendAndCloseInput.Arg0 = make(chan *v2.IngressResponse, 100)
	m.SendAndCloseOutput.Ret0 = make(chan error, 100)
	m.RecvCalled = make(chan bool, 100)
	m.RecvOutput.Ret0 = make(chan *v2.Envelope, 100)
	m.RecvOutput.Ret1 = make(chan error, 100)
	m.SetHeaderCalled = make(chan bool, 100)
	m.SetHeaderInput.Arg0 = make(chan metadata.MD, 100)
	m.SetHeaderOutput.Ret0 = make(chan error, 100)
	m.SendHeaderCalled = make(chan bool, 100)
	m.SendHeaderInput.Arg0 = make(chan metadata.MD, 100)
	m.SendHeaderOutput.Ret0 = make(chan error, 100)
	m.SetTrailerCalled = make(chan bool, 100)
	m.SetTrailerInput.Arg0 = make(chan metadata.MD, 100)
	m.ContextCalled = make(chan bool, 100)
	m.ContextOutput.Ret0 = make(chan context.Context, 100)
	m.SendMsgCalled = make(chan bool, 100)
	m.SendMsgInput.M = make(chan interface{}, 100)
	m.SendMsgOutput.Ret0 = make(chan error, 100)
	m.RecvMsgCalled = make(chan bool, 100)
	m.RecvMsgInput.M = make(chan interface{}, 100)
	m.RecvMsgOutput.Ret0 = make(chan error, 100)
	return m
}
func (m *mockIngress_SenderServer) SendAndClose(arg0 *v2.IngressResponse) error {
	m.SendAndCloseCalled <- true
	m.SendAndCloseInput.Arg0 <- arg0
	return <-m.SendAndCloseOutput.Ret0
}
func (m *mockIngress_SenderServer) Recv() (*v2.Envelope, error) {
	m.RecvCalled <- true
	return <-m.RecvOutput.Ret0, <-m.RecvOutput.Ret1
}
func (m *mockIngress_SenderServer) SetHeader(arg0 metadata.MD) error {
	m.SetHeaderCalled <- true
	m.SetHeaderInput.Arg0 <- arg0
	return <-m.SetHeaderOutput.Ret0
}
func (m *mockIngress_SenderServer) SendHeader(arg0 metadata.MD) error {
	m.SendHeaderCalled <- true
	m.SendHeaderInput.Arg0 <- arg0
	return <-m.SendHeaderOutput.Ret0
}
func (m *mockIngress_SenderServer) SetTrailer(arg0 metadata.MD) {
	m.SetTrailerCalled <- true
	m.SetTrailerInput.Arg0 <- arg0
}
func (m *mockIngress_SenderServer) Context() context.Context {
	m.ContextCalled <- true
	return <-m.ContextOutput.Ret0
}
func (m *mockIngress_SenderServer) SendMsg(m_ interface{}) error {
	m.SendMsgCalled <- true
	m.SendMsgInput.M <- m_
	return <-m.SendMsgOutput.Ret0
}
func (m *mockIngress_SenderServer) RecvMsg(m_ interface{}) error {
	m.RecvMsgCalled <- true
	m.RecvMsgInput.M <- m_
	return <-m.RecvMsgOutput.Ret0
}
