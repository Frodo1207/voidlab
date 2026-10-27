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

- 时间：2026-10-12 15:54:02 CST
- 执行人：TRAE
- 执行位置：本地 / 服务器
- 动作：部署每日资讯“最多 5 篇文章”后端改动到线上 API，并重启 `nginx`
- 命令：
  ```bash
  COPYFILE_DISABLE=1 tar --no-mac-metadata -czf dist-packages/api-build-context-slim.tgz \
    apps/api/Dockerfile \
    apps/api/go.mod apps/api/go.sum \
    apps/api/cmd apps/api/internal

  scp dist-packages/api-build-context-slim.tgz \
    root@142.248.136.161:/opt/voidlab/packages/api-build-context-slim.tgz

  ssh root@142.248.136.161 '
    set -e
    rm -rf /opt/voidlab/build/api-src
    mkdir -p /opt/voidlab/build/api-src
    tar -xzf /opt/voidlab/packages/api-build-context-slim.tgz -C /opt/voidlab/build/api-src
    cd /opt/voidlab/build/api-src
    docker build -t voidlab-api:latest -f apps/api/Dockerfile .
    docker image inspect voidlab-api:latest --format "{{.Id}} {{.Created}} {{.Architecture}}/{{.Os}}"
    docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
      -f /opt/voidlab/deploy/docker-compose.server.yml \
      up -d --force-recreate --no-deps api
    systemctl restart nginx
  '
  ```
- 结果：
  - 新 API 镜像已在服务器本机构建完成：`sha256:815b31a79fac91790cf70646fce1063cb6380e148a2fb26e1c79c9f6ffc3285e`
  - 最新镜像架构已确认：`amd64/linux`
  - `voidlab-api` 已通过 `docker compose ... up -d --force-recreate --no-deps api` 标准替换
  - 公网验证通过：
    - `GET https://voidlab.club/healthz` -> `200`
    - `GET https://voidlab.club/api/v1/public/knowledge/spaces` -> `200`
    - `HEAD https://admin.voidlab.club` -> `200`

### [S4-06] 部署 daily-news reset 接口到线上 API

- 时间：2026-10-12 17:15:11 CST
- 执行人：TRAE
- 执行位置：本地 / 服务器
- 动作：将 `daily-news reset` 后端接口打包上传，并在服务器重建 `voidlab-api` 镜像后替换线上 API 容器
- 命令：
  ```bash
  COPYFILE_DISABLE=1 tar --no-mac-metadata -czf dist-packages/api-build-context-slim.tgz \
    apps/api/Dockerfile \
    apps/api/go.mod apps/api/go.sum \
    apps/api/cmd apps/api/internal

  scp dist-packages/api-build-context-slim.tgz \
    root@142.248.136.161:/opt/voidlab/packages/api-build-context-slim.tgz

  ssh root@142.248.136.161 '
    set -e
    rm -rf /opt/voidlab/build/api-src
    mkdir -p /opt/voidlab/build/api-src
    tar -xzf /opt/voidlab/packages/api-build-context-slim.tgz -C /opt/voidlab/build/api-src
    cd /opt/voidlab/build/api-src
    docker build -t voidlab-api:latest -f apps/api/Dockerfile .
    docker image inspect voidlab-api:latest --format "{{.Id}} {{.Created}} {{.Architecture}}/{{.Os}}"
    docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
      -f /opt/voidlab/deploy/docker-compose.server.yml \
      up -d --force-recreate --no-deps api
    chmod -R 777 /opt/voidlab/data/sqlite /opt/voidlab/data/uploads
    curl -sS http://127.0.0.1:8080/healthz
  '
  ```
- 结果：
  - 新 API 镜像已在服务器本机构建完成：`sha256:9c5eaf0ad390b85944ef118a0070782352e36f53dfb882118969dce968c15286`
  - 最新镜像架构已确认：`amd64/linux`
  - `voidlab-api` 已成功重建并替换
  - 服务器侧健康检查通过：`GET http://127.0.0.1:8080/healthz` -> `200`
- 是否回滚：否
- 备注：本次发布只更新 API，未动官网与后台静态资源

### [OPS-01] 通过正式接口重置并重跑 2026-10-12 每日资讯

