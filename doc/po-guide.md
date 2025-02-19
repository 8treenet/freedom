# Freedom PO 指南

## 简介

`freedom new-po` 是 Freedom 框架提供的一个强大的代码生成工具，用于自动生成持久化对象（Persistent Object，PO）和相关的 CRUD 操作代码。本工具可以显著提高开发效率，确保代码质量的一致性。

## 功能特性

- 自动生成与数据库表结构对应的 Go 结构体
- 生成标准的 CRUD 操作方法
- 支持多种数据源（MySQL 数据库和 JSON Schema）
- 自动处理字段命名转换（下划线到驼峰）
- 内置智能类型映射
- 支持主键和索引识别
- 生成 Gorm 标签
- 自动生成字段更新方法

## 使用方法

### 1. 通过数据库连接生成

```bash
freedom new-po --dsn "username:password@tcp(host:port)/database_name?charset=utf8"
```

参数说明：
- username: 数据库用户名
- password: 数据库密码
- host: 数据库主机地址
- port: 数据库端口
- database_name: 数据库名称

示例：
```bash
freedom new-po --dsn "root:123456@tcp(127.0.0.1:3306)/myapp?charset=utf8"
```

### 2. 通过 JSON Schema 生成

```bash
freedom new-po --json ./path/to/schema.json
```

JSON Schema 文件格式示例：
```json
[
  {
    "tableName": "users",
    "primaryKey": "id",
    "columns:int": ["id", "age", "status"],
    "columns:string": ["name", "email", "phone"],
    "columns:time.Time": ["created_at", "updated_at"]
  },
  {
    "tableName": "orders",
    "primaryKey": "order_id",
    "columns:int": ["order_id", "user_id", "amount"],
    "columns:string": ["order_no", "status"],
    "columns:time.Time": ["created_at"]
  }
]
```

### 3. 更多选项

```bash
freedom new-po -h
```

常用选项：
- `--prefix`: 设置表名前缀，生成代码时会自动去除
- `--dir`: 指定生成文件的目录
- `--package`: 指定生成代码的包名

## 生成的文件结构

执行命令后，将会生成如下文件：

```
├── po/
│   ├── user.go          # 用户表对应的 PO 结构体
│   ├── order.go         # 订单表对应的 PO 结构体
│   └── ...
└── repository/
    ├── user_repo.go     # 用户表的仓储实现
    ├── order_repo.go    # 订单表的仓储实现
    ├── generate.go      # 生成的仓储方法
    └── ...
```

## 生成的代码详解

### 1. PO 对象结构

每个生成的 PO 对象都包含以下特性：

```go
// 以 User 表为例
type User struct {
    // 1. 数据库字段映射
    ID        int       `gorm:"primaryKey;column:id"`
    Name      string    `gorm:"column:name"`
    Email     string    `gorm:"column:email"`
    CreatedAt time.Time `gorm:"column:created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at"`
    
    // 2. 变更跟踪字段
    changes   map[string]interface{}
}

// 3. 表名方法
func (obj *User) TableName() string {
    return "users"
}

// 4. 获取变更内容
func (obj *User) GetChanges() map[string]interface{} {
    if obj.changes == nil {
        return nil
    }
    result := make(map[string]interface{})
    for k, v := range obj.changes {
        result[k] = v
    }
    obj.changes = nil
    return result
}

// 5. 更新字段方法
func (obj *User) Update(name string, value interface{}) {
    if obj.changes == nil {
        obj.changes = make(map[string]interface{})
    }
    obj.changes[name] = value
}
```

### 2. 字段访问方法

每个字段都会生成对应的 Set 和 Add 方法（数值类型）：

```go
// Set 方法 - 适用于所有类型
func (obj *User) SetName(name string) {
    obj.Name = name
    obj.Update("name", name)
}

func (obj *User) SetEmail(email string) {
    obj.Email = email
    obj.Update("email", email)
}

// Add 方法 - 仅适用于数值类型（如 int, float64）
func (obj *User) AddAge(age int) {
    obj.Age += age
    obj.Update("age", gorm.Expr("age + ?", age))
}

func (obj *User) AddBalance(amount float64) {
    obj.Balance += amount
    obj.Update("balance", gorm.Expr("balance + ?", amount))
}
```

