# Env
Env is a helper utility which attempts to unmarshal struct values from Environment variables

## Examples

### Unmarsal
```go 
func ExampleUnmarshal() {
	type config struct {
		A string `env:"a"`
		B string `env:"b"`
		C string `env:"c"`
	}

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}
}
```