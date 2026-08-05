# VOIDLAB HTTP 部署记录

这份记录用于沉淀每一次 HTTP 部署的真实操作过程。  
规则很简单：只记录实际发生的动作，不写计划中的动作。

## 使用方式

每做完一个关键步骤，就追加一条记录。  
不要等全部部署完再回忆补写。

建议记录字段：

- 时间
- 执行人
- 执行位置：`本地` / `服务器`
- 步骤编号
- 动作说明
- 实际命令
- 结果
- 后续备注

---

## 当前部署上下文

- 项目：`VOIDLAB`
- 当前策略：`HTTP 单机部署`
- 服务器：已确认可 SSH 访问
- 当前目标：先完成一版可重复的 HTTP 部署流程

---

## 记录模板

### 记录项模板

```md
### [步骤编号] 标题

- 时间：
- 执行人：
- 执行位置：
- 动作：
- 命令：
  ```bash
  # 在这里写实际执行过的命令
  ```
- 结果：
- 是否回滚：
- 备注：
```

---

## 本次会话记录

### [P0-01] 服务器连通性检查

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：本地
- 动作：确认服务器 `22` 端口可达，并验证可通过 SSH 登录
- 命令：
  ```bash
  nc -zv 142.248.136.161 22
  ssh root@142.248.136.161
  ```
- 结果：SSH 端口可达，服务器可登录
- 是否回滚：否
- 备注：本次检查仅用于确认可部署性，不代表已开始正式发布

### [P0-02] 服务器基础环境摸底

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：服务器
- 动作：检查系统版本、CPU、内存、磁盘、Nginx、Docker、Compose 状态
- 命令：
  ```bash
  hostname
  cat /etc/os-release
  uname -a
  nproc
  free -h
  df -h /
  docker --version
  docker compose version
  nginx -v
  ss -ltnp
  ```
- 结果：
  - 系统为 `CentOS 7`
  - 机器规格约为 `2C / 3.7G / 50G`
  - `nginx` 已安装并运行
  - `docker` 与 `docker compose` 尚未安装
- 是否回滚：否
- 备注：后续应走“本地打包，服务器运行”的部署路径

---

## 发布阶段记录区

### [S1-01] 安装 Docker

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：服务器
- 动作：安装 Docker 运行时及其依赖，并启用 `docker` 服务
- 命令：
  ```bash
  yum install -y yum-utils device-mapper-persistent-data lvm2
  yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
  rpm --import https://download.docker.com/linux/centos/gpg
  yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  systemctl enable docker
  systemctl start docker
  docker --version
  ```
- 结果：
  - Docker 官方仓库已添加
  - `docker-ce`、`docker-ce-cli`、`containerd.io` 已安装
  - `docker` 服务已启用并启动
  - 当前版本：`Docker version 26.1.4`
- 是否回滚：否
- 备注：安装过程中首次因 Docker GPG key 未导入而中断，二次执行后完成

### [S1-02] 安装 Docker Compose

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：服务器
- 动作：安装 Docker Compose 插件并验证可用性
- 命令：
  ```bash
  yum install -y docker-compose-plugin
  docker compose version
  ```
- 结果：`docker compose` 可正常执行，当前版本为 `v2.27.1`
- 是否回滚：否
- 备注：Compose 作为 Docker 插件随本次 Docker 安装一并完成

### [S1-03] 创建部署目录

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：服务器
- 动作：初始化部署目录、数据目录、站点目录，并备份当前 `nginx` 配置
- 命令：
  ```bash
  mkdir -p /opt/voidlab/deploy
  mkdir -p /opt/voidlab/packages
  mkdir -p /opt/voidlab/data/sqlite
  mkdir -p /opt/voidlab/data/uploads
  mkdir -p /opt/voidlab/data/minio
  mkdir -p /opt/voidlab/backup
  mkdir -p /var/www/voidlab/web
  mkdir -p /var/www/voidlab/admin
  mkdir -p /opt/voidlab/backup/nginx-pre-voidlab-http-20260801
  cp -a /etc/nginx/nginx.conf /opt/voidlab/backup/nginx-pre-voidlab-http-20260801/
  cp -a /etc/nginx/conf.d /opt/voidlab/backup/nginx-pre-voidlab-http-20260801/
  systemctl is-active docker
  ```
- 结果：
  - `/opt/voidlab` 及其部署、数据、备份目录已创建
  - `/var/www/voidlab/web` 与 `/var/www/voidlab/admin` 已创建
  - `nginx` 配置已备份到 `/opt/voidlab/backup/nginx-pre-voidlab-http-20260801`
  - `docker` 服务状态为 `active`
