package schema

type AddMemoryRequestBody struct {
	Messages []OpenAIMessage `json:"messages"`
}

type QueryMemoryRequestBody struct {
	Query    string `json:"query"`
	Strategy string `json:"strategy"`
}
