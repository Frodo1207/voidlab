# VOIDLAB.AI Production Deployment

这份文档收口当前仓库的上线方式，目标是让 `apps/web`、`apps/admin`、`apps/api` 可以稳定发布到正式环境。

## 当前推荐部署拓扑

- `voidlab.ai`：官网静态站点
- `admin.voidlab.ai`：后台静态站点
- `127.0.0.1:8080`：Gin API 容器
- `127.0.0.1:9000/9001`：MinIO 与控制台
- `nginx`：对外域名入口，负责静态文件与 `/api`、`/uploads` 反代

## 上线前准备

### 1. 域名与目录

- 官网目录：`/var/www/voidlab/web`
- 后台目录：`/var/www/voidlab/admin`
- 仓库目录：`/opt/voidlabai`

### 2. 必备环境

- Node.js `>= 20`
- Docker 与 Docker Compose
- `nginx`

### 3. 部署配置文件

先复制部署环境模板：

```bash
cp deploy/.env.example deploy/.env
```

至少要改这些值：

- `PUBLIC_BASE_URL`
- `MINIO_ROOT_USER`
- `MINIO_ROOT_PASSWORD`

## 前端构建配置

### 官网 `apps/web`

默认同域部署时可以不额外配置。  
如果官网和 API 分域部署，需要在构建前写入：

```bash
cp apps/web/.env.production.example apps/web/.env.production
```

然后把：

```bash
VITE_PUBLIC_API_BASE_URL=https://voidlab.ai
```

改成你的正式 API 域名或网关地址。

### 后台 `apps/admin`

后台强烈建议显式配置 API 地址，避免 `admin` 子域名误打相对路径：

```bash
cp apps/admin/.env.production.example apps/admin/.env.production
```

然后设置：

```bash
VITE_API_BASE_URL=https://voidlab.ai
```

如果你后面把 API 独立成 `api.voidlab.ai`，这里就改成对应域名。

## 标准上线步骤

### 1. 构建前端

```bash
npm run build:web
npm run build:admin
```

生成目录：

- `apps/web/dist`
- `apps/admin/dist`

### 2. 发布静态文件

把构建产物同步到服务器：

```bash
rsync -av --delete apps/web/dist/ /var/www/voidlab/web/
rsync -av --delete apps/admin/dist/ /var/www/voidlab/admin/
```

### 3. 启动或更新 API

首次部署或更新容器：

```bash
docker compose --env-file deploy/.env -f deploy/docker-compose.yml up -d --build
```

### 4. 配置 nginx

参考：

- `deploy/nginx/voidlab.conf.example`

关键点：

- `voidlab.ai` 需要代理 `/api/` 与 `/uploads/`
- `admin.voidlab.ai` 也需要代理 `/api/` 与 `/uploads/`
- 两个静态站点都需要 `try_files ... /index.html`

## 上线验收清单

### API

- `GET /healthz` 返回 `200`
- 管理员能登录后台
- 后台可读取文章、活动、Builder 列表

### 官网

- 首页能加载站点配置
- 活动页能看到真实活动
- Builder 页能看到真实成员
- 资讯页能看到真实文章
- 图片资源不出现 `/uploads` 404

### 后台

- 登录成功后能进入 Dashboard
- 媒体库上传成功
- Agent Token 页面可访问
- 审计日志正常显示

### Agent

- 已创建至少一个有效 Agent Token
- `voidlab-ops-agent` 按 URL + token 可访问后台
- 一次写操作后，审计日志出现 `actor_type=agent`

## 当前上线边界

现在已经具备：

- 真实官网
- 真实后台
- 真实 API
- Agent Token 与 scope 鉴权
- Agent 审计链路

现在还未进入的部分：

- 一键部署脚本
- 自动 SSL 申请与续期编排
- 单命令的 Agent 执行器

所以当前状态可以定义为：

**可以正式上线，也可以开始使用，但运维流程仍是偏轻量的 V1。**