### 3. 仓储查询方法

生成的仓储方法包含丰富的查询选项：

```go
// 1. 基础查询方法
func findUser(repo GORMRepository, result *po.User, builders ...Builder) error
func findUserByID(repo GORMRepository, id int) (po.User, error)
func findUserListByID(repo GORMRepository, ids ...int) ([]*po.User, error)

// 2. 条件查询方法
func findUserByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (po.User, error)
func findUserByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (po.User, error)

// 3. 列表查询方法
func findUserList(repo GORMRepository, query po.User, builders ...Builder) ([]*po.User, error)
func findUserListByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) ([]*po.User, error)
func findUserListByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) ([]*po.User, error)

// 4. 创建和保存方法
func createUser(repo GORMRepository, object *po.User) (rowsAffected int64, error)
func saveUser(repo GORMRepository, object saveObject) (rowsAffected int64, error)

// 5. 辅助方法
func UserListToMap(list []*po.User, inErr error) (map[int]*po.User, error)
func UserToPoint(object po.User, inErr error) (*po.User, error)
```

### 4. Builder 接口

框架提供了两种内置的 Builder 实现，用于分页和限制查询结果：

```go
// Pager - 分页查询
pager := NewDescPager("created_at") // 按创建时间降序
pager.SetPage(1, 10) // 第1页，每页10条
users, err := findUserList(repo, po.User{}, pager)

// Limiter - 限制结果数量
limiter := NewDescLimiter("created_at", 5) // 获取最新5条记录
users, err := findUserList(repo, po.User{}, limiter)
```

### 5. Repository 层使用示例

生成的方法都在 repository 包中，需要在 repository 层中使用。以下是完整的使用示例：

```go
// repository/user_repository.go
type UserRepository struct {
    Worker freedom.Worker //依赖倒置
}

func NewUserRepository(worker freedom.Worker) *UserRepository {
    return &UserRepository{worker: worker}
}

// 实现 GORMRepository 接口
func (repo *UserRepository) db() *gorm.DB {
    return repo.worker.DB()
}

func (repo *UserRepository) Worker() freedom.Worker {
    return repo.worker
}

// 1. 创建用户
func (repo *UserRepository) Create(user *po.User) error {
    affected, err := createUser(repo, user)
    if err != nil {
        return err
    }
    if affected == 0 {
        return errors.New("创建用户失败")
    }
    return nil
}

// 2. 更新用户信息
func (repo *UserRepository) Update(user *po.User) error {
    affected, err := saveUser(repo, user)
    if err != nil {
        return err
    }
    if affected == 0 {
        return errors.New("更新用户失败")
    }
    return nil
}

// 3. 根据ID查询用户
func (repo *UserRepository) FindByID(id int) (*po.User, error) {
    user, err := findUserByID(repo, id)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// 4. 条件查询
func (repo *UserRepository) FindByEmail(email string) (*po.User, error) {
    user, err := findUserByMap(repo, map[string]interface{}{
        "email": email,
    })
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// 5. 分页查询活跃用户
func (repo *UserRepository) FindActiveUsers(page, pageSize int) ([]*po.User, int, error) {
    pager := NewDescPager("created_at")
    pager.SetPage(page, pageSize)
    
    users, err := findUserListByWhere(repo, "status = ?", []interface{}{1}, pager)
    if err != nil {
        return nil, 0, err
    }
    return users, pager.TotalPage(), nil
}

// 6. 批量查询
func (repo *UserRepository) FindByIDs(ids ...int) (map[int]*po.User, error) {
    users, err := findUserListByID(repo, ids...)
    if err != nil {
        return nil, err
    }
    return UserListToMap(users, nil)
}
```

