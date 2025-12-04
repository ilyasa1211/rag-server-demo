package configloader

import (
	"fmt"

	"github.com/ilyasa1211/rag-server/internal/domain"
	"github.com/spf13/viper"
)

func NewViperConfig(name string) (domain.Config, error) {
	viper.SetConfigFile(name)

	if err := viper.ReadInConfig(); err != nil {
		return domain.Config{}, fmt.Errorf("failed to retrieve config: %w", err)
	}

	return domain.Config{
		App: struct{ Listen string }{
			Listen: viper.GetString("app.listen"),
		},
		Milvus: struct {
			Address string
			ApiKey  string
		}{
			Address: viper.GetString("milvus.address"),
			ApiKey:  viper.GetString("milvus.apiKey"),
		},
		OpenAI: struct {
			BaseURL string
			Token   string
			Model   string
		}{
			BaseURL: viper.GetString("openai.baseUrl"),
			Token:   viper.GetString("openai.token"),
			Model:   viper.GetString("openai.model"),
		},
		Embedding: struct {
			BaseURL   string
			Token     string
			Model     string
			Dimension int
		}{
			BaseURL:   viper.GetString("embedding.baseUrl"),
			Token:     viper.GetString("embedding.token"),
			Model:     viper.GetString("embedding.model"),
			Dimension: viper.GetInt("embedding.dimension"),
		},
	}, nil
}
