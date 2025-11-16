package source

import (
	"context"
	"fmt"
	"os"

	"github.com/ju4n97/relic/internal/config"
)

// Downloader downloads a model to local cache.
type Downloader interface {
	Download(ctx context.Context, modelConfig *config.ModelConfig, targetDir string) (string, error)
}

// registry maps source types to their downloader.
var registry = map[config.SourceType]Downloader{
	config.SourceTypeHuggingFace: &HuggingFaceDownloader{},
}

// EnsureModelsDirectory ensures that the models directory exists.
func EnsureModelsDirectory(targetDir string) error {
	err := os.MkdirAll(targetDir, 0o755)
	if err != nil {
		return fmt.Errorf("manager: failed to create models directory: %w", err)
	}

	return nil
}

// GetDownloader returns the downloader for the given source type.
func GetDownloader(ctx context.Context, sourceType config.SourceType) (Downloader, error) {
	downloader, ok := registry[sourceType]
	if !ok {
		return nil, fmt.Errorf("unknown source type: %s", sourceType)
	}

	return downloader, nil
}
