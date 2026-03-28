# ArticleServer 使用说明

## 1. 环境依赖

*   **运行时**: Go 1.20+
*   **数据库**: MySQL 8.0+
*   **文档工具**: Swag CLI (用于自动生成 Swagger 契约)
    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    ```

## 2. 快速启动

### 步骤 A：初始化数据库
在 MySQL 中创建逻辑库（表结构由程序启动时自动迁移）：
```sql
CREATE DATABASE IF NOT EXISTS BlogData;
```

### 步骤 B：编译依赖与同步契约
在项目根目录下执行，确保代码注释与生成的文档一致：
```bash
# 整理依赖
go mod tidy

# 生成 Swagger 文档
swag init
```

### 步骤 C：启动服务
```bash
go run main.go
```

## 3. 环境变量配置表 (Environment Variables)

| 变量名 | 说明 | 默认值 |
| :--- | :--- | :--- |
| `DB_USER` | 数据库用户名 | `root` |
| `DB_PASS` | 数据库密码 | `114514` |
| `DB_HOST` | 数据库物理地址 | `127.0.0.1` |
| `DB_PORT` | 数据库监听端口 | `3306` |
| `DB_NAME` | 逻辑数据库名称 | `BlogData` |

## 4. 接口契约说明 (API Contract)

### 文章模块 (Articles)
| 功能 | 方法 | 路径 | 鉴权 | 成功码 |
| :--- | :--- | :--- | :--- | :--- |
| **文章列表** | `GET` | `/api/v1/articles` | 否 | `200` |
| **文章详情** | `GET` | `/api/v1/articles/:id` | 否 | `200` |
| **发布文章** | `POST` | `/api/v1/articles` | 是 | `201` |
| **编辑文章** | `PUT` | `/api/v1/articles/:id` | 是 | `200` |
| **删除文章** | `DELETE` | `/api/v1/articles/:id` | 是 | `204` |

### 资源模块 (Assets)
| 功能 | 方法 | 路径 | 鉴权 | 成功码 |
| :--- | :--- | :--- | :--- | :--- |
| **资源列表** | `GET` | `/api/v1/assets` | 否 | `200` |
| **上传资源** | `POST` | `/api/v1/assets` | 是 | `201` |
| **删除资源** | `DELETE` | `/api/v1/assets` | 是 | `204` |

*注：受限接口需在 Header 中携带 `Authorization: Bearer {token}`。*

## 5. 存储与访问规范

### 物理存储
*   **根路径**: `./storage/assets/`
*   **分卷逻辑**: `/{scope}/{filename}`
*   **自动清理**: 删除接口会同步移除磁盘文件。

### 静态访问
*   **映射规则**: 访问 `http://{host}:8080/assets/` 将映射至 `./storage/assets/`。

## 6. 项目物理结构

```text
├── cmd/
│   └── server/          # 环境调度中心、依赖注入、静态资源路由映射
├── internal/
│   ├── article/         # 文章模块：支持小驼峰契约、UUID标识、Markdown存储
│   ├── asset/           # 资源模块：支持本地/OSS双引擎切换、物理文件管理
│   └── auth/            # 鉴权模块：BearerAuth 中间件逻辑实现
├── storage/             # 物理存储根目录 (运行时自动创建)
├── docs/                # OpenAPI/Swagger 静态文档
├── main.go              # 程序唯一入口
├── go.mod               # 模块依赖描述
└── README.md
```