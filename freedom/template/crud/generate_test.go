package crud

import (
	"encoding/json"
	"testing"
)

func TestJSONGenerate(t *testing.T) {
	data := []byte(
		`
	[{
		"tableName": "fuck_user",
		"primaryKey": "userId",
		"columns:int": ["userId", "age"],
		"columns:int8": ["sex"],
		"columns:string": ["name", "address"],
		"columns:timestamp": ["created", "updated"]
	},
	{
		"tableName": "fuck_admin",
		"primaryKey": "id",
		"columns:int": ["id", "age", "role_id"],
		"columns:string": ["name", "address"],
		"columns:timestamp": ["created", "updated"]
	}
]`)

	var tables []interface{}
	json.Unmarshal(data, &tables)

	cmd := NewGenerate()
	cmd.prefix = "fuck_"
	tableColumns, _ := cmd.getJSONColumns(tables)

	t.Log(cmd.shcema(tableColumns))
}