### 6. Service 层调用示例

```go
// service/user_service.go
type UserService struct {
    Worker freedom.Worker
    UserRepo *repository.UserRepository
}

func NewUserService(worker freedom.Worker) *UserService {
    return &UserService{
        worker: worker,
        userRepo: repository.NewUserRepository(worker),
    }
}

// 创建用户
func (s *UserService) CreateUser(name, email string) error {
    user := &po.User{
        Name:  name,
        Email: email,
    }
    return s.userRepo.Create(user)
}

// 更新用户信息
func (s *UserService) UpdateUserProfile(userID int, newName, newEmail string) error {
    // 1. 先获取 PO
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return fmt.Errorf("查询用户失败: %v", err)
    }
    
    // 2. 使用 Set 方法修改字段
    user.SetName(newName)
    user.SetEmail(newEmail)
    
    // 3. 保存修改
    return s.userRepo.Update(user)
}

// 用户充值示例
func (s *UserService) RechargeBalance(userID int, amount float64) error {
    if amount <= 0 {
        return errors.New("充值金额必须大于0")
    }

    // 1. 获取用户 PO
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return fmt.Errorf("查询用户失败: %v", err)
    }

    // 2. 使用 Add 方法修改数值字段
    user.AddBalance(amount)

    // 3. 保存修改
    return s.userRepo.Update(user)
}

// 更新用户状态示例
func (s *UserService) UpdateUserStatus(userID int, newStatus int) error {
    // 1. 获取用户 PO
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return fmt.Errorf("查询用户失败: %v", err)
    }

    // 2. 使用 Set 方法更新状态
    user.SetStatus(newStatus)
    
    // 3. 可以同时更新多个字段
    user.SetUpdateTime(time.Now())
    user.SetUpdater("system")

    // 4. 保存所有修改
    return s.userRepo.Update(user)
}

// 分页获取用户列表
func (s *UserService) GetActiveUsers(page int) ([]*po.User, int, error) {
    return s.userRepo.FindActiveUsers(page, 10)
}
```

## 生成的代码示例

### PO 结构体示例

```go
type User struct {
    ID        int       `gorm:"primaryKey;column:id"`
    Name      string    `gorm:"column:name"`
    Email     string    `gorm:"column:email"`
    CreatedAt time.Time `gorm:"column:created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at"`
    changes   map[string]interface{}
}

func (obj *User) TableName() string {
    return "users"
}
```

### 仓储方法示例

生成的仓储方法包括：
- `findUser`: 查找单个用户
- `findUserByID`: 根据 ID 查找用户
- `findUserList`: 查找用户列表
- `findUserListByWhere`: 条件查询用户列表
- `createUser`: 创建用户
- `saveUser`: 保存用户更改

## 最佳实践

1. **表结构规范**
   - 建议每个表都包含主键
   - 使用统一的命名规范（建议使用下划线命名）
   - 合理设置字段类型

2. **代码组织**
   - PO 对象应只包含与数据库表结构对应的字段
   - 业务逻辑应在领域层实现，而不是在 PO 层

3. **性能考虑**
   - 合理使用索引
   - 避免生成不必要的字段
   - 适当使用批量操作方法

## 常见问题

1. **生成的字段类型不正确**
   - 检查数据库字段类型是否标准
   - 使用 JSON Schema 时确保类型定义正确

2. **表名前缀处理**
   - 使用 `--prefix` 参数指定要去除的前缀

3. **自定义类型映射**
   - 可以通过修改模板来自定义类型映射规则

## 注意事项

1. 生成代码前请确保已经备份现有代码
2. 确保数据库连接信息正确
3. 建议在开发环境中使用此工具
4. 生成的代码可能需要根据实际需求进行调整

## 更多资源

- [Freedom 框架文档](https://github.com/8treenet/freedom)
- [示例项目](https://github.com/8treenet/freedom/tree/master/example)
- [API 文档](https://godoc.org/github.com/8treenet/freedom) 