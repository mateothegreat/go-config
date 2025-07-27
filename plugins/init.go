package plugins

import (
	"github.com/mateothegreat/go-config/plugins/sources"
	"github.com/mateothegreat/go-config/plugins/validators"
)

// Adapter types to bridge between interfaces
type sourceFactoryAdapter struct {
	factory sources.PluginFactory
}

func (sfa *sourceFactoryAdapter) Create(opts any) (Plugin, error) {
	return sfa.factory.Create(opts)
}

func (sfa *sourceFactoryAdapter) Metadata() PluginMetadata {
	meta := sfa.factory.Metadata()
	return PluginMetadata{
		Name:        meta.Name,
		Version:     meta.Version,
		Description: meta.Description,
		Author:      meta.Author,
		Type:        meta.Type,
		Features:    meta.Features,
	}
}

type validatorFactoryAdapter struct {
	factory validators.ValidatorFactory
}

func (vfa *validatorFactoryAdapter) Create() ValidatorPlugin {
	return &validatorPluginAdapter{plugin: vfa.factory.Create()}
}

func (vfa *validatorFactoryAdapter) Metadata() PluginMetadata {
	meta := vfa.factory.Metadata()
	return PluginMetadata{
		Name:        meta.Name,
		Version:     meta.Version,
		Description: meta.Description,
		Author:      meta.Author,
		Type:        meta.Type,
		Features:    meta.Features,
	}
}

type validatorPluginAdapter struct {
	plugin validators.ValidatorPlugin
}

func (vpa *validatorPluginAdapter) Name() string {
	return vpa.plugin.Name()
}

func (vpa *validatorPluginAdapter) Validate(field interface{}) error {
	return vpa.plugin.Validate(field)
}

func (vpa *validatorPluginAdapter) SupportedTypes() []string {
	return vpa.plugin.SupportedTypes()
}

// init registers built-in plugins
func init() {
	// Register source plugin factories
	RegisterSourceFactory("env", &sourceFactoryAdapter{factory: &sources.EnvFactory{}})
	RegisterSourceFactory("yaml", &sourceFactoryAdapter{factory: &sources.YAMLFactory{}})

	// Register validator plugin factories
	RegisterValidatorFactory("ip", &validatorFactoryAdapter{factory: &validators.IPValidatorFactory{}})
	RegisterValidatorFactory("uuid", &validatorFactoryAdapter{factory: &validators.UUIDValidatorFactory{}})
	RegisterValidatorFactory("creditcard", &validatorFactoryAdapter{factory: &validators.CreditCardValidatorFactory{}})
	RegisterValidatorFactory("phone", &validatorFactoryAdapter{factory: &validators.PhoneValidatorFactory{}})
}
