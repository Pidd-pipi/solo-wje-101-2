# BrewNotes（咖啡品鉴社区）

面向咖啡爱好者的全栈社区：记录每次咖啡品鉴的详细笔记（风味、评分、冲煮方式），浏览和收藏豆种信息，创建和分享冲煮配方，通过评论与点赞互动，个人主页提供品鉴历史与偏好画像统计。

## Docker Compose 一键启动（推荐）

```bash
cp .env.example .env
docker compose up -d --build
```

访问地址：

- 前端：http://localhost:28601
- 后端 API：http://localhost:29601
- 健康检查：http://localhost:29601/healthz

默认账号：`admin / admin123`（管理员）、`barista / user123`（咖啡爱好者）。

关闭并清理：

```bash
docker compose down -v --remove-orphans
```

## 本地开发（备选）

后端：

```bash
cd backend
go mod tidy
go run ./cmd/server
go build ./...
go test ./...
```

前端：

```bash
cd frontend
npm install
npm run dev
npm run build
```

## 技术栈

| 分层 | 技术 |
| --- | --- |
| 前端 | Vue 3 + TypeScript + Element Plus + ECharts + Vite + Pinia + Vue Router |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | PostgreSQL 16 |
| 认证 | JWT（github.com/golang-jwt/jwt/v5）+ RBAC（user/admin） |
| 其他 | validator/v10、log/slog、Nginx |

## 项目目录结构

```
wje-101/
├── docker-compose.yml
├── .env.example
├── README.md
├── database/init.sql
├── backend/
│   ├── cmd/server/            # main.go + migrate/seed
│   └── internal/
│       ├── config/            # DB/JWT/限流/上传配置
│       ├── model/             # 7 个实体
│       ├── repository/        # 按实体分文件
│       ├── service/           # 按实体分文件
│       ├── handler/           # 按实体分文件 + upload
│       ├── router/            # router.go + 按实体分文件
│       ├── middleware/        # auth/rbac/rate_limiter/error_handler/logger/cors
│       ├── dto/
│       ├── constants/         # note/bean/user/error_codes/log_templates/messages
│       └── util/              # jwt/logger/formatters/file
└── frontend/
    ├── nginx.conf
    └── src/
        ├── api/               # user/note/bean/recipe
        ├── stores/            # useUserStore/useNoteStore/useBeanStore
        ├── components/common/ # ScoreStars/FlavorTags/EmptyState/UserAvatar/ErrorToast/ImageUploader/SearchFilter
        ├── hooks/             # useAuth/usePagination
        ├── pages/             # Home/NoteCreate/NoteDetail/BeanLibrary/Profile/RecipeSquare/Login
        ├── router/            # index.ts（含守卫）
        ├── utils/             # request/storage/dateFormat
        └── constants/         # note/bean/user/errorCodes
```

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | wjecoffeetaste | Compose 项目名/容器前缀 |
| DB_NAME / DB_USER / DB_PASSWORD | coffeetaste / coffeetaste / coffeetaste_pwd | PostgreSQL |
| JWT_SECRET | change_me_to_a_long_random_string | JWT 密钥（生产必改） |
| FRONTEND_PORT / BACKEND_PORT / DB_PORT | 28601 / 29601 / 57601 | 端口映射 |

## Docker 部署说明

- 端口：前端 `28601:80`、后端 `${BACKEND_PORT:-29601}:8080`、DB `${DB_PORT:-57601}:5432`
- 数据卷：`db_data`（PostgreSQL）、`uploads`（笔记配图）
- 依赖顺序：db healthcheck → backend `depends_on: service_healthy` → frontend
- 常见问题：端口冲突改 `.env`；重置数据 `docker compose down -v`；配图通过 `/uploads/` 由 Nginx 代理后端静态目录

## API 接口清单

