package envvar

const (
	// RelicEnv is the environment variable used to determine the environment.
	RelicEnv = "RELIC_ENV"

	// RelicServerHTTPPort is the environment variable used to determine the HTTP port.
	RelicServerHTTPPort = "RELIC_SERVER_HTTP_PORT"

	// RelicServerGRPCPort is the environment variable used to determine the gRPC port.
	RelicServerGRPCPort = "RELIC_SERVER_GRPC_PORT"

	// RelicModelsPath is the environment variable used to determine the path to the models.
	RelicModelsPath = "RELIC_MODELS_PATH"

	// RelicConfigPath is the environment variable used to determine the path to the config.
	RelicConfigPath = "RELIC_CONFIG_PATH"
)