- 时间：2026-10-12 17:15:11 CST
- 执行人：TRAE
- 执行位置：本地
- 动作：使用后台登录 token 调用正式 daily-news 接口，先重置当天任务，再采集、补齐至 5 条、审核并发布
- 命令：
  ```bash
  TOKEN=$(curl -sS -X POST https://voidlab.club/api/v1/auth/login \
    -H 'Content-Type: application/json' \
    -d '{"username":"admin","password":"admin"}' \
    | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

  EXTRA=$(curl -sS 'https://aihot.virxact.com/api/v1/items?mode=all&window=24h&limit=10' \
    | node -e 'let s="";process.stdin.on("data",d=>s+=d);process.stdin.on("end",()=>{const j=JSON.parse(s);const seen=new Set(["svvvhr5j2nnywgdw29pso7o49","hvxwaxeyj6i82xv44ef4ce6fz","u9f1ozdbs5nobwugj3yss3z0e","z54sznsd0r2teb6gqyvwu4nc2"]);const item=(j.items||[]).find(x=>x && !seen.has(x.id) && x.links && x.links.aihot && x.links.original && x.title);process.stdout.write(JSON.stringify({items:item?[item]:[]}));});')

  curl -sS -X POST https://voidlab.club/api/v1/daily-news/2026-10-12/reset \
    -H "Authorization: Bearer $TOKEN"

  curl -sS -X POST https://voidlab.club/api/v1/daily-news/collect \
    -H "Authorization: Bearer $TOKEN" \
    -H 'Content-Type: application/json' \
    -d '{"date":"2026-10-12"}'

  curl -sS -X POST https://voidlab.club/api/v1/daily-news/2026-10-12/items \
    -H "Authorization: Bearer $TOKEN" \
    -H 'Content-Type: application/json' \
    -d "$EXTRA"

  curl -sS -X POST https://voidlab.club/api/v1/daily-news/2026-10-12/approve \
    -H "Authorization: Bearer $TOKEN"

  curl -sS -X POST https://voidlab.club/api/v1/daily-news/2026-10-12/publish \
    -H "Authorization: Bearer $TOKEN"
  ```
- 结果：
  - 旧任务已通过 `reset` 正式清理，删除旧文章 `5` 篇
  - 当天精选池原始返回 `4` 条，因此从 AI HOT 全量池补充 `1` 条后重新发布
  - 新文章 ID 为 `18,19,20,21,22`
  - 公网 `GET https://voidlab.club/api/v1/public/articles` 已可读到新的第 `5` 篇文章
- 是否回滚：否
- 备注：本次仍未生成封面图；`cover_url` 为空的问题留待下一阶段“文章 + 图片生成上传”链路处理
- 是否回滚：否
- 备注：
  - 本次只发布 API，未重新上传 `web` 与 `admin` 静态资源
  - 解压源码包时出现 `LIBARCHIVE.xattr.com.apple.provenance` 扩展头警告，但不影响服务器解包与镜像构建

### [S3-04] 本地验证图库版本构建产物

- 时间：2026-10-13 22:42:43 CST
- 执行人：TRAE
- 执行位置：本地
- 动作：对包含图库能力的前后端代码进行本地验证，执行后端测试、前端构建，并生成可发布产物
- 命令：
  ```bash
  go test ./internal/service ./internal/http
  npm run lint
  bash scripts/package-frontends.sh
  COPYFILE_DISABLE=1 tar --no-mac-metadata -czf dist-packages/api-build-context-slim.tgz \
    apps/api/Dockerfile \
    apps/api/go.mod apps/api/go.sum \
    apps/api/cmd apps/api/internal
  ```
- 结果：
  - 后端测试通过：`go test ./internal/service ./internal/http`
  - `apps/admin` 的图库页新增类型错误已修复，前端可正常构建
  - `npm run lint` 仍残留两条历史旧错误：
    - `src/services/dashboardService.ts` 缺少 `dedupeKey`
    - `src/services/mockData.ts` 的 `EventRecord` mock 字段不完整
  - 前端产物已生成：
    - `dist-packages/web-dist.tar.gz`：`3392770` bytes
    - `dist-packages/admin-dist.tar.gz`：`425652` bytes
  - API 服务器构建源码包已生成：
    - `dist-packages/api-build-context-slim.tgz`：`78937` bytes
