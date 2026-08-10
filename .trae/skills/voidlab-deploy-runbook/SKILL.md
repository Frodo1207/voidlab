---
name: "voidlab-deploy-runbook"
description: "VOIDLAB 单机 HTTP 部署跑通手册（web/admin 静态资源 + API 容器）。当用户说“部署”“上线”“更新到服务器”或需要排查线上版本时调用。"
---

# VOIDLAB 部署 Runbook（独立版）

本 Skill 只描述 VOIDLAB 当前单机 HTTP 部署流程，不依赖也不引用其它历史 Skill 内容，避免混在一起。

## 适用场景

- 用户说「我更新了代码，帮我部署下」
- 需要把 `apps/web`、`apps/admin`、`apps/api` 的最新改动发布到线上服务器
- 线上出现「接口 404/500」「知识库偶发打不开」需要通过标准重建方式恢复

## 线上基础信息（约定）

- 线上公网地址：`http://142.248.136.161/`
- 后台地址：`http://142.248.136.161:8088/`
- API 基地址：`http://142.248.136.161/api/v1`
- 服务器侧 API 直连（调试）：`http://127.0.0.1:8080`

## 部署总体策略

1. **前端（web/admin）**：本地打包成 `tar.gz`，上传服务器并替换静态目录
2. **后端（api）**：优先采用“服务器端构建镜像 + `docker compose up -d --force-recreate` 标准替换”
3. **验证**：至少验证 `/healthz` 与知识库公共接口

> 注意：避免“直接替换容器内二进制”这类热修方式，风险高且不易回滚。

---

# A. 部署 web/admin（静态资源）

## 1) 本地打包

在项目根目录执行：

```bash
bash scripts/package-frontends.sh
```

预期产物（在 `dist-packages/`）：

- `web-dist.tar.gz`
- `admin-dist.tar.gz`

## 2) 上传并替换服务器静态目录

```bash
scp dist-packages/web-dist.tar.gz dist-packages/admin-dist.tar.gz root@142.248.136.161:/opt/voidlab/packages/

ssh root@142.248.136.161 '
  rm -rf /var/www/voidlab/web/* /var/www/voidlab/admin/* &&
  tar -xzf /opt/voidlab/packages/web-dist.tar.gz -C /var/www/voidlab/web &&
  tar -xzf /opt/voidlab/packages/admin-dist.tar.gz -C /var/www/voidlab/admin
'
```

## 3) 快速验证

- `http://142.248.136.161/` 应可打开
- `http://142.248.136.161:8088/` 应可打开

---

# B. 部署 API（服务器端重建镜像 + force-recreate）

## 0) 前置约束：镜像架构一致

API 镜像必须与服务器 Docker 主机架构一致（通常为 `amd64/linux`）。

API 的 `Dockerfile` 必须使用：

```Dockerfile
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build ...
```

避免 `GOARCH` 写死导致的架构不匹配。

## 1) 本地生成“精简源码包”（避免携带二进制/数据）

在项目根目录执行：

```bash
COPYFILE_DISABLE=1 tar --no-mac-metadata -czf api-build-context-slim.tgz \
  apps/api/Dockerfile \
  apps/api/go.mod apps/api/go.sum \
  apps/api/cmd apps/api/internal
```

## 2) 上传并解压到服务器构建目录

```bash
scp api-build-context-slim.tgz root@142.248.136.161:/opt/voidlab/packages/api-build-context-slim.tgz

ssh root@142.248.136.161 '
  rm -rf /opt/voidlab/build/api-src &&
  mkdir -p /opt/voidlab/build/api-src &&
  tar -xzf /opt/voidlab/packages/api-build-context-slim.tgz -C /opt/voidlab/build/api-src
'
```

## 3) 在服务器后台构建镜像并标准替换

```bash
ssh root@142.248.136.161 '
  nohup bash -lc "
    cd /opt/voidlab/build/api-src &&
    docker build -t voidlab-api:latest -f apps/api/Dockerfile . &&
    docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
      -f /opt/voidlab/deploy/docker-compose.server.yml \
      up -d --force-recreate --no-deps api
  " >/opt/voidlab/build/api-build.log 2>&1 < /dev/null &
'
```

查看构建日志：

```bash
ssh root@142.248.136.161 'tail -n 60 /opt/voidlab/build/api-build.log'
```

## 4) 验证（必须）

至少验证：

- `GET http://142.248.136.161/healthz` -> `200`
- `GET http://142.248.136.161/api/v1/public/knowledge/spaces` -> `200`

如果知识库常用：

- `GET http://142.248.136.161/api/v1/public/knowledge/spaces/agent-builder-playbook/toc` -> `200`

---

# C. 常见问题与处理

## 1) 线上偶发 “failed to get knowledge space”

常见原因：

- SQLite 并发读写瞬时竞争

建议：

- 确认 `sqlite` 初始化已启用 `WAL` 与 `busy_timeout`
- 前端对 `/api/v1/public/knowledge/*` GET 请求可做轻量重试

## 2) SSH 断线导致构建中断

解决：

- 使用 `nohup ... >/opt/voidlab/build/api-build.log &` 后台跑构建，不把构建绑在 SSH 会话上

## 3) tar 解压报 Unexpected EOF

解决：

- 不要传超大包，打“精简源码包”
- 用 `COPYFILE_DISABLE=1` + `--no-mac-metadata` 生成干净 tar 包

---

# D. 什么时候调用这个 Skill（触发词）

- “帮我部署下 / 上线一下 / 更新到服务器”
- “API 404 了 / healthz 不通 / 知识库打不开”
- “重新 build 镜像 / force-recreate 一下”

