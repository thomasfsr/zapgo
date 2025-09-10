package main

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// type RequestBody struct {
// 	Model               string    `json:"model"`
// 	Messages            []Message `json:"messages"`
// 	MaxCompletionTokens int       `json:"max_completion_tokens"`
// 	Temperature         float32   `json:"temperature"`
// }

type RequestBodyWithTool struct {
	Model               string      `json:"model"`
	Messages            []Message   `json:"messages"`
	MaxCompletionTokens int         `json:"max_completion_tokens"`
	Temperature         float32     `json:"temperature"`
	Tools               []Tool      `json:"tools,omitempty"`
	ToolChoice          interface{} `json:"tool_choice,omitempty"`
}

type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type FunctionDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type UpdateFunctionParameters struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}

type ToolCallResponse struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type ResponseBody struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role      string             `json:"role"`
			Content   string             `json:"content"`
			ToolCalls []ToolCallResponse `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}