- 是否回滚：否
- 备注：这一步完成后，服务器已具备进入发布阶段的基础目录结构

### [S2-01] 整理部署配置

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：本地
- 动作：整理当前 HTTP 发布所需配置，补充 IP 直连版 `nginx` 示例，并将示例环境变量改为 `http` 语境
- 命令：
  ```bash
  # 修改 deploy/.env.example
  # 修改 apps/web/.env.production.example
  # 修改 apps/admin/.env.production.example
  # 新增 deploy/nginx/voidlab.http-ip.conf.example
  ```
- 结果：
  - `deploy/.env.example` 已改为 `PUBLIC_BASE_URL=http://YOUR_PUBLIC_HOST`
  - 前端与后台生产环境变量示例已改为 `http://YOUR_PUBLIC_HOST`
  - 已新增 `deploy/nginx/voidlab.http-ip.conf.example`
  - 已新增 `deploy/docker-compose.server.yml`，供服务器直接运行已导出的镜像
  - 当前在没有域名的情况下，可采用：
    - 官网：`80`
    - 后台：`8088`
- 是否回滚：否
- 备注：这一步是为了先完成一版可运行的 HTTP 发布，后续接入域名和 HTTPS 时可切回域名版 Nginx 配置

### [S3-01] 打包官网与后台

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：本地
- 动作：构建官网与后台前端并导出压缩包
- 命令：
  ```bash
  bash scripts/package-frontends.sh
  ```
- 结果：
  - 官网包：`dist-packages/web-dist.tar.gz`
  - 后台包：`dist-packages/admin-dist.tar.gz`
  - 产物大小：
    - `web-dist.tar.gz` 约 `2.9M`
    - `admin-dist.tar.gz` 约 `397K`
- 是否回滚：否
- 备注：构建过程中出现本地 npm 配置告警，但不影响产物生成

### [S3-02] 导出 API 镜像

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：本地
- 动作：构建 API Docker 镜像并导出镜像包
- 命令：
  ```bash
  bash scripts/package-api-image.sh
  ```
- 结果：
  - 镜像包：`dist-packages/voidlab-api.tar`
  - 产物大小约 `27M`
  - 本地 Docker 构建成功，镜像名为 `voidlab-api:latest`
- 是否回滚：否
- 备注：当前镜像按 `linux/amd64` 构建，适用于当前这台 `x86_64` 服务器

### [S4-01] 上传部署产物

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：本地 -> 服务器
- 动作：上传前端包、API 镜像包、部署配置、Nginx 配置和当前 SQLite 数据
- 命令：
  ```bash
  scp dist-packages/web-dist.tar.gz root@142.248.136.161:/opt/voidlab/packages/
  scp dist-packages/admin-dist.tar.gz root@142.248.136.161:/opt/voidlab/packages/
  scp dist-packages/voidlab-api.tar root@142.248.136.161:/opt/voidlab/packages/
  scp deploy/docker-compose.yml root@142.248.136.161:/opt/voidlab/deploy/
  scp deploy/docker-compose.server.yml root@142.248.136.161:/opt/voidlab/deploy/
  scp deploy/nginx/voidlab.http-ip.conf.example root@142.248.136.161:/opt/voidlab/deploy/
  scp data/sqlite/voidlab.db root@142.248.136.161:/opt/voidlab/data/sqlite/
  ```
- 结果：
  - 产物已上传到 `/opt/voidlab/packages/`
  - 部署文件已上传到 `/opt/voidlab/deploy/`
  - 当前 SQLite 数据已上传到 `/opt/voidlab/data/sqlite/voidlab.db`
- 是否回滚：否
- 备注：实际发布时还额外上传了服务器专用 `.env` 和两份实际生效的 Nginx 配置文件

### [S4-02] 导入 API 镜像并启动容器

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：服务器
- 动作：解压静态站点、导入 API 镜像并通过服务器运行版 Compose 启动 `api + minio`
- 命令：
  ```bash
  rm -rf /var/www/voidlab/web/*
  rm -rf /var/www/voidlab/admin/*
  tar -xzf /opt/voidlab/packages/web-dist.tar.gz -C /var/www/voidlab/web
  tar -xzf /opt/voidlab/packages/admin-dist.tar.gz -C /var/www/voidlab/admin
  docker load -i /opt/voidlab/packages/voidlab-api.tar
  docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
    -f /opt/voidlab/deploy/docker-compose.server.yml up -d
  ```