- 是否回滚：否
- 备注：本机 Docker 在尝试本地构建 `voidlab-api` 镜像时出现卡住现象，因此沿用“本地打前端包 + 上传 API 精简源码包到服务器重建镜像”的稳妥路径

### [S4-07] 部署图库版本到线上环境

- 时间：2026-10-13 22:42:43 CST
- 执行人：TRAE
- 执行位置：本地 / 服务器
- 动作：上传最新前端包和 API 源码包，替换官网与后台静态资源，并在服务器重建 `voidlab-api` 镜像后重启 API 容器
- 命令：
  ```bash
  scp dist-packages/web-dist.tar.gz \
    dist-packages/admin-dist.tar.gz \
    dist-packages/api-build-context-slim.tgz \
    root@142.248.136.161:/opt/voidlab/packages/

  ssh root@142.248.136.161 '
    set -e
    rm -rf /var/www/voidlab/web/* /var/www/voidlab/admin/*
    tar -xzf /opt/voidlab/packages/web-dist.tar.gz -C /var/www/voidlab/web
    tar -xzf /opt/voidlab/packages/admin-dist.tar.gz -C /var/www/voidlab/admin
    rm -rf /opt/voidlab/build/api-src
    mkdir -p /opt/voidlab/build/api-src
    tar -xzf /opt/voidlab/packages/api-build-context-slim.tgz -C /opt/voidlab/build/api-src
    cd /opt/voidlab/build/api-src
    docker build -t voidlab-api:latest -f apps/api/Dockerfile .
    docker image inspect voidlab-api:latest --format "{{.Id}} {{.Created}} {{.Architecture}}/{{.Os}}"
    docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
      -f /opt/voidlab/deploy/docker-compose.server.yml up -d --force-recreate --no-deps api
    chmod -R 777 /opt/voidlab/data/sqlite /opt/voidlab/data/uploads
    sleep 3
    curl -sS http://127.0.0.1:8080/healthz
  '
  ```
- 结果：
  - 官网与后台静态资源已替换为最新版本
  - 服务器已重建最新 API 镜像：`sha256:421ef1a64a88752b5a9d7afee3fa6cba8918a7d22846a568b58d42659443d2e1`
  - 镜像架构确认：`amd64/linux`
  - `voidlab-api` 已通过 `docker compose ... up -d --force-recreate --no-deps api` 完成标准替换
  - 服务器侧健康检查通过：`GET http://127.0.0.1:8080/healthz` -> `200`
- 是否回滚：否
- 备注：解包过程中出现 `LIBARCHIVE.xattr.com.apple.provenance` 扩展头警告，但未影响服务器解包、镜像构建与容器启动

### [S5-03] 验证图库功能已上线

- 时间：2026-10-13 22:42:43 CST
- 执行人：TRAE
- 执行位置：本地 / 服务器
- 动作：验证公网站点、健康检查与后台图库接口，确认图库能力已随本次版本上线
- 命令：
  ```bash
  curl -sS https://voidlab.club/healthz
  curl -I -sS https://voidlab.club
  curl -I -sS https://admin.voidlab.club

  TOKEN=$(curl -sS -X POST https://voidlab.club/api/v1/auth/login \
    -H 'Content-Type: application/json' \
    -d '{"username":"admin","password":"admin"}' \
    | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

  curl -sS https://voidlab.club/api/v1/media/galleries \
    -H "Authorization: Bearer $TOKEN"
  ```
- 结果：
  - `GET https://voidlab.club/healthz` -> `200`
  - `HEAD https://voidlab.club` -> `200`
  - `HEAD https://admin.voidlab.club` -> `200`
  - 图库接口返回成功：`GET /api/v1/media/galleries` -> `200`
  - 线上已存在默认图库：
    - `title=基础图库`
    - `slug=base-gallery`
    - `item_count=80`
- 是否回滚：否
- 备注：说明历史可显示图片已被成功导入基础图库，当前线上已经具备“图库作为稳定封面池”的基础能力

### [S3-05] 本地验证图库自动补封面链路

