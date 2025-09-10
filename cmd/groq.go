package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

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

func callGroq(input string) (ToolCallResponse, error) {
	// Get API key from environment variable
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("GROQ_API_KEY environment variable not set")
		return ToolCallResponse{}, fmt.Errorf("API not set")
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
	if err != nil {
		return ToolCallResponse{}, fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	fmt.Print(req)
	if err != nil {
		return ToolCallResponse{}, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return ToolCallResponse{}, fmt.Errorf("fail to get API key")
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return ToolCallResponse{}, fmt.Errorf("fail to get API key")
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
		return ToolCallResponse{}, fmt.Errorf("fail to get API key")
	}

	// Parse response
	var response ResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		return ToolCallResponse{}, fmt.Errorf("error unmarshaling response: %v\nRaw: %s", err, string(body))
	}
	if len(response.Choices) == 0 || len(response.Choices[0].Message.ToolCalls) == 0 {
		return ToolCallResponse{}, fmt.Errorf("no tool call found in response")
	}
	toolCall := response.Choices[0].Message.ToolCalls[0]
	fmt.Printf("Tool call received: %+v\n", toolCall)
	var parsed ToolCallResponse
	err = json.Unmarshal([]byte(toolCall.Function.Arguments), &parsed)
	if err != nil {
		return ToolCallResponse{}, fmt.Errorf("error parsing tool args: %v\nArgs: %s", err, toolCall.Function.Arguments)
	}

	return parsed, nil
}

// func callHttp(prompt string) string {
// 	_ = godotenv.Load()
// 	apiKey := os.Getenv("GROQ_API_KEY")
// 	return prompt
// }
