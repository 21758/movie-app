# Movie API - 设计文档

## 概述
基于 Go 语言的电影信息管理 RESTful API，提供电影信息存储、用户评分和票房数据集成功能。

## 1. 数据库选型与设计

### 选型理由
**MySQL 8.0** 作为关系型数据库的选择基于：
- **数据一致性需求**：电影、评分等核心业务数据需要严格的事务保证
- **复杂查询支持**：支持多条件过滤、聚合查询等复杂搜索场景
- **JSON 字段支持**：MySQL 8.0+ 原生 JSON 类型适合存储灵活的票房数据
- **成熟生态**：丰富的工具链和运维经验

### 核心表设计

#### Movies 表
```sql
CREATE TABLE movies (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL UNIQUE,
    release_date DATE NOT NULL,
    genre VARCHAR(100) NOT NULL,
    distributor VARCHAR(255),
    budget BIGINT,
    mpa_rating VARCHAR(10),
    box_office_data JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Ratings 表
```sql
CREATE TABLE ratings (
    id VARCHAR(36) PRIMARY KEY,
    movie_title VARCHAR(255) NOT NULL,
    rater_id VARCHAR(255) NOT NULL,
    rating DECIMAL(2,1) NOT NULL CHECK (rating >= 0.5 AND rating <= 5.0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE(movie_title, rater_id)
);
```

### 索引策略
- **主键索引**：UUID 作为主键，保证分布式环境唯一性
- **唯一约束**：电影标题唯一，防止重复创建
- **复合唯一索引**：(movie_title, rater_id) 确保用户对同一电影只能评分一次
- **查询优化索引**：在 title、genre、release_date 上建立索引，优化搜索性能

### 设计亮点
1. **JSON 字段灵活存储**：box_office_data 使用 JSON 类型，适应票房数据结构变化
2. **数据完整性约束**：CHECK 约束确保评分在 0.5-5.0 范围内，步长为 0.5
3. **反范式设计**：ratings 表存储 movie_title 而非 movie_id，减少 JOIN 操作

## 2. 后端服务选型与设计

### 技术栈选型
| 组件 | 选型 | 理由 |
|------|------|------|
| Web 框架 | **Gin** | 高性能、轻量级、灵活中间件 |
| ORM | **GORM** | 功能完善、开发效率高 |
| 配置管理 | **godotenv** | 环境隔离、部署友好 |
| HTTP Client | **net/http** | 标准库、稳定可靠 |

### 架构设计

#### 分层架构
```
HTTP Layer (handlers) → Business Logic → Data Access (database) → External Services
```

#### 核心组件职责

**Config 模块**
- 环境隔离：支持开发、测试、生产环境独立配置
- 安全敏感：认证令牌和 API 密钥通过环境变量管理，避免硬编码

**Database 模块**
- 封装数据库连接和连接池配置
- 实现数据访问层逻辑
- 处理数据序列化/反序列化
- 提供健康检查接口

**Handlers 模块**
- 输入验证：使用 Gin 的绑定机制自动验证请求数据结构
- 业务规则检查：自定义验证逻辑确保数据合规性
- 标准化响应：统一的错误码和消息格式，便于客户端处理

**Middleware 模块**
- 用户身份管理：通过 X-Rater-Id 头标识评分用户，支持匿名评分
- 细粒度权限控制：读写操作差异化认证策略

**Services 模块**
- 服务降级和超时控制
- 数据格式转换

### API 设计原则

#### RESTful 设计
```
GET    /movies           # 搜索电影列表
POST   /movies           # 创建新电影
POST   /movies/:title/ratings   # 提交评分
GET    /movies/:title/rating    # 获取评分统计
GET    /healthz          # 健康检查
```

#### 认证授权策略
- **写操作保护**：创建电影需要 Bearer Token 认证
- **读操作开放**：查询接口无需认证
- **评分身份管理**：通过 X-Rater-Id 头标识评分用户

#### 分页设计
- **游标分页**：使用 offset-based 分页，避免深度分页性能问题
- **Limit 控制**：默认 10 条，支持自定义分页大小

### 错误处理设计
- **统一错误格式**：包含 code 和 message 字段
- **HTTP 状态码映射**：遵循 REST 语义
- **业务错误分类**：BAD_REQUEST、UNAUTHORIZED、NOT_FOUND 等

## 3. 项目优化方向

### 性能优化

#### 缓存策略
```go
// 可以引入 Redis 缓存层，提高性能
type Cache struct {
    redisClient *redis.Client
}

// 缓存热点数据
func (c *Cache) GetMovieByTitle(title string) (*models.Movie, error) {
    // 先查缓存，缓存未命中再查数据库
}
```

**缓存场景**：
- 电影基本信息：TTL 1小时
- 评分聚合结果：TTL 5分钟
- 搜索查询结果：TTL 30分钟

#### 数据库优化
1. **读写分离**：查询操作路由到只读副本
2. **分表策略**：按年份对 movies 表进行水平分表
3. **索引优化**：为复杂查询条件创建复合索引

### 功能完善方向

#### 搜索功能扩展
- **全文搜索**：集成 Elasticsearch 支持复杂搜索
- **模糊匹配**：支持拼音、同义词搜索
- **排序选项**：按评分、票房、时间等多维度排序

#### 用户系统完善
```go
type User struct {
    ID       string `gorm:"primaryKey"`
    Email    string `gorm:"uniqueIndex"`
    Password string // bcrypt 加密存储
    Role     string // admin, user, guest
}
```

#### 数据统计分析
- 电影评分分布分析
- 用户评分行为分析
- 票房趋势统计


#### 高可用设计
- 多实例部署 + 负载均衡
- 数据库主从复制
- 故障自动转移