- 时间：2026-10-13 23:04:13 CST
- 执行人：TRAE
- 执行位置：本地
- 动作：将默认图库封面池接入 `dailynews.Publish()`，并通过单元测试验证“空封面自动补图、已有封面不覆盖”的行为
- 命令：
  ```bash
  go test ./internal/dailynews
  go test ./internal/service ./internal/http
  ```
- 结果：
  - `dailynews.Publish()` 已增加默认图库封面分配逻辑
  - 仅对 `cover_url` 为空的文章补图，已有封面保持不变
  - 当图库相关表不存在时，发布流程会自动跳过补图，不影响旧库兼容
  - 测试通过：
    - `go test ./internal/dailynews`
    - `go test ./internal/service ./internal/http`
- 是否回滚：否
- 备注：当前补图策略为“从默认图库/基础图库按顺序轮转取图”，先保证链路稳定，后续可再升级为更复杂的选图规则

### [S4-08] 部署 daily-news 自动补封面版本到线上 API

- 时间：2026-10-13 23:04:13 CST
- 执行人：TRAE
- 执行位置：本地 / 服务器
- 动作：重新打包 API 精简源码包，上传服务器后重建 `voidlab-api` 镜像并替换线上 API 容器
- 命令：
  ```bash
  COPYFILE_DISABLE=1 tar --no-mac-metadata -czf dist-packages/api-build-context-slim.tgz \
    apps/api/Dockerfile \
    apps/api/go.mod apps/api/go.sum \
    apps/api/cmd apps/api/internal

  scp dist-packages/api-build-context-slim.tgz \
    root@142.248.136.161:/opt/voidlab/packages/api-build-context-slim.tgz

  ssh root@142.248.136.161 '
    set -e
    rm -rf /opt/voidlab/build/api-src
    mkdir -p /opt/voidlab/build/api-src
    tar -xzf /opt/voidlab/packages/api-build-context-slim.tgz -C /opt/voidlab/build/api-src
    cd /opt/voidlab/build/api-src
    docker build -t voidlab-api:latest -f apps/api/Dockerfile .
    docker image inspect voidlab-api:latest --format "{{.Id}} {{.Created}} {{.Architecture}}/{{.Os}}"
    docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
      -f /opt/voidlab/deploy/docker-compose.server.yml up -d --force-recreate --no-deps api
    chmod -R 777 /opt/voidlab/data/sqlite /opt/voidlab/data/uploads
    sleep 3
    curl -sS http://127.0.0.1:8080/healthz
  '
  ```
- 结果：
  - 新 API 镜像已在服务器本机构建完成：`sha256:0326a42272d9135e88d9316d2a3e94fe50c685f23f6ffb2fde8903e8379fbd65`
  - 镜像架构确认：`amd64/linux`
  - `voidlab-api` 已通过 `docker compose ... up -d --force-recreate --no-deps api` 完成替换
  - 服务器侧健康检查通过：`GET http://127.0.0.1:8080/healthz` -> `200`
- 是否回滚：否
- 备注：解包过程中仍出现 `LIBARCHIVE.xattr.com.apple.provenance` 扩展头警告，但未影响服务器构建与重启

### [S5-04] 验证 daily-news 与图库接口发布后正常

- 时间：2026-10-13 23:04:13 CST
- 执行人：TRAE
- 执行位置：本地
- 动作：发布完成后验证健康检查、图库接口和每日资讯后台接口仍可正常访问
- 命令：
  ```bash
  curl -sS https://voidlab.club/healthz

  TOKEN=$(curl -sS -X POST https://voidlab.club/api/v1/auth/login \
    -H 'Content-Type: application/json' \
    -d '{"username":"admin","password":"admin"}' \
    | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

  curl -sS https://voidlab.club/api/v1/media/galleries \
    -H "Authorization: Bearer $TOKEN"

  curl -sS https://voidlab.club/api/v1/daily-news \
    -H "Authorization: Bearer $TOKEN"
  ```
- 结果：
  - `GET https://voidlab.club/healthz` -> `200`
  - 图库接口正常返回，线上仍有默认图库 `base-gallery`，当前 `item_count=80`
  - 每日资讯后台接口正常返回，说明本次 API 发布未破坏 daily-news 主链路
