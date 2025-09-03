package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model               string    `json:"model"`
	Messages            []Message `json:"messages"`
	MaxCompletionTokens int       `json:"max_completion_tokens"`
	Temperature         float32   `json:"temperature"`
}

type ToolCall struct {
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
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

const UpdateSystemPrompt string = `
Analise a tarefa fornecida e extraia as seguintes informações:

   1. action: O tipo de ação a ser executada. As opções possíveis são:
        - add: Adicionar quantidades a um item existente ou criar um novo item.
        - subtract: Remover quantidades de um item.
        - discard_all Tudo: Remover todas as quantidades de um item.
        - rename: Alterar o nome de um item.
        - change_unit: Alterar a unidade de medida de um item.  
   2. item_name: O nome do item. Converta para substantivo no singular. Não traduza o nome nem resuma nem omita a marca caso o usuário forneça, apenas converta para singular.
   Se o usuário passar o nome do item em '' ou "" salve do jeito que ele passar, apenas removendo as ''/"".
   3. quantity: A quantidade, que pode ser um número inteiro ou decimal. Caso não seja especificado a quantidade atribua = 1.
   4. unit: A unidade da quantidade (por exemplo, "kg", "un", "m", "ft", "sq ft"). Caso não seja informada, defina como "un".
   5. category: A categoria a qual o item pertence. Caso não seja mencionado setar como "geral".
   6. location: O local do item, opcional. Caso não seja fornecido pode ser Nulo.
   7. description: A descrição do item, opcional. Caso não seja fornecido pode ser Nulo.

Somente para a ação rename, extraia também: 
   8. old_item_name: O nome atual do item (string ou None se não for informado). 
   9. new_item_name: O novo nome para o item (string ou None se não for informado).

o Schema dos dados segue:
    action: Optional[ActionOptions] = Field(description='Action required for the task: add, subtract, discard_all,rename')
    item_name: str = Field(description='Item da tarefa')
    quantity: Optional[Union[float, int]] = Field(description='Quantidade')
    unit: UnitOptions = Field(description='unidade de medida.')
    old_item_name: Optional[str] | None 
    new_item_name: Optional[str] | None 
    category: [str = Field(description='Category of the item') `

type ActionOptions string
type UnitOptions string

const (
	ActionAdd      ActionOptions = "add"
	ActionSubtract ActionOptions = "subtract"
	ActionDiscard  ActionOptions = "discard"

	UnitGrams     UnitOptions = "grams"
	UnitKilograms UnitOptions = "kilograms"
	UnitLiters    UnitOptions = "liters"
	UnitUnits     UnitOptions = "units"
	// Add more units as needed
)

type UpdateBaseModel struct {
	Action      *ActionOptions `json:"action,omitempty"`
	ItemName    string         `json:"item_name"`
	Quantity    interface{}    `json:"quantity,omitempty"` // Use interface{} for float|int
	Unit        UnitOptions    `json:"unit"`
	OldItemName *string        `json:"old_item_name,omitempty"`
	NewItemName *string        `json:"new_item_name,omitempty"`
	Category    *string        `json:"category,omitempty"`
	Description *string        `json:"description,omitempty"`
	Location    *string        `json:"location,omitempty"`
}

type UpdateRequestBody struct {
	Model               string      `json:"model"`
	Messages            []Message   `json:"messages"`
	MaxCompletionTokens int         `json:"max_completion_tokens"`
	Temperature         float32     `json:"temperature"`
	Tools               []Tool      `json:"tools,omitempty"`
	ToolChoice          interface{} `json:"tool_choice,omitempty"`
}

type FunctionDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type UpdateFunctionParameters struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}

var updateFunctionSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"action": map[string]interface{}{
			"type":        "string",
			"description": "Action required for the task: add, subtract, discard",
			"enum":        []string{"add", "subtract", "discard"},
		},
		"item_name": map[string]interface{}{
			"type":        "string",
			"description": "Item of the task",
		},
		"quantity": map[string]interface{}{
			"type":        "number",
			"description": "Quantity of the item in the task",
		},
		"unit": map[string]interface{}{
			"type":        "string",
			"description": "Unit of the items quantity",
			"enum":        []string{"grams", "kilograms", "liters", "units"},
		},
		"old_item_name": map[string]interface{}{
			"type":        "string",
			"description": "Previous name of the item if being renamed",
		},
		"new_item_name": map[string]interface{}{
			"type":        "string",
			"description": "New name of the item if being renamed",
		},
		"category": map[string]interface{}{
			"type":        "string",
			"description": "Category of the item",
		},
		"description": map[string]interface{}{
			"type":        "string",
			"description": "Description of the item",
		},
		"location": map[string]interface{}{
			"type":        "string",
			"description": "Location of the item",
		},
	},
	"required": []string{"item_name", "unit"},
}

func callGroq(input string) (ToolCall, error) {
	fmt.Print("\nolá1\n")
	fmt.Println("\nOla2")
	// Get API key from environment variable
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("GROQ_API_KEY environment variable not set")
		return ToolCall{}, fmt.Errorf("API not set")
	}

	// Create request body
	requestBody := UpdateRequestBody{
		Model: "openai/gpt-oss-120b",
		Messages: []Message{
			{
				Role:    "system",
				Content: UpdateSystemPrompt,
			},
			{
				Role:    "user",
				Content: input,
			},
		},
		MaxCompletionTokens: 4000,
		Temperature:         0.0,
		Tools: []Tool{
			{
				Type: "function",
				Function: FunctionDefinition{
					Name:        "update_inventory",
					Description: "Update inventory items with specific actions",
					Parameters:  updateFunctionSchema,
				},
			},
		},
		ToolChoice: map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name": "update_inventory",
			},
		},
	}

	// Marshal request body to JSON
	jsonData, err := json.Marshal(requestBody)
	fmt.Println(jsonData)
	if err != nil {
		return ToolCall{}, fmt.Errorf("error marshaling JSON: %v", err)
	}

	fmt.Println("Request JSON:", string(jsonData))
	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	fmt.Print(req)
	if err != nil {
		return ToolCall{}, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return ToolCall{}, fmt.Errorf("fail to get API key")
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return ToolCall{}, fmt.Errorf("fail to get API key")
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
		return ToolCall{}, fmt.Errorf("fail to get API key")
	}

	// Parse response
	var response ResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		return ToolCall{}, fmt.Errorf("error unmarshaling response: %v\nRaw: %s", err, string(body))
	}
	if len(response.Choices) == 0 || len(response.Choices[0].Message.ToolCalls) == 0 {
		return ToolCall{}, fmt.Errorf("no tool call found in response")
	}
	toolCall := response.Choices[0].Message.ToolCalls[0]
	fmt.Printf("Tool call received: %+v\n", toolCall)
	var parsed ToolCall
	err = json.Unmarshal([]byte(toolCall.Function.Arguments), &parsed)
	if err != nil {
		return ToolCall{}, fmt.Errorf("error parsing tool args: %v\nArgs: %s", err, toolCall.Function.Arguments)
	}

	return parsed, nil
}

// func callHttp(prompt string) string {
// 	_ = godotenv.Load()
// 	apiKey := os.Getenv("GROQ_API_KEY")
// 	return prompt
// }
