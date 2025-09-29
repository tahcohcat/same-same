package embedders

type Embedder interface {
	Embed(text string) ([]float64, error)
	Name() string
}
