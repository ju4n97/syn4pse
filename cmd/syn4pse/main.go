package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	syn4psegrpc "github.com/ju4n97/syn4pse/api/grpc"
	syn4psehttp "github.com/ju4n97/syn4pse/api/http"
	"github.com/ju4n97/syn4pse/internal/backend"
	"github.com/ju4n97/syn4pse/internal/backend/llama"
	"github.com/ju4n97/syn4pse/internal/backend/piper"
	"github.com/ju4n97/syn4pse/internal/backend/whisper"
	"github.com/ju4n97/syn4pse/internal/config"
	"github.com/ju4n97/syn4pse/internal/env"
	"github.com/ju4n97/syn4pse/internal/logger"
	"github.com/ju4n97/syn4pse/internal/model"
	"github.com/ju4n97/syn4pse/internal/service"
	inferencev1 "github.com/ju4n97/syn4pse/sdk-go/pb/inference/v1"
)

func main() {
	ctx := context.Background()

	var (
		flagHTTPPort   = flag.Int("http-port", config.DefaultHTTPPort(), "HTTP port to listen on")
		flagGRPCPort   = flag.Int("grpc-port", config.DefaultGRPCPort(), "gRPC port to listen on")
		flagConfigPath = flag.String("config", path.Join(config.DefaultConfigPath(), "config.yaml"), "Path to config file")
		flagSchemaPath = flag.String("schema", path.Join(config.DefaultConfigPath(), "syn4pse.v1.schema.json"), "Path to schema file")
		flagLlamaBin   = flag.String("llama-bin", "./bin/llama-server-cuda", "Path to llama")
		flagWhisperBin = flag.String("whisper-bin", "./bin/whisper-server-cuda", "Path to whisper")
		flagPiperBin   = flag.String("piper-bin", "./bin/piper-cpu/piper", "Path to piper")
	)
	flag.Parse()

	environment := env.FromEnv()

	slog.SetDefault(
		logger.New(environment,
			logger.WithLogToFile(true),
			logger.WithLogFile("logs/syn4pse.log"),
		),
	)

	modelManager := model.NewManager()

	watcher, err := config.NewWatcher(*flagConfigPath, *flagSchemaPath, func(cfg *config.Config, err error) {
		if err != nil {
			slog.Error("Failed to reload config", "error", err)
			return
		}

		if err := modelManager.LoadModelsFromConfig(ctx, cfg); err != nil {
			slog.Error("Failed to load models from config", "error", err)
			return
		}
	})
	if err != nil {
		slog.Error("Failed to create config watcher", "error", err)
		return
	}

	cfg := watcher.Snapshot()
	if err := modelManager.LoadModelsFromConfig(ctx, cfg); err != nil {
		slog.Error("Failed to load models from config", "error", err)
		return
	}

	slog.Info("Config loaded successfully", "config", *flagConfigPath, "schema", *flagSchemaPath)

	backends := backend.NewRegistry()
	defer func() {
		if err := backends.Close(); err != nil {
			slog.Error("Failed to close backends", "error", err)
		}
	}()

	serverManager := backend.NewServerManager()
	defer serverManager.StopAll()

	backendLlama, err := llama.NewBackend(*flagLlamaBin, serverManager)
	if err != nil {
		slog.Error("Failed to create Llama backend", "error", err)
	}
	if err := backends.Register(backendLlama); err != nil {
		slog.Error("Failed to register Llama backend", "error", err)
	}

	backendWhisper, err := whisper.NewBackend(*flagWhisperBin, serverManager)
	if err != nil {
		slog.Error("Failed to create Whisper backend", "error", err)
	}
	if err := backends.Register(backendWhisper); err != nil {
		slog.Error("Failed to register Whisper backend", "error", err)
	}

	backendPiper, err := piper.NewBackend(*flagPiperBin)
	if err != nil {
		slog.Error("Failed to create Piper backend", "error", err)
	}
	if err := backends.Register(backendPiper); err != nil {
		slog.Error("Failed to register Piper backend", "error", err)
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	httpServer := buildHTTPServer(*flagHTTPPort, backends, modelManager.Registry())
	grpcServer := buildGRPCServer(backends, modelManager.Registry())

	g.Go(func() error {
		slog.Info("Starting HTTP server", "port", *flagHTTPPort)
		return runHTTPServer(ctx, httpServer)
	})

	g.Go(func() error {
		slog.Info("Starting gRPC server", "port", *flagGRPCPort)
		return runGRPCServer(ctx, grpcServer, *flagGRPCPort)
	})

	if err := g.Wait(); err != nil {
		slog.Error("Error running servers", "error", err)
	}

	slog.Info("Shutting down...")
}

// runHTTPServer runs the HTTP server.
func runHTTPServer(ctx context.Context, server *http.Server) error {
	go func() {
		<-ctx.Done()
		slog.Info("Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Error shutting down HTTP server", "error", err)
		}
	}()

	slog.Info("Server starting",
		"protocol", "HTTP",
		"address", "http://localhost"+server.Addr,
		"docs_v1", fmt.Sprintf("http://localhost%s/v1/docs", server.Addr),
	)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		if ctx.Err() == nil {
			return fmt.Errorf("HTTP server error: %w", err)
		}
	}

	return nil
}