- 结果：
  - `voidlab-api` 与 `voidlab-minio` 容器已启动
  - 首次启动时 `api` 因 SQLite 目录权限不足失败
  - 已执行 `chmod -R 777 /opt/voidlab/data/sqlite /opt/voidlab/data/uploads`
  - 修复后 `http://127.0.0.1:8080/healthz` 返回 `200`
- 是否回滚：否
- 备注：当前这版是为了快速完成 HTTP 首发，后续建议把数据目录权限改成更收敛的用户级授权，而不是继续保留 `777`

### [S4-03] 发布 Nginx 配置

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：服务器
- 动作：替换主 `nginx.conf`，新增后台 `8088` 站点配置，并修正 `/api` 代理规则
- 命令：
  ```bash
  cp /opt/voidlab/deploy/nginx.http-main.conf /etc/nginx/nginx.conf
  cp /opt/voidlab/deploy/nginx.http-admin.conf /etc/nginx/conf.d/voidlab-admin.conf
  nginx -t
  systemctl reload nginx
  ```
- 结果：
  - 官网通过 `80` 端口提供服务
  - 后台通过 `8088` 端口提供服务
  - 初始 Nginx 规则将 `/api/v1/...` 错误转发为 `/v1/...`
  - 已修正 `proxy_pass` 规则，公网 `/api/v1/public/site-configs` 返回 `200`
- 是否回滚：否
- 备注：系统默认 `nginx.conf` 自带 `80` 端口站点，因此本次采用“修改主配置承接官网 + 新增 `8088` 后台配置”的方式，避免 `default_server` 冲突

### [S5-01] 上线验收

- 时间：2026-08-01
- 执行人：TRAE
- 执行位置：本地 / 服务器
- 动作：从服务器内网与公网分别验证官网、后台和公共 API
- 命令：
  ```bash
  curl -I http://127.0.0.1/
  curl -I http://127.0.0.1:8088/
  curl http://127.0.0.1:8080/healthz
  curl http://127.0.0.1/api/v1/public/site-configs
  curl -I http://142.248.136.161/
  curl -I http://142.248.136.161:8088/
  curl http://142.248.136.161/api/v1/public/site-configs
  ```
- 结果：
  - 官网公网地址可访问：`http://142.248.136.161/`
  - 后台公网地址可访问：`http://142.248.136.161:8088/`
  - 公共接口公网可访问：`http://142.248.136.161/api/v1/public/site-configs`
  - 容器健康检查直连可用：`http://127.0.0.1:8080/healthz`
- 是否回滚：否
- 备注：`/healthz` 目前建议先走 `8080` 直连做服务器侧检查；后续接入域名和 HTTPS 时再统一整理成标准健康检查入口

### [S3-03] 重新打包知识库版本前端与 API

- 时间：2026-08-02
- 执行人：TRAE
- 执行位置：本地
- 动作：为知识库模块上线重新构建官网、后台和 API 镜像包
- 命令：
  ```bash
  bash scripts/package-frontends.sh
  bash scripts/package-api-image.sh
  ```
- 结果：
  - 官网与后台静态包重新生成到 `dist-packages/`
  - `voidlab-api.tar` 已按 `linux/amd64` 重新导出
  - 本次镜像已包含知识库公共与受保护路由
- 是否回滚：否
- 备注：构建期间仍有本地 npm 配置告警，但未影响产物生成

### [S4-04] 发布知识库版本到服务器

- 时间：2026-08-02
- 执行人：TRAE
- 执行位置：本地 -> 服务器
- 动作：上传最新官网、后台、API 包，并在服务器替换静态站点与重启 API
- 命令：
  ```bash
  scp dist-packages/web-dist.tar.gz root@142.248.136.161:/opt/voidlab/packages/
  scp dist-packages/admin-dist.tar.gz root@142.248.136.161:/opt/voidlab/packages/
  scp dist-packages/voidlab-api.tar root@142.248.136.161:/opt/voidlab/packages/
  scp deploy/docker-compose.server.yml root@142.248.136.161:/opt/voidlab/deploy/

  ssh root@142.248.136.161 '
    rm -rf /var/www/voidlab/web/* /var/www/voidlab/admin/* &&
    tar -xzf /opt/voidlab/packages/web-dist.tar.gz -C /var/www/voidlab/web &&
    tar -xzf /opt/voidlab/packages/admin-dist.tar.gz -C /var/www/voidlab/admin &&
    docker load -i /opt/voidlab/packages/voidlab-api.tar &&
    docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
      -f /opt/voidlab/deploy/docker-compose.server.yml up -d &&
    chmod -R 777 /opt/voidlab/data/sqlite /opt/voidlab/data/uploads &&
    docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
      -f /opt/voidlab/deploy/docker-compose.server.yml restart api &&
    sleep 3 &&
    curl -s http://127.0.0.1:8080/healthz
  '
  ```
