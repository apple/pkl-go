//go:build libpkl

package libpkl

import (
	"bytes"
	"github.com/vmihailenco/msgpack/v5"
	"testing"
	"unsafe"

	"github.com/apple/pkl-go/pkl/internal/msgapi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_LibPkl_New_Close(t *testing.T) {
	events := make(chan []byte, 10)
	defer close(events)

	testHandler := func(message []byte, userData unsafe.Pointer) {
		events <- message
	}

	c, err := New(testHandler, "metadata")
	require.Nil(t, err, "Failed to start libpkl")

	err = c.Close()
	require.Nil(t, err)
}

func Test_LibPkl_SendMessage(t *testing.T) {
	events := make(chan []byte, 10)
	defer close(events)

	testHandler := func(message []byte, userData unsafe.Pointer) {
		events <- message
	}

	c, err := New(testHandler, "metadata")
	require.Nil(t, err, "Failed to start libpkl")

	create := &msgapi.CreateEvaluator{
		RequestId:               1,
		ResourceReaders:         nil,
		ModuleReaders:           nil,
		ExternalReaderCommands:  nil,
		ModulePaths:             nil,
		Env:                     nil,
		Properties:              nil,
		OutputFormat:            "",
		AllowedModules:          nil,
		AllowedResources:        nil,
		RootDir:                 "",
		CacheDir:                "",
		Project:                 nil,
		Http:                    nil,
		TimeoutSeconds:          3,
		ExternalModuleReaders:   nil,
		ExternalResourceReaders: nil,
	}

	createMsg, err := create.ToMsgPack()
	require.Nil(t, err)

	err = c.SendMessage(createMsg)
	require.Nil(t, err)

	require.Len(t, events, 1)
	event, err := decode(<-events)
	assert.Nil(t, err, "couldn't deserialize MsgPack")
	t.Logf("event=%#v\n", event)

	closer := &msgapi.CloseEvaluator{EvaluatorId: 1}
	closerMsg, err := closer.ToMsgPack()
	require.Nil(t, err)

	err = c.SendMessage(closerMsg)
	require.Nil(t, err)

	err = c.Close()
	require.Nil(t, err, "Failed to close libpkl")
}

func Test_LibPkl_Version(t *testing.T) {
	events := make(chan []byte, 10)
	defer close(events)

	testHandler := func(message []byte, userData unsafe.Pointer) {
		events <- message
	}

	c, err := New(testHandler, "metadata")
	require.Nil(t, err, "Failed to start libpkl")

	version, err := c.Version()
	require.Nil(t, err)
	assert.NotEmpty(t, version)

	err = c.Close()
	require.Nil(t, err, "Failed to close libpkl")
}

func Test_LibPkl_Version_Err_On_Closed(t *testing.T) {
	events := make(chan []byte, 10)
	defer close(events)

	testHandler := func(message []byte, userData unsafe.Pointer) {
		events <- message
	}

	c, err := New(testHandler, "metadata")
	require.Nil(t, err, "Failed to start libpkl")

	err = c.Close()
	require.Nil(t, err, "Failed to close libpkl")

	version, err := c.Version()
	require.NotNil(t, err, "expected an error when trying to get version on closed libpkl")
	assert.Empty(t, version)
}

func decode(message []byte) (msgapi.IncomingMessage, error) {
	r := bytes.NewBuffer(message)
	dec := msgpack.NewDecoder(r)
	return msgapi.Decode(dec)
}
