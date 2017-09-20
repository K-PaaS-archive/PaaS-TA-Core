// This file was generated by counterfeiter
package kawasakifakes

import (
	"net"
	"sync"

	"code.cloudfoundry.org/guardian/kawasaki"
	"code.cloudfoundry.org/lager"
)

type FakeResolvFileCompiler struct {
	CompileStub        func(log lager.Logger, resolvConfPath string, containerIp net.IP, overrideServers []net.IP) ([]byte, error)
	compileMutex       sync.RWMutex
	compileArgsForCall []struct {
		log             lager.Logger
		resolvConfPath  string
		containerIp     net.IP
		overrideServers []net.IP
	}
	compileReturns struct {
		result1 []byte
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeResolvFileCompiler) Compile(log lager.Logger, resolvConfPath string, containerIp net.IP, overrideServers []net.IP) ([]byte, error) {
	var overrideServersCopy []net.IP
	if overrideServers != nil {
		overrideServersCopy = make([]net.IP, len(overrideServers))
		copy(overrideServersCopy, overrideServers)
	}
	fake.compileMutex.Lock()
	fake.compileArgsForCall = append(fake.compileArgsForCall, struct {
		log             lager.Logger
		resolvConfPath  string
		containerIp     net.IP
		overrideServers []net.IP
	}{log, resolvConfPath, containerIp, overrideServersCopy})
	fake.recordInvocation("Compile", []interface{}{log, resolvConfPath, containerIp, overrideServersCopy})
	fake.compileMutex.Unlock()
	if fake.CompileStub != nil {
		return fake.CompileStub(log, resolvConfPath, containerIp, overrideServers)
	} else {
		return fake.compileReturns.result1, fake.compileReturns.result2
	}
}

func (fake *FakeResolvFileCompiler) CompileCallCount() int {
	fake.compileMutex.RLock()
	defer fake.compileMutex.RUnlock()
	return len(fake.compileArgsForCall)
}

func (fake *FakeResolvFileCompiler) CompileArgsForCall(i int) (lager.Logger, string, net.IP, []net.IP) {
	fake.compileMutex.RLock()
	defer fake.compileMutex.RUnlock()
	return fake.compileArgsForCall[i].log, fake.compileArgsForCall[i].resolvConfPath, fake.compileArgsForCall[i].containerIp, fake.compileArgsForCall[i].overrideServers
}

func (fake *FakeResolvFileCompiler) CompileReturns(result1 []byte, result2 error) {
	fake.CompileStub = nil
	fake.compileReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *FakeResolvFileCompiler) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.compileMutex.RLock()
	defer fake.compileMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeResolvFileCompiler) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ kawasaki.ResolvFileCompiler = new(FakeResolvFileCompiler)