- 是否回滚：否
- 备注：本次线上验收以“接口可用 + 发布流程无回归”为主，自动补封面行为已在本地测试覆盖，未直接改写线上现有已发布文章数据

### [OPS-02] 切换飞书日报目标群到 `voidlab运营`

- 时间：2026-10-15 23:23:59 CST
- 执行人：TRAE
- 执行位置：服务器
- 动作：将线上 `FEISHU_CHAT_ID` 从 `VoidLab灵感中` 切换为 `voidlab运营` 对应群 ID，并重启 API 使环境变量生效
- 命令：
  ```bash
  ssh root@142.248.136.161 '
    set -e
    cp /opt/voidlab/deploy/voidlab.deploy.env /opt/voidlab/deploy/voidlab.deploy.env.bak-$(date +%Y%m%d%H%M%S)
    sed -i "s/^FEISHU_CHAT_ID=.*/FEISHU_CHAT_ID=oc_72cf3401c86eb2026d931d2fed2b6b22/" /opt/voidlab/deploy/voidlab.deploy.env
    docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
      -f /opt/voidlab/deploy/docker-compose.server.yml up -d --no-deps api
    docker exec voidlab-api sh -lc "printenv FEISHU_CHAT_ID"
    curl -sS http://127.0.0.1:8080/healthz
  '
  ```
- 结果：
  - 线上 `FEISHU_CHAT_ID` 已切换为 `oc_72cf3401c86eb2026d931d2fed2b6b22`
  - `voidlab-api` 容器已完成重建并恢复健康
  - 服务器侧健康检查通过：`GET http://127.0.0.1:8080/healthz` -> `200`
- 是否回滚：否
- 备注：此前 `FEISHU_CHAT_ID=oc_79d4c3853b7a26bb4856e05de33a7642` 对应群 `VoidLab灵感中`，飞书原始返回 `230002`，表示机器人不在该群

### [OPS-03] 重发 2026-10-15 每日资讯并验证飞书送达

- 时间：2026-10-15 23:23:59 CST
- 执行人：TRAE
- 执行位置：本地
- 动作：使用后台登录 token 调用正式 `daily-news send` 接口，重发当天日报并验证发送状态
- 命令：
  ```bash
  TOKEN=$(curl -sS -X POST https://voidlab.club/api/v1/auth/login \
    -H 'Content-Type: application/json' \
    -d '{"username":"admin","password":"admin"}' \
    | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

  curl -sS -X POST https://voidlab.club/api/v1/daily-news/2026-10-15/send \
    -H "Authorization: Bearer $TOKEN" \
    -H 'Content-Type: application/json'

  curl -sS https://voidlab.club/api/v1/daily-news \
    -H "Authorization: Bearer $TOKEN"
  ```
- 结果：
  - `POST /api/v1/daily-news/2026-10-15/send` 返回成功
  - 当日日报发送状态已更新为 `sent`
  - `last_error` 已清空
  - 飞书消息 ID：`om_x100b623b4ee9d0a4dd280503ce2cfca`
- 是否回滚：否
- 备注：当前日报已成功投递到 `voidlab运营`

### [WEB-01] 发布 BuilderNetwork 市场详情页与 Header 统一

- 时间：2026-10-22 01:33:00 CST
- 执行人：TRAE
- 执行位置：本地 / 服务器
- 动作：发布官网前端静态资源（market 详情页 + header 全宽一致性相关改动）
- 命令：
  ```bash
  npm run build:web
  mkdir -p dist-packages
  tar -czf dist-packages/web-dist.tar.gz -C apps/web/dist .

  TS=$(date +%Y%m%d-%H%M%S)
  scp dist-packages/web-dist.tar.gz \
    root@142.248.136.161:/opt/voidlab/packages/web-dist-$TS.tar.gz

  ssh root@142.248.136.161 '
    set -euo pipefail
    PKG=/opt/voidlab/packages/web-dist-20261022-012943.tar.gz
    DEST=/var/www/voidlab/web
    TS=$(date +%Y%m%d-%H%M%S)
    mkdir -p /opt/voidlab/backup/web
    mkdir -p "$DEST"
    if [ -n "$(ls -A "$DEST" 2>/dev/null || true)" ]; then
      tar -czf /opt/voidlab/backup/web/web-$TS.tar.gz -C "$DEST" .
    fi
    find "$DEST" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
    tar -xzf "$PKG" -C "$DEST"
    find "$DEST" \( -name "._*" -o -name ".DS_Store" \) -delete
  '

  ssh root@142.248.136.161 'curl -kfsSL https://127.0.0.1/market/projects | head -n 5'
  ssh root@142.248.136.161 'curl -kfsSL https://127.0.0.1/market/projects/agent-workflow-platform-mvp | head -n 5'
  ```
