package crud

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"unicode"

	//github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

//map for converting mysql type to golang types
var typeForMysqlToGo = map[string]string{
	"int":                "int",
	"integer":            "int",
	"tinyint":            "int",
	"smallint":           "int",
	"mediumint":          "int",
	"bigint":             "int",
	"int unsigned":       "int",
	"integer unsigned":   "int",
	"tinyint unsigned":   "int",
	"smallint unsigned":  "int",
	"mediumint unsigned": "int",
	"bigint unsigned":    "int",
	"bit":                "int",
	"bool":               "bool",
	"enum":               "string",
	"set":                "string",
	"varchar":            "string",
	"char":               "string",
	"tinytext":           "string",
	"mediumtext":         "string",
	"text":               "string",
	"longtext":           "string",
	"blob":               "string",
	"tinyblob":           "string",
	"mediumblob":         "string",
	"longblob":           "string",
	"date":               "time.Time", // time.Time
	"datetime":           "time.Time", // time.Time
	"timestamp":          "time.Time", // time.Time
	"time":               "time.Time", // time.Time
	"float":              "float64",
	"double":             "float64",
	"decimal":            "float64",
	"binary":             "string",
	"varbinary":          "string",
}

// Generate .
type Generate struct {
	dsn            string
	db             *sql.DB
	table          string
	prefix         string
	realNameMethod string
}

// NewGenerate .
func NewGenerate() *Generate {
	return &Generate{}
}

// Dsn .
func (t *Generate) Dsn(d string) *Generate {
	t.dsn = d
	return t
}

// SetPrefix .
func (t *Generate) SetPrefix(prefix string) *Generate {
	t.prefix = prefix
	return t
}

// Method .
//func (obj *{{.ObjectName}}) Set{{.Name}} ({{.Variable}} {{.VariableType}}) {
//	obj.{{.Name}} = {{.Variable}}
//	obj.setChanges("{{.Column}}", {{.Variable}})
//}
//{{ end }}
type Method struct {
	ObjectName   string //结构体名称
	Name         string //方法名称 GoodsName
	VariableType string //类型 int|string
	Variable     string //goodsName
	Column       string //列名 Column id | goods_name
}

// ObjectContent .
type ObjectContent struct {
	Name          string   //结构体名称
	TableRealName string   //表名
	Content       string   //结构体内容
	SetMethods    []Method //set的方法
	AddMethods    []Method //add的方法
}

// RunDsn .
func (t *Generate) RunDsn() (result []ObjectContent, e error) {
	if e = t.dialMysql(); e != nil {
		return
	}

	// 获取表和字段的schema
	tableColumns, err := t.getColumns()
	if err != nil {
		e = err
		return
	}

	return t.schema(tableColumns), nil
}

// RunJSON .
func (t *Generate) RunJSON(jsonFileName string) (result []ObjectContent, e error) {
	buffer, err := ioutil.ReadFile(jsonFileName)
	if err != nil {
		e = err
		return
	}
	var tables []interface{}
	json.Unmarshal(buffer, &tables)

	// 获取表和字段的schema
	tableColumns, err := t.getJSONColumns(tables)
	if err != nil {
		e = err
		return
	}
	return t.schema(tableColumns), nil
}

func (t *Generate) schema(tableColumns map[string][]column) (result []ObjectContent) {
	for tableRealName, item := range tableColumns {
		var structContent string
		tableName := tableRealName
		sc := ObjectContent{
			TableRealName: tableName,
			SetMethods:    make([]Method, 0),
		}
		// 去除前缀
		if t.prefix != "" {
			tableName = tableRealName[len(t.prefix):]
		}

		switch len(tableName) {
		case 0:
		case 1:
			tableName = strings.ToUpper(tableName[0:1])
		default:
			// 字符长度大于1时
			tableName = strings.ToUpper(tableName[0:1]) + tableName[1:]
		}

		depth := 1
		tableName = t.camelCase(tableName)
		sc.Name = tableName
		structContent += "// " + tableName + " .\n"
		structContent += "type " + tableName + " struct {\n"
		structContent += "	changes map[string]interface{}\n"
		for _, v := range item {
			column := v.Tag
			if v.Primary == "PRI" {
				v.Tag = "`" + `gorm:"primary_key;column:` + v.Tag + `"` + "`"
			} else {
				v.Tag = "`" + `gorm:"column:` + v.Tag + `"` + "`"
			}
			//structContent += tab(depth) + v.ColumnName + " " + v.Type + " " + v.Json + "\n"
			// 字段注释
			var clumnComment string
			if v.ColumnComment != "" {
				clumnComment = fmt.Sprintf(" // %s", v.ColumnComment)
			}
			structContent += fmt.Sprintf("%s%s %s %s%s\n",
				tab(depth), v.ColumnName, v.Type, v.Tag, clumnComment)

			if v.Primary == "PRI" {
				continue
			}
			fieldItem := Method{
				Column:       column,
				VariableType: v.Type,
				Name:         v.ColumnName,
				ObjectName:   tableName,
				Variable:     lowerCamelCase(v.ColumnName),
			}
			sc.SetMethods = append(sc.SetMethods, fieldItem)
			if v.Type == "int" || v.Type == "float64" {
				sc.AddMethods = append(sc.AddMethods, fieldItem)
			}
		}
		structContent += tab(depth-1) + "}\n\n"

		t.realNameMethod = "TableName"
		// 添加 method 获取真实表名
		if t.realNameMethod != "" {
			structContent += "// " + t.realNameMethod + " .\n"
			structContent += fmt.Sprintf("func (obj *%s) %s() string {\n",
				tableName, t.realNameMethod)
			structContent += fmt.Sprintf("%sreturn \"%s\"\n",
				tab(depth), tableRealName)
			structContent += "}\n\n"
		}
		sc.Content = structContent
		result = append(result, sc)
	}
	return
}

