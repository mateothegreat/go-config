module go-generate-example

go 1.21

require gopkg.in/yaml.v3 v3.0.1

// Use local version for development
replace github.com/mateothegreat/go-config => ../../