- 结果：
  - 新包已上传：`/opt/voidlab/packages/web-dist-20261022-012943.tar.gz`
  - 官网静态资源已替换到：`/var/www/voidlab/web`
  - 自动备份已生成：`/opt/voidlab/backup/web/web-20261021-173251.tar.gz`
  - 服务器侧验证通过（HTTPS 本地回环，忽略自签证书）：`/market/projects`、`/market/projects/agent-workflow-platform-mvp` 可返回 HTML
- 是否回滚：否
- 备注：本次仅更新 `web` 静态资源，未触发 API / admin 的更新与容器重启

### [WEB-02] 同步热门项目封面静态资源

- 时间：2026-10-22 10:18:00 CST
- 执行人：TRAE
- 执行位置：本地 / 服务器
- 动作：发布官网前端静态资源（新增 `/market/covers/*.jpg`，用于热门项目与项目详情页封面）
- 命令：
  ```bash
  npm run build:web
  TS=$(date +%Y%m%d-%H%M%S)
  tar -czf dist-packages/web-dist-$TS.tar.gz -C apps/web/dist .
  scp dist-packages/web-dist-$TS.tar.gz \
    root@142.248.136.161:/opt/voidlab/packages/web-dist-$TS.tar.gz

  ssh root@142.248.136.161 "
    set -euo pipefail
    PKG=/opt/voidlab/packages/web-dist-20261022-101703.tar.gz
    DEST=/var/www/voidlab/web
    BK_DIR=/opt/voidlab/backup/web
    BK_TS=\$(date +%Y%m%d-%H%M%S)
    mkdir -p \"\$BK_DIR\" \"\$DEST\"
    if [ -n \"\$(ls -A \"\$DEST\" 2>/dev/null || true)\" ]; then
      tar -czf \"\$BK_DIR/web-\$BK_TS.tar.gz\" -C \"\$DEST\" .
    fi
    find \"\$DEST\" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
    tar -xzf \"\$PKG\" -C \"\$DEST\"
    find \"\$DEST\" \( -name '._*' -o -name '.DS_Store' \) -delete
  "

  curl -sSI https://voidlab.club/market/covers/agent-workflow-platform-mvp.jpg | head -n 8
  ```
- 结果：
  - 新包已上传：`/opt/voidlab/packages/web-dist-20261022-101703.tar.gz`
  - 官网静态资源已替换到：`/var/www/voidlab/web`
  - 自动备份已生成：`/opt/voidlab/backup/web/web-20261022-021815.tar.gz`
  - 静态封面可用：`GET https://voidlab.club/market/covers/agent-workflow-platform-mvp.jpg` 返回 `200 image/jpeg`
- 是否回滚：否
- 备注：本次仅更新 `web` 静态资源；tar 解压时出现 `LIBARCHIVE.xattr.com.apple.provenance` 提示但不影响文件落盘

### [DEPLOY-20261023-P3] BuilderNetwork 与 daily-news 发布包预构建

- 时间：2026-10-23 11:45 CST
- 执行人：TRAE
- 执行位置：本地
- 动作：执行发布前测试、前端打包和 API 镜像打包
- 命令：
  ```bash
  (cd apps/api && go test ./...)
  npm run lint:web
  npm run lint:admin
  bash scripts/package-frontends.sh
  bash scripts/package-api-image.sh
  curl --silent --show-error --max-time 5 \
    --unix-socket "$HOME/.docker/run/docker.sock" http://localhost/_ping
  ```