- 结果：
  - 官网与后台静态资源已替换为最新版本
  - `voidlab-api` 镜像已更新并重启成功
  - `http://127.0.0.1:8080/healthz` 返回 `200`
- 是否回滚：否
- 备注：服务已从“不包含知识库接口的旧版本”切换到“包含知识库路由的新版本”

### [S5-02] 验证知识库接口并补充示例内容

- 时间：2026-08-02
- 执行人：TRAE
- 执行位置：本地 / 服务器
- 动作：验证知识库公共接口已上线，并直接写入两组公开知识空间样例数据
- 命令：
  ```bash
  curl http://142.248.136.161/api/v1/public/knowledge/spaces
  curl http://142.248.136.161/api/v1/public/knowledge/spaces/agent-builder-playbook/toc
  curl http://142.248.136.161/api/v1/public/knowledge/spaces/founder-knowledge-os/toc

  scp /Users/vensonheyuhao/.trae/work/6a6d4cf189d2927c4e600a62/seed_knowledge_content.py \
    root@142.248.136.161:/opt/voidlab/packages/seed_knowledge_content.py
  ssh root@142.248.136.161 'python3 /opt/voidlab/packages/seed_knowledge_content.py'

  curl http://142.248.136.161/api/v1/public/knowledge/spaces/agent-builder-playbook/entries/why-agent-workbench-first
  curl http://142.248.136.161/api/v1/public/knowledge/spaces/founder-knowledge-os/entries/why-founders-need-knowledge-os
  ```
- 结果：
  - `/api/v1/public/knowledge/spaces` 已从原来的 `404` 变为正常返回
  - 已新增 2 个公开知识空间：
    - `agent-builder-playbook`
    - `founder-knowledge-os`
  - 已新增 5 篇已发布知识条目，可直接从前台阅读
- 是否回滚：否
- 备注：当前测试 agent token 访问 `/api/v1/knowledge/spaces` 仍为 `403`，说明线上版本已更新，但 token 还没有知识库写权限

### [S4-05] 服务器原生重建 API 正确架构镜像并标准替换

- 时间：2026-08-03
- 执行人：TRAE
- 执行位置：本地 / 服务器
- 动作：修正 API 镜像构建架构策略，上传精简源码包到服务器，使用服务器本机 Docker 重新构建 `amd64/linux` 镜像，并通过 `docker compose up -d --force-recreate` 标准替换 `api`
- 命令：
  ```bash
  # 本地修正 Dockerfile，改为按目标平台构建
  # 仅打包 apps/api 必需源码，不包含本地生成的 server 二进制
  tar -czf api-build-context-slim.tgz \
    apps/api/Dockerfile apps/api/go.mod apps/api/go.sum apps/api/cmd apps/api/internal

  scp api-build-context-slim.tgz root@142.248.136.161:/opt/voidlab/packages/api-build-context-slim.tgz

  ssh root@142.248.136.161 '
    rm -rf /opt/voidlab/build/api-src &&
    mkdir -p /opt/voidlab/build/api-src &&
    tar -xzf /opt/voidlab/packages/api-build-context-slim.tgz -C /opt/voidlab/build/api-src &&
    nohup bash -lc "
      cd /opt/voidlab/build/api-src &&
      docker build -t voidlab-api:latest -f apps/api/Dockerfile . &&
      docker image inspect voidlab-api:latest --format \"{{.Id}} {{.Created}} {{.Architecture}}/{{.Os}}\" &&
      docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
        -f /opt/voidlab/deploy/docker-compose.server.yml up -d --force-recreate --no-deps api
    " >/opt/voidlab/build/api-build.log 2>&1 < /dev/null &
  '
  ```
- 结果：
  - 新 API 镜像已在服务器本机构建完成
  - 最新镜像架构已确认：`amd64/linux`
  - `voidlab-api` 已通过 `docker compose ... up -d --force-recreate --no-deps api` 标准替换
  - 连续 3 轮公网验证通过：
    - `GET /healthz` -> `200`
    - `GET /api/v1/public/knowledge/spaces` -> `200`
    - `GET /api/v1/public/knowledge/spaces/agent-builder-playbook/toc` -> `200`
- 是否回滚：否
- 备注：
  - 本次避免继续使用“容器内直接替换二进制”的高风险路径
  - 服务器上曾存在旧的 `arm64/linux` API 镜像，本次已切换为匹配主机的 `amd64/linux` 新镜像
