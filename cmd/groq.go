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
Você é um assistente responsável por interpretar tarefas de inventário e gerar uma chamada de função para atualizar o estoque.

Analise a tarefa fornecida pelo usuário e extraia as seguintes informações para preencher a função "update_inventory":

1. action: Tipo de ação a ser executada. Valores válidos:
   - add: Adicionar quantidades a um item existente ou criar um novo item.
   - subtract: Remover quantidades de um item.
   - discard: Remover todas as quantidades de um item.
   - rename: Alterar o nome de um item.

2. item_name: O nome do item. Mantenha a marca e a nomenclatura fornecida pelo usuário. Converta para singular, exceto se o usuário usar aspas simples ou duplas ('' ou "").

3. quantity: Quantidade do item. Se não for especificada, atribua 1.

4. unit: Unidade da quantidade. Valores válidos: "grams", "kilograms", "liters", "units". Se não informado, defina "units".

5. category: Categoria do item. Se não fornecida, defina como "geral".

6. location: Local do item, opcional. Se não fornecido, pode ser nulo.

7. description: Descrição do item, opcional. Se não fornecido, pode ser nulo.

8. old_item_name e new_item_name: Apenas se a ação for "rename". Se não informados, podem ser nulos.

**Instruções importantes:**
- Retorne **somente** um JSON válido compatível com o schema da função "update_inventory".
- Não inclua explicações ou texto adicional fora do JSON.
- Preencha todos os campos obrigatórios do schema ("item_name" e "unit") e, se possível, os opcionais.

Exemplo:
Usuário: "subtraia 1 kilograma de arroz"
Saída JSON esperada:
{
  "action": "subtract",
  "item_name": "arroz",
  "quantity": 1,
  "unit": "kilograms",
  "category": "geral",
  "location": null,
  "description": null,
  "old_item_name": null,
  "new_item_name": null
}
 `

func callGroq(input string) (ToolCallResponse, error) {
	// Get API key from environment variable
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("GROQ_API_KEY environment variable not set")
		return ToolCallResponse{}, fmt.Errorf("API not set")
	}

	// Create request body
	requestBody := RequestBodyWithTool{
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
