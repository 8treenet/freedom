package crud

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode"

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

type Table2Struct struct {
	dsn            string
	db             *sql.DB
	table          string
	prefix         string
	err            error
	realNameMethod string
	tagKey         string // tag字段的key值,默认是orm
}

func NewTable2Struct() *Table2Struct {
	return &Table2Struct{}
}

func (t *Table2Struct) Dsn(d string) *Table2Struct {
	t.dsn = d
	return t
}

func (t *Table2Struct) TagKey(r string) *Table2Struct {
	t.tagKey = r
	return t
}

func (t *Table2Struct) RealNameMethod(r string) *Table2Struct {
	t.realNameMethod = r
	return t
}

func (t *Table2Struct) DB(d *sql.DB) *Table2Struct {
	t.db = d
	return t
}

type Field struct {
	Column     string
	Type       string
	Value      string
	Arg        string
	StructName string
}

type SturctContent struct {
	Name          string
	TableRealName string
	Content       string
	Fields        []Field
	NumberFields  []Field
}

func (t *Table2Struct) Run() (result []SturctContent, e error) {
	// 链接mysql, 获取db对象
	t.dialMysql()
	if t.err != nil {
		e = t.err
		return
	}

	// 获取表和字段的shcema
	tableColumns, err := t.getColumns()
	if err != nil {
		e = err
		return
	}

	for tableRealName, item := range tableColumns {
		var structContent string
		// 去除前缀
		if t.prefix != "" {
			tableRealName = tableRealName[len(t.prefix):]
		}
		tableName := tableRealName
		sc := SturctContent{
			TableRealName: tableName,
			Fields:        make([]Field, 0),
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
			fieldItem := Field{
				Column:     column,
				Type:       v.Type,
				Value:      v.ColumnName,
				StructName: tableName,
				Arg:        lowerCamelCase(v.ColumnName),
			}
			sc.Fields = append(sc.Fields, fieldItem)
			if v.Type == "int" || v.Type == "float64" {
				sc.NumberFields = append(sc.NumberFields, fieldItem)
			}
		}
		structContent += tab(depth-1) + "}\n\n"

		t.realNameMethod = "TableName"
		// 添加 method 获取真实表名
		if t.realNameMethod != "" {
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

func (t *Table2Struct) dialMysql() {
	if t.db == nil {
		if t.dsn == "" {
			t.err = errors.New("dsn数据库配置缺失")
			return
		}
		t.db, t.err = sql.Open("mysql", t.dsn)
	}
	return
}

type column struct {
	ColumnName    string
	Type          string
	Nullable      string
	TableName     string
	ColumnComment string
	Tag           string
	Primary       string
}

// Function for fetching schema definition of passed table
func (t *Table2Struct) getColumns(table ...string) (tableColumns map[string][]column, err error) {
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
		err = rows.Scan(&col.ColumnName, &col.Type, &col.Nullable, &col.TableName, &col.ColumnComment, &col.Primary)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		//col.Json = strings.ToLower(col.ColumnName)
		col.Tag = col.ColumnName
		col.ColumnComment = col.ColumnComment
		col.ColumnName = t.camelCase(col.ColumnName)
		col.Type = typeForMysqlToGo[col.Type]
		if _, ok := tableColumns[col.TableName]; !ok {
			tableColumns[col.TableName] = []column{}
		}
		tableColumns[col.TableName] = append(tableColumns[col.TableName], col)
	}
	return
}

func (t *Table2Struct) camelCase(str string) string {
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
	// "ACL":   true,
	// "API":   true,
	// "ASCII": true,
	// "CPU":   true,
	// "CSS":   true,
	// "DNS":   true,
	// "EOF":   true,
	// "GUID":  true,
	// "HTML":  true,
	// "HTTP":  true,
	// "HTTPS": true,
	// "ID":    true,
	// "IP":    true,
	// "JSON":  true,
	// "LHS":   true,
	// "QPS":   true,
	// "RAM":   true,
	// "RHS":   true,
	// "RPC":   true,
	// "SLA":   true,
	// "SMTP":  true,
	// "SQL":   true,
	// "SSH":   true,
	// "TCP":   true,
	// "TLS":   true,
	// "TTL":   true,
	// "UDP":   true,
	// "UI":    true,
	// "UID":   true,
	// "UUID":  true,
	// "URI":   true,
	// "URL":   true,
	// "UTF8":  true,
	// "VM":    true,
	// "XML":   true,
	// "XMPP":  true,
	// "XSRF":  true,
	// "XSS":   true,
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
