# VOIDLAB.AI Workspace

当前仓库已经从单文件原型整理成一个轻量 monorepo，用来先跑通工程、打包和部署流程。

## 目录结构

```text
apps/
  web/      Vue 官网工程
  admin/    Vue 运营后台工程
  api/      Gin 后端
deploy/
  docker-compose.yml
  nginx/voidlab.conf.example
scripts/
  启动、打包、封装脚本
legacy/
  static-prototype/  原始单文件原型归档
data/
  sqlite/
  minio/
```

## 第一次启动

### 1. 安装依赖

```bash
bash scripts/bootstrap.sh
```

### 2. 启动后端容器

```bash
bash scripts/up-services.sh
```

启动后可访问：

- API 健康检查: `http://127.0.0.1:8080/healthz`
- MinIO API: `http://127.0.0.1:9000`
- MinIO Console: `http://127.0.0.1:9001`

默认 MinIO 账号密码：

- `minioadmin`
- `minioadmin`

### 3. 前端本地开发

官网：

```bash
npm run dev:web
```

后台：

```bash
npm run dev:admin
```

### 4. 本地单独跑 Gin

```bash
bash scripts/dev-api.sh
```

## 打包命令

### 打包前端

```bash
bash scripts/package-frontends.sh
```

输出目录：

- `dist-packages/web-dist.tar.gz`
- `dist-packages/admin-dist.tar.gz`

### 打包后端 Docker 镜像

```bash
bash scripts/package-api-image.sh
```

输出目录：

- `dist-packages/voidlab-api.tar`

## 推荐部署方式

完整上线说明见：

- `docs/deployment/production-deployment.md`

### 前端

1. 本地打包得到 `dist`
2. 上传到服务器静态目录
3. 覆盖旧文件
4. `nginx` 不动

### 后端

1. 本地打 Docker 镜像包
2. 上传 `voidlab-api.tar`
3. 服务器执行 `docker load -i voidlab-api.tar`
4. 再执行 `docker compose up -d`

`deploy/nginx/voidlab.conf.example` 提供了一版宿主机 `nginx` 参考配置。

## 生产环境配置模板

- `deploy/.env.example`
- `apps/web/.env.production.example`
- `apps/admin/.env.production.example`
