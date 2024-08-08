# Config Managemer

This package provides a simple way to manage configuration for a Go application. It allows you to read configuration from multiple sources and merge them into a single struct.

## Example

```go
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

config, err := GetConfig[ExtraConfig](GetConfigArgs{
  Paths: []string{
    ".env.json",
    ".env.base.yaml",
    ".env.local.yaml",
  },
  WalkDepth: 7,
})

println(config.Foo.Bar)
```

## Usage

| Method       | Description                                                                     |
| ------------ | ------------------------------------------------------------------------------- |
| `GetConfig`  | Get the configuration from the given paths and merge them into a single struct. |
| `SaveConfig` | Save the configuration to the given path.                                       |

### `GetConfig`

```go
GetConfig[T any](args GetConfigArgs) (T, error)
```

Returns a new instance of the configuration struct with the configuration merged from the given paths.

| Argument | Type            | Description                                                                |
| -------- | --------------- | -------------------------------------------------------------------------- |
| `T`      | `interface{}`   | The type of the configuration to get. Must be a struct.                    |
| `args`   | `GetConfigArgs` | An arguments [struct](config/config.go#L14) to get the configuration with. |

### `SaveConfig`

```go
SaveConfig(config interface{}, path string, walkDepth int) error
```

Saves the configuration to the given path.

| Argument    | Type          | Description                                                |
| ----------- | ------------- | ---------------------------------------------------------- |
| `config`    | `interface{}` | The configuration to save.                                 |
| `path`      | `string`      | The path to save the configuration to.                     |
| `walkDepth` | `int`         | The depth to walk the path to find the configuration file. |