> 后端统一前缀 `/api/v1`，响应统一为 `{ "code": 0, "message": "ok", "data": ... }`。标注「登录」的接口需携带 `Authorization: Bearer <JWT>`。

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | /healthz | 公开 | 健康检查 |
| POST | /api/v1/uploads | 登录（限流） | 上传图片 |
| POST | /api/v1/users/register | 公开（限流） | 注册并返回 JWT |
| POST | /api/v1/users/login | 公开（限流） | 登录并返回 JWT |
| GET | /api/v1/users/me | 登录 | 获取当前用户 |
| PUT | /api/v1/users/me | 登录 | 更新当前用户资料 |
| GET | /api/v1/users/:id/profile | 公开 | 用户主页（含统计） |
| POST | /api/v1/users/:id/follow | 登录（限流） | 关注用户 |
| DELETE | /api/v1/users/:id/follow | 登录 | 取消关注 |
| GET | /api/v1/notes | 公开 | 品鉴笔记列表/筛选 |
| GET | /api/v1/notes/:id | 公开 | 品鉴笔记详情 |
| POST | /api/v1/notes | 登录（限流） | 发布品鉴笔记 |
| PUT | /api/v1/notes/:id | 登录 | 编辑自己的笔记 |
| DELETE | /api/v1/notes/:id | 登录 | 删除自己的笔记 |
| GET | /api/v1/notes/:id/comments | 公开 | 笔记评论列表 |
| POST | /api/v1/notes/:id/comments | 登录（限流） | 发表评论 |
| DELETE | /api/v1/comments/:id | 登录 | 删除自己的评论 |
| POST | /api/v1/notes/:id/like | 登录（限流） | 点赞笔记 |
| DELETE | /api/v1/notes/:id/like | 登录 | 取消点赞 |
| GET | /api/v1/recipes | 公开 | 冲煮配方列表/筛选 |
| GET | /api/v1/recipes/:id | 公开 | 冲煮配方详情 |
| POST | /api/v1/recipes | 登录（限流） | 分享冲煮配方 |
| GET | /api/v1/beans | 公开 | 咖啡豆库列表/筛选 |
| POST | /api/v1/beans | admin（限流） | 新增咖啡豆 |
| PUT | /api/v1/beans/:id | admin | 更新咖啡豆 |
| DELETE | /api/v1/beans/:id | admin | 删除咖啡豆 |

## 枚举出现位置清单

### RoastLevel（light/medium/dark）

- 后端：`internal/constants/note.go`（定义）、`internal/model/tasting_note.go`（模型）、`internal/service/note_service.go`（校验）、`internal/util/formatters.go`（RoastText）、`internal/constants/log_templates.go`、`database/init.sql`
- 前端：`src/constants/note.ts`（定义）、`src/pages/Home.vue`（筛选器）、`src/pages/NoteCreate.vue`（表单）、`src/pages/NoteDetail.vue`（详情）、`src/pages/Profile.vue`（品鉴历史）

### ProcessMethod（washed/natural/honey/anaerobic）

- 后端：`internal/constants/bean.go`（定义）、`internal/model/coffee_bean.go`（模型）、`internal/service/bean_service.go`（校验）、`internal/util/formatters.go`（ProcessText）、`internal/constants/log_templates.go`、`database/init.sql`
- 前端：`src/constants/bean.ts`（定义）、`src/pages/BeanLibrary.vue`（筛选器+新增表单）

### UserRole（user/admin）

- 后端：`internal/constants/user.go`（定义）、`internal/model/user.go`、`internal/middleware/rbac.go`、`internal/router/beans.go`（管理员路由）、`internal/util/formatters.go`（RoleText）、`database/init.sql`
- 前端：`src/constants/user.ts`（定义）、`src/router/index.ts`（守卫）、`src/pages/BeanLibrary.vue`（管理员按钮显隐）、`src/pages/Profile.vue`（角色标签）

## 横切关注点

- 认证授权（JWT + RBAC）：`middleware/auth.go`、`rbac.go`、`util/jwt.go`、管理员豆种路由、前端路由守卫 + `hooks/useAuth.ts` + `utils/request.ts` 拦截器
- 全局错误处理：`middleware/error_handler.go`、`util/app_error.go`、`constants/error_codes.go`、前端 `utils/request.ts` + `components/common/ErrorToast.vue`
- 请求日志：`middleware/logger.go`（request_id/method/path/status/latency_ms）
- 文件上传：`handler/upload_handler.go`、`util/file.go`、`components/common/ImageUploader.vue`、Nginx `/uploads/` 代理

## License

MIT
