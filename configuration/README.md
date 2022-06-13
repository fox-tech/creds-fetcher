# Configuration
Configuration is a helper utility which attempts load Configuration values from various sources.

## Examples

### New
```go

func ExampleNew() {
	cfg, err := New("")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("We have our configuration!", cfg)
}

```

### New (with override)
```go

func ExampleNew_with_override() {
	cfg, err := New("./path/to/config/config.json")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("We have our configuration!", cfg)
}
```

### Configuration.OverrideWith
```go
func ExampleConfiguration_OverrideWith() {
	cfg := Configuration{
		AWSProviderARN: "1",
		AWSRoleARN:     "2",
		OktaClientID:   "3",
		OktaAppID:      "4",
		OktaURL:        "5",
	}

	overrides := Configuration{
		AWSProviderARN: "1new",
		AWSRoleARN:     "2new",
	}

	cfg.OverrideWith(&overrides)

	fmt.Printf("Updated values: %+v\n", cfg)
}
```

### Configuration.Validate
```go
func ExampleConfiguration_Validate() {
	cfg := Configuration{
		AWSProviderARN: "1",
		AWSRoleARN:     "2",
		OktaClientID:   "3",
		OktaURL:        "4",
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
}
```