// runGRPCServer runs the gRPC server.
func runGRPCServer(ctx context.Context, server *grpc.Server, port int) error {
	addr := fmt.Sprintf(":%d", port)

	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("manager: failed to listen on %s: %w", addr, err)
	}

	go func() {
		<-ctx.Done()
		slog.Info("Gracefully shutting down gRPC server...")
		server.GracefulStop()
	}()

	slog.Info("Server starting",
		"protocol", "gRPC",
		"address", "grpc://localhost"+addr,
	)

	if err := server.Serve(listener); err != nil {
		if ctx.Err() == nil {
			return fmt.Errorf("gRPC server error: %w", err)
		}
	}

	return nil
}

// buildHTTPServer builds the HTTP server.
func buildHTTPServer(port int, backends *backend.Registry, models *model.Registry) *http.Server {
	router := buildHTTPRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("syn4pse HTTP service is running."))
	})

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	router.Route("/v1", func(r chi.Router) {
		cfg := huma.DefaultConfig("SYN4PSE", "1.0.0")
		cfg.Servers = []*huma.Server{{URL: "/v1"}}
		api := humachi.New(r, cfg)

		llm := service.NewLLM(backends, models)
		stt := service.NewSTT(backends, models)
		tts := service.NewTTS(backends, models)

		syn4psehttp.NewLLMHandler(api, llm)
		syn4psehttp.NewSTTHandler(api, stt)
		syn4psehttp.NewTTSHandler(api, tts)
	})

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
}

// buildGRPCServer builds the gRPC server.
func buildGRPCServer(backends *backend.Registry, models *model.Registry) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			unaryLoggingInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			streamLoggingInterceptor(),
		),
	)

	inferenceServer := syn4psegrpc.NewInferenceServer(backends, models)
	inferencev1.RegisterInferenceServiceServer(server, inferenceServer)

	// Enable reflection for development (allows using grpcurl, grpcui, etc.)
	if env.FromEnv() == env.EnvDevelopment {
		reflection.Register(server)
	}

	return server
}

// buildHTTPRouter builds the HTTP router.
func buildHTTPRouter() *chi.Mux {
	router := chi.NewMux()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.Use(
		middleware.RequestID,
		middleware.RealIP,
		httplog.RequestLogger(slog.Default(), &httplog.Options{
			Level:         slog.LevelInfo,
			Schema:        httplog.SchemaECS,
			RecoverPanics: true,
			LogExtraAttrs: func(req *http.Request, reqBody string, respStatus int) []slog.Attr {
				reqID := middleware.GetReqID(req.Context())
				realIP := req.RemoteAddr

				return []slog.Attr{
					slog.String("request_id", reqID),
					slog.String("real_ip", realIP),
				}
			},
		}),
		middleware.Recoverer,
		middleware.Compress(5),
		middleware.Timeout(60*time.Second),
	)
	return router
}

// unaryLoggingInterceptor logs unary RPC calls.
func unaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			slog.ErrorContext(ctx, "gRPC unary call failed",
				"method", info.FullMethod,
				"duration", duration,
				"error", err,
			)
		} else {
			slog.InfoContext(ctx, "gRPC unary call",
				"method", info.FullMethod,
				"duration", duration,
			)
		}

		return resp, err
	}
}

// streamLoggingInterceptor logs streaming RPC calls.
func streamLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		err := handler(srv, ss)

		duration := time.Since(start)
		if err != nil {
			slog.ErrorContext(ss.Context(), "gRPC stream call failed",
				"method", info.FullMethod,
				"duration", duration,
				"is_client_stream", info.IsClientStream,
				"is_server_stream", info.IsServerStream,
				"error", err,
			)
		} else {
			slog.InfoContext(ss.Context(), "gRPC stream call",
				"method", info.FullMethod,
				"duration", duration,
				"is_client_stream", info.IsClientStream,
				"is_server_stream", info.IsServerStream,
			)
		}

		return err
	}
}
