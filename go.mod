module github.com/ju4n97/syn4pse

go 1.25.4

require (
	github.com/danielgtaylor/huma/v2 v2.34.1
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-chi/cors v1.2.2
	github.com/go-chi/httplog/v3 v3.3.0
	github.com/ju4n97/syn4pse/sdk-go v0.0.0-20251116022054-a59e331016fe
	github.com/lmittmann/tint v1.1.2
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
	github.com/stretchr/testify v1.11.1
	go.yaml.in/yaml/v3 v3.0.4
	golang.org/x/sync v0.18.0
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/ju4n97/syn4pse/sdk-go => ./sdk-go
