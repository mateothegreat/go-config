package config

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CustomConfig struct {
	Base struct {
		A int    `yaml:"a" json:"a" env:"A" required:"true"`
		B string `yaml:"b" json:"b" env:"B" required:"true"`
		C string `yaml:"c" json:"c" env:"C" required:"false"`
	} `yaml:"base" json:"base" env-prefix:"BASE_" required:"true"`
	Foo struct {
		Bar string `yaml:"bar" json:"bar" env:"BAR" required:"true"`
	} `yaml:"foo" json:"foo" required:"true"`
}

type TestSuite struct {
	suite.Suite
	config *CustomConfig
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestGetConfig() {
	var err error
	s.config, err = GetConfig[CustomConfig](GetConfigArgs{
		Paths: []string{
			".env.json",
			".env.base.yaml",
			".env.local.yaml",
		},
		WalkDepth: 7,
	})

	s.NoError(err)
	s.Equal(1, s.config.Base.A)
	s.Equal("2", s.config.Base.B)
	s.Equal("4", s.config.Foo.Bar)
}
