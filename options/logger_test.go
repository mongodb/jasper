package options

import (
	"encoding/json"
	"testing"

	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/send"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestLoggerConfigValidate(t *testing.T) {
	t.Run("NoType", func(t *testing.T) {
		config := LoggerConfig{Format: RawLoggerConfigFormatJSON}
		assert.Error(t, config.Validate())
	})
	t.Run("InvalidLoggerConfigFormat", func(t *testing.T) {
		config := LoggerConfig{
			Type:   LogDefault,
			Format: "foo",
			Config: []byte("some bytes"),
		}
		assert.Error(t, config.Validate())
	})
	t.Run("UnsetRegistry", func(t *testing.T) {
		config := LoggerConfig{
			Type:   LogDefault,
			Format: RawLoggerConfigFormatBSON,
		}
		assert.NoError(t, config.Validate())
		assert.Equal(t, globalLoggerRegistry, config.Registry)
	})
	t.Run("SetRegistry", func(t *testing.T) {
		registry := NewBasicLoggerRegistry()
		config := LoggerConfig{
			Type:     LogDefault,
			Format:   RawLoggerConfigFormatBSON,
			Registry: registry,
		}
		assert.NoError(t, config.Validate())
		assert.Equal(t, registry, config.Registry)
	})
}

func TestLoggerConfigSet(t *testing.T) {
	t.Run("UnregisteredLogger", func(t *testing.T) {
		config := LoggerConfig{
			Format:   RawLoggerConfigFormatBSON,
			Registry: NewBasicLoggerRegistry(),
		}
		assert.Error(t, config.Set(&DefaultLoggerOptions{}))
		assert.Empty(t, config.Type)
		assert.Nil(t, config.producer)
	})
	t.Run("RegisteredLogger", func(t *testing.T) {
		config := LoggerConfig{
			Format:   RawLoggerConfigFormatBSON,
			Registry: globalLoggerRegistry,
		}
		require.NoError(t, config.Set(&DefaultLoggerOptions{}))
		assert.Equal(t, LogDefault, config.Type)
		assert.Equal(t, &DefaultLoggerOptions{}, config.producer)
	})
}

func TestLoggerConfigResolve(t *testing.T) {
	t.Run("InvalidConfig", func(t *testing.T) {
		config := LoggerConfig{}
		require.Error(t, config.Validate())
		sender, err := config.Resolve()
		assert.Nil(t, sender)
		assert.Error(t, err)
	})
	t.Run("UnregisteredLogger", func(t *testing.T) {
		config := LoggerConfig{
			Type:     LogDefault,
			Format:   RawLoggerConfigFormatBSON,
			Registry: NewBasicLoggerRegistry(),
		}
		require.NoError(t, config.Validate())
		sender, err := config.Resolve()
		assert.Nil(t, sender)
		assert.Error(t, err)
	})
	t.Run("MismatchingFormatAndConfig", func(t *testing.T) {
		rawData, err := json.Marshal(&DefaultLoggerOptions{Prefix: "prefix"})
		require.NoError(t, err)
		config := LoggerConfig{
			Type:     LogDefault,
			Format:   RawLoggerConfigFormatBSON,
			Config:   rawData,
			Registry: globalLoggerRegistry,
		}
		require.NoError(t, config.Validate())
		require.True(t, config.Registry.Check(config.Type))
		sender, err := config.Resolve()
		assert.Nil(t, sender)
		assert.Error(t, err)
	})
	t.Run("MismatchingConfigAndProducer", func(t *testing.T) {
		rawData, err := json.Marshal(&DefaultLoggerOptions{Prefix: "prefix"})
		require.NoError(t, err)
		config := LoggerConfig{
			Type:     LogFile,
			Format:   RawLoggerConfigFormatJSON,
			Config:   rawData,
			Registry: globalLoggerRegistry,
		}
		require.NoError(t, config.Validate())
		require.True(t, config.Registry.Check(config.Type))
		sender, err := config.Resolve()
		assert.Nil(t, sender)
		assert.Error(t, err)
	})
	t.Run("InvalidProducerConfig", func(t *testing.T) {
		config := LoggerConfig{
			Type:     LogFile,
			Format:   RawLoggerConfigFormatBSON,
			Registry: globalLoggerRegistry,
			producer: &FileLoggerOptions{},
		}
		require.NoError(t, config.Validate())
		require.True(t, config.Registry.Check(config.Type))
		sender, err := config.Resolve()
		assert.Nil(t, sender)
		assert.Error(t, err)
	})
	t.Run("SenderUnset", func(t *testing.T) {
		config := LoggerConfig{
			Type:     LogDefault,
			Format:   RawLoggerConfigFormatBSON,
			Registry: globalLoggerRegistry,
			producer: &DefaultLoggerOptions{Base: BaseOptions{Format: LogFormatPlain}},
		}
		sender, err := config.Resolve()
		assert.NotNil(t, sender)
		assert.NoError(t, err)
	})
	t.Run("ProducerAndSenderUnsetJSON", func(t *testing.T) {
		rawConfig, err := json.Marshal(&DefaultLoggerOptions{Base: BaseOptions{Format: LogFormatPlain}})
		require.NoError(t, err)
		config := LoggerConfig{
			Type:     LogDefault,
			Format:   RawLoggerConfigFormatJSON,
			Config:   rawConfig,
			Registry: globalLoggerRegistry,
		}
		sender, err := config.Resolve()
		assert.NotNil(t, sender)
		assert.NoError(t, err)
	})
	t.Run("ProducerAndSenderUnsetBSON", func(t *testing.T) {
		rawConfig, err := bson.Marshal(&DefaultLoggerOptions{Base: BaseOptions{Format: LogFormatPlain}})
		require.NoError(t, err)
		config := LoggerConfig{
			Type:     LogDefault,
			Format:   RawLoggerConfigFormatBSON,
			Config:   rawConfig,
			Registry: globalLoggerRegistry,
		}
		sender, err := config.Resolve()
		assert.NotNil(t, sender)
		assert.NoError(t, err)
	})
}

