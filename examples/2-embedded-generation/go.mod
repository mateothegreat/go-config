module go-generate-example

go 1.21

require github.com/mateothegreat/go-config v0.0.0-00010101000000-000000000000

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Use local version for development
replace github.com/mateothegreat/go-config => ../../
