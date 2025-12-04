package domain

type Config struct {
	App struct {
		Listen string
	}
	Milvus struct {
		Address string
		ApiKey  string
	}
	OpenAI struct {
		BaseURL string
		Token   string
		Model   string
	}
	Embedding struct {
		BaseURL   string
		Token     string
		Model     string
		Dimension int
	}
}