func TestLoggerConfigMarshalBSON(t *testing.T) {
	t.Run("MismatchingFormat", func(t *testing.T) {
		config := LoggerConfig{
			Type:     LogDefault,
			Format:   RawLoggerConfigFormatJSON,
			producer: &DefaultLoggerOptions{},
		}
		data, err := config.MarshalBSON()
		assert.Nil(t, data)
		assert.Error(t, err)
		assert.Nil(t, config.Config)
	})
	t.Run("NilProducer", func(t *testing.T) {
		config := LoggerConfig{
			Type:   LogDefault,
			Format: RawLoggerConfigFormatBSON,
			Config: []byte("some bytes"),
		}
		data, err := bson.Marshal(&config)
		require.NoError(t, err)
		assert.NotNil(t, data)
		unmarshalledConfig := &LoggerConfig{}
		require.NoError(t, bson.Unmarshal(data, unmarshalledConfig))
		assert.Equal(t, config.Type, unmarshalledConfig.Type)
		assert.Equal(t, config.Format, unmarshalledConfig.Format)
		assert.Equal(t, []byte("some bytes"), unmarshalledConfig.Config)
	})
	t.Run("CorrectFormat", func(t *testing.T) {
		config := LoggerConfig{
			Type:   LogDefault,
			Format: RawLoggerConfigFormatBSON,
			Config: []byte("some bytes"),
			producer: &DefaultLoggerOptions{
				Prefix: "jasper",
				Base: BaseOptions{
					Level: send.LevelInfo{
						Default:   level.Info,
						Threshold: level.Info,
					},
					Format: LogFormatPlain,
				},
			},
		}
		data, err := bson.Marshal(&config)
		require.NoError(t, err)
		assert.NotNil(t, data)
		unmarshalledConfig := &LoggerConfig{}
		require.NoError(t, bson.Unmarshal(data, unmarshalledConfig))
		_, err = unmarshalledConfig.Resolve()
		require.NoError(t, err)
		assert.Equal(t, config.Type, unmarshalledConfig.Type)
		assert.Equal(t, config.Format, unmarshalledConfig.Format)
		assert.Equal(t, config.producer, unmarshalledConfig.producer)
	})
}

func TestLoggerConfigMarshalJSON(t *testing.T) {
	t.Run("MismatchingFormat", func(t *testing.T) {
		config := LoggerConfig{
			Type:     LogDefault,
			Format:   RawLoggerConfigFormatBSON,
			producer: &DefaultLoggerOptions{},
		}
		data, err := config.MarshalJSON()
		assert.Nil(t, data)
		assert.Error(t, err)
		assert.Nil(t, config.Config)
	})
	t.Run("NilProducer", func(t *testing.T) {
		config := LoggerConfig{
			Type:   LogDefault,
			Format: RawLoggerConfigFormatJSON,
			Config: []byte("some bytes"),
		}
		data, err := json.Marshal(&config)
		require.NoError(t, err)
		assert.NotNil(t, data)
		unmarshalledConfig := &LoggerConfig{}
		require.NoError(t, json.Unmarshal(data, unmarshalledConfig))
		assert.Equal(t, config.Type, unmarshalledConfig.Type)
		assert.Equal(t, config.Format, unmarshalledConfig.Format)
		assert.Equal(t, []byte("some bytes"), unmarshalledConfig.Config)
	})
	t.Run("CorrectFormat", func(t *testing.T) {
		config := LoggerConfig{
			Type:   LogDefault,
			Format: RawLoggerConfigFormatJSON,
			Config: []byte("some bytes"),
			producer: &DefaultLoggerOptions{
				Prefix: "jasper",
				Base: BaseOptions{
					Level: send.LevelInfo{
						Default:   level.Info,
						Threshold: level.Info,
					},
					Format: LogFormatPlain,
				},
			},
		}
		data, err := json.Marshal(&config)
		require.NoError(t, err)
		assert.NotNil(t, data)
		unmarshalledConfig := &LoggerConfig{}
		require.NoError(t, json.Unmarshal(data, unmarshalledConfig))
		_, err = unmarshalledConfig.Resolve()
		require.NoError(t, err)
		assert.Equal(t, config.Type, unmarshalledConfig.Type)
		assert.Equal(t, config.Format, unmarshalledConfig.Format)
		assert.Equal(t, config.producer, unmarshalledConfig.producer)
	})

}