- 结果：
  - Go 全量测试、web/admin 类型检查通过；npm 输出配置兼容性警告。
  - `dist-packages/web-dist.tar.gz`（4.1 MB）和 `dist-packages/admin-dist.tar.gz`（418 KB）构建成功。官网包含 BuilderNetwork Skill 下载文件和 5 张项目封面。
  - 官网与后台 bundle 均有超过 500 KB 的 chunk 警告，不影响构建完成。
  - API 镜像包未生成。本机 Docker socket 的 5 秒健康探测超时，API 镜像构建持续无响应后已停止；线上尚未上传或更新。
  - 原有 `admin-dist.tar.gz` 已备份至 `/tmp/voidlab-admin-dist-before-20261023.tar.gz`，备份哈希与覆盖前文件一致。
- 是否回滚：否（未触碰线上）
- 备注：继续前需要恢复本机 Docker Desktop；当前日报生产开关为关闭状态，runner 不会执行自动采集或分发。

### [DEPLOY-20261023-P3-RETRY] 完成 BuilderNetwork 与 daily-news 发布包构建

- 时间：2026-10-23 17:27 CST
- 执行人：TRAE
- 执行位置：本地
- 动作：重启本机 Docker Desktop，按生产服务器架构构建 API 镜像并导出发布包
- 命令：
  ```bash
  osascript -e 'tell application "Docker" to quit' && open -a Docker
  curl --silent --show-error --max-time 5 \
    --unix-socket "$HOME/.docker/run/docker.sock" http://localhost/_ping
  bash scripts/package-api-image.sh
  docker image inspect voidlab-api:latest --format '{{.Id}} {{.Os}}/{{.Architecture}}'
  ```
- 结果：
  - Docker Desktop 恢复，daemon 返回 `OK`。
  - 本机为 `linux/arm64`，生产服务器为 `x86_64`；已将 `scripts/build-api-image.sh` 设置为显式构建 `linux/amd64`，避免架构不匹配。
  - `dist-packages/web-dist.tar.gz`：4.1 MB，SHA-256 `fa87acc9b5a37674f50984c5c38f31c1f9db3463825b5495ddd734445209fc1f`。
  - `dist-packages/admin-dist.tar.gz`：418 KB，SHA-256 `273df7750820245a98b4c231e279f008a7c402ffe971d372e0ff116b98fb37be`。
  - `dist-packages/voidlab-api.tar`：35 MB，SHA-256 `a292ca63ece74526fe766749fcfd643c6a66d7a486e62df8f8988c1822870122`；镜像为 `linux/amd64`。
  - API 镜像包含 `voidlab-api` 与 `voidlab-dailynews-runner`；日报生产开关关闭且 Agent Token 未配置，runner 保持 idle。
- 是否回滚：否（尚未上传）
- 备注：线上 Nginx 已有 `/market` 精确 SPA 规则，仓库部署配置未包含该规则；本次不覆盖或重载现有 Nginx 配置。

### [DEPLOY-20261023-P4-STAGE] 上传发布包并建立线上回滚点

- 时间：2026-10-23 17:30 CST
- 执行人：TRAE
- 执行位置：本地 / 服务器
- 动作：上传版本化 web、admin、API 镜像和 Compose 包；校验哈希并备份现有线上资源
- 命令：
  ```bash
  scp dist-packages/web-dist.tar.gz root@142.248.136.161:/opt/voidlab/packages/web-dist-20261023-1727.tar.gz
  scp dist-packages/admin-dist.tar.gz root@142.248.136.161:/opt/voidlab/packages/admin-dist-20261023-1727.tar.gz
  scp dist-packages/voidlab-api.tar root@142.248.136.161:/opt/voidlab/packages/voidlab-api-20261023-1727.tar
  scp deploy/docker-compose.server.yml root@142.248.136.161:/opt/voidlab/packages/docker-compose.server-20261023-1727.yml
  ssh root@142.248.136.161 'sha256sum /opt/voidlab/packages/*20261023-1727*'
  ```