func (t *Generate) dialMysql() (e error) {
	if t.db == nil {
		if t.dsn == "" {
			return errors.New("dsn数据库配置缺失")
		}
		t.db, e = sql.Open("mysql", t.dsn)
	}
	return
}

type column struct {
	ColumnName    string
	Type          string
	TableName     string
	ColumnComment string
	Tag           string
	Primary       string
}

// Function for fetching schema definition of passed table
func (t *Generate) getColumns(table ...string) (tableColumns map[string][]column, err error) {
	tableColumns = make(map[string][]column)
	// sql
	var sqlStr = `SELECT COLUMN_NAME,DATA_TYPE,IS_NULLABLE,TABLE_NAME,COLUMN_COMMENT,COLUMN_KEY
		FROM information_schema.COLUMNS 
		WHERE table_schema = DATABASE()`
	// 是否指定了具体的table
	if t.table != "" {
		sqlStr += fmt.Sprintf(" AND TABLE_NAME = '%s'", t.prefix+t.table)
	}
	// sql排序
	sqlStr += " order by TABLE_NAME asc, ORDINAL_POSITION asc"

	rows, err := t.db.Query(sqlStr)
	if err != nil {
		fmt.Println("Error reading table information: ", err.Error())
		return
	}

	defer rows.Close()

	for rows.Next() {
		col := column{}
		var Nullable string
		err = rows.Scan(&col.ColumnName, &col.Type, &Nullable, &col.TableName, &col.ColumnComment, &col.Primary)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		//col.Json = strings.ToLower(col.ColumnName)
		col.Tag = col.ColumnName
		col.ColumnName = t.camelCase(col.ColumnName)
		col.Type = typeForMysqlToGo[col.Type]
		if _, ok := tableColumns[col.TableName]; !ok {
			tableColumns[col.TableName] = []column{}
		}
		tableColumns[col.TableName] = append(tableColumns[col.TableName], col)
	}
	return
}

// Function for fetching schema definition of passed table
func (t *Generate) getJSONColumns(jsonData []interface{}) (tableColumns map[string][]column, err error) {
	tableColumns = make(map[string][]column)
	for _, tableObjData := range jsonData {
		tableObj := tableObjData.(map[string]interface{})
		tableName := tableObj["tableName"].(string)
		primaryKey, _ := tableObj["primaryKey"].(string)

		for key, value := range tableObj {
			if !strings.Contains(key, "columns") {
				continue
			}
			columnType := strings.Split(key, ":")[1]
			if newType, ok := typeForMysqlToGo[columnType]; ok {
				columnType = newType
			}

			for _, columnInterface := range value.([]interface{}) {
				columnName := columnInterface.(string)
				col := column{}
				col.Tag = columnName
				col.ColumnName = t.camelCase(columnName)
				col.Type = columnType
				col.TableName = tableName
				if columnName == primaryKey {
					col.Primary = "PRI"
				}
				tableColumns[tableName] = append(tableColumns[tableName], col)
			}
		}
	}
	return
}

func (t *Generate) camelCase(str string) string {
	// 是否有表前缀, 设置了就先去除表前缀
	if t.prefix != "" {
		str = strings.Replace(str, t.prefix, "", 1)
	}
	var text string
	//for _, p := range strings.Split(name, "_") {
	for _, p := range strings.Split(str, "_") {
		// 字段首字母大写的同时, 是否要把其他字母转换为小写
		switch len(p) {
		case 0:
		case 1:
			text += strings.ToUpper(p[0:1])
		default:
			// 字符长度大于1时
			text += strings.ToUpper(p[0:1]) + p[1:]
		}
	}
	return lintName(text)
}
func tab(depth int) string {
	return strings.Repeat("\t", depth)
}

// lintName returns a different name if it should be different.
func lintName(name string) (should string) {
	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}
	allLower := true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}
	if allLower {
		return name
	}

	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word
		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); commonInitialisms[u] {
			// Keep consistent case, which is lowercase only at the start.
			if w == 0 && unicode.IsLower(runes[w]) {
				u = strings.ToLower(u)
			}
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))
		} else if w > 0 && strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}

var commonInitialisms = map[string]bool{
	"ACL":   true,
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XMPP":  true,
	"XSRF":  true,
	"XSS":   true,
}

func lowerCamelCase(field string) string {
	var lowerStr string
	vv := []rune(field)
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			lowerStr += strings.ToLower(string(vv[i]))
		} else {
			lowerStr += string(vv[i])
		}
	}
	return lowerStr
}

func isNumber(columnType string) bool {
	switch columnType {
	case "uint8", "uint16", "uint32", "uint64", "uint":
		return true
	case "int8", "int16", "int32", "int64", "int":
		return true
	case "float32", "float64":
		return true
	}
	return false
}
