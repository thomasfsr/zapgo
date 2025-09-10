package main

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