- 结果：
  - 三个应用包的服务器端哈希与本地一致；Compose 文件已上传。
  - 线上静态目录已备份至 `/opt/voidlab/backup/web/web-pre-20261023-1730.tar.gz` 和 `/opt/voidlab/backup/admin/admin-pre-20261023-1730.tar.gz`。
  - 原 Compose 文件已备份至 `/opt/voidlab/backup/deploy/docker-compose.server-pre-20261023-1730.yml`。
  - 原 API 镜像已保留回滚 tag `voidlab-api:rollback-20261023-1730`（image ID `sha256:5a02cb4ceac52450e5b214b4e31c4cfa59c053254023e920b4ebdf716209c629`）。
  - 生产环境 `PUBLIC_BASE_URL`、`DAILY_NEWS_SITE_URL` 均为 `https://voidlab.club`；日报和 runner 开关关闭。新 Compose 将增加一个保持 idle 的 runner 服务。
- 是否回滚：否（尚未切换线上服务）
- 备注：生产 Nginx 已有 `/market` 精确 SPA 规则；本次保留现有 Nginx 配置，不上传仓库中缺少该规则的 Nginx 文件。

### [DEPLOY-20261023-P4-API] 更新生产 API 与 Compose

- 时间：2026-10-23 17:32 CST
- 执行人：TRAE
- 执行位置：服务器
- 动作：安装已校验的 Compose 文件，导入 amd64 镜像并更新 API 服务
- 命令：
  ```bash
  cp -p /opt/voidlab/packages/docker-compose.server-20261023-1727.yml \
    /opt/voidlab/deploy/docker-compose.server.yml
  docker load -i /opt/voidlab/packages/voidlab-api-20261023-1727.tar
  docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
    -f /opt/voidlab/deploy/docker-compose.server.yml up -d
  curl -fsS http://127.0.0.1:8080/healthz
  curl -fsS -o /dev/null -w 'status=%{http_code} bytes=%{size_download}\n' \
    http://127.0.0.1:8080/api/v1/public/network/projects
  ```
- 结果：
  - `voidlab-api` 使用新镜像启动，health 返回 `sqlite: ok`；Project 公共 API 返回 `200`。
  - `voidlab-dailynews-runner` 已创建并启动，但由于生产 runner 开关关闭、Token 未设置，日志确认保持 idle。
  - MinIO 保持运行；旧 API 镜像 rollback tag 仍可检查到。
- 是否回滚：否
- 备注：API 启动并通过健康与业务路由检查后，继续发布 web/admin 静态资源。

### [DEPLOY-20261023-P4-P5] 发布静态资源并完成线上验收

- 时间：2026-10-23 17:34 CST
- 执行人：TRAE
- 执行位置：服务器 / 公网
- 动作：发布官网与后台静态资源，校验 Nginx、页面、API、封面和上传资源
- 命令：
  ```bash
  rsync -a --delete --exclude="index.html" "$WEB_STAGE/" /var/www/voidlab/web/
  rsync -a --delay-updates "$WEB_STAGE/index.html" /var/www/voidlab/web/index.html
  rsync -a --delete --exclude="index.html" "$ADMIN_STAGE/" /var/www/voidlab/admin/
  rsync -a --delay-updates "$ADMIN_STAGE/index.html" /var/www/voidlab/admin/index.html
  nginx -t
  systemctl reload nginx
  curl -fsS http://127.0.0.1:8080/healthz
  curl -fsS https://voidlab.club/api/v1/public/network/projects
  ```
- 结果：
  - Nginx 配置检查和 reload 成功；线上 web/admin 的 `index.html` 哈希与本地构建产物一致。
  - `https://voidlab.club/`、`/market`、`/market/projects`、项目详情页、`https://admin.voidlab.club/` 均返回 `200`；web/admin JS 资源均返回 `200`。
  - Project 公共 API 返回 `200`，返回 5 个项目；API 直连 health 返回 `sqlite: ok`。
  - 热门项目封面返回 `200 image/jpeg`；已存在的 `/uploads/20260801033420-d98f6d2a-611b-41c6-83e2-6d838a92dc9a.jpg` 返回 `200 image/jpeg`。
  - `voidlab-api`、`voidlab-dailynews-runner`、MinIO 均处于运行状态；runner 保持 idle。
  - 未认证访问日报管理接口返回 `401`，符合受保护接口行为；本次未使用管理员账号验证登录后后台操作。
- 是否回滚：否
- 备注：tar 解压提示 macOS xattr 扩展头未知，但不影响资源发布。Nginx `/market` 精确 SPA 规则在本次发布前已存在并保留。
