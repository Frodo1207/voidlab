---
name: "voidlab-http-deploy"
description: "Plans and executes single-server VOIDLAB HTTP deployment. Invoke when deploying or updating web, admin, API, and nginx before HTTPS is introduced."
---

# VOIDLAB HTTP Deploy

这个 skill 用来执行 VOIDLAB 当前阶段的标准部署流程。  
适用范围是：

- 单机服务器
- `HTTP`
- `Nginx + 静态前端 + Docker Compose(api/minio)`
- 手动发布，不包含自动 HTTPS

如果用户说的是以下事情，就应该调用这个 skill：

- “帮我部署到服务器”
- “发布一版官网和后台”
- “更新线上环境”
- “把这套先用 http 跑起来”
- “帮我做服务器上线”

如果用户明确要的是这些内容，则不要优先用这个 skill：

- 纯本地开发
- 只改前端页面，不涉及发布
- 只讨论 HTTPS、证书、CDN、WAF
- 多机编排、Kubernetes、蓝绿发布

## 目标

这个 skill 的目标不是只把服务拉起来，而是把整个部署过程标准化：

- 先检查服务器是否具备部署条件
- 再确认部署输入是否齐全
- 再检查仓库部署配置
- 再构建、上传、发布
- 再做上线验收
- 每一步都写入部署记录

## 标准输入

执行前优先收集这些信息：

- 服务器 IP
- SSH 登录方式
- 官网域名
- 后台域名
- 本次是否只走 `HTTP`
- 本次是否沿用 `sqlite + minio`
- 是否允许在服务器安装 `docker` 与 `docker compose`

如果用户没有一次性提供完，不要一次问很多问题，只补最关键缺口。

## 固定参考文件

执行时优先查看并复用这些文件：

- `docs/deployment/http-deployment-plan.md`
- `docs/deployment/http-deployment-log.md`
- `docs/deployment/production-deployment.md`
- `deploy/docker-compose.yml`
- `deploy/docker-compose.server.yml`
- `deploy/nginx/voidlab.conf.example`
- `deploy/nginx/voidlab.http-ip.conf.example`
- `scripts/package-frontends.sh`
- `scripts/package-api-image.sh`

## 执行原则

### 1. 优先复用现有结构

不要先重做部署体系。  
当前项目已经有：

- 前端构建脚本
- API Dockerfile
- Compose 文件
- Nginx 示例配置

所以默认应采用最小改动方案。

### 2. 服务器不做源码构建

如果服务器缺少 `node` 或 `go`，不要优先在服务器补齐源码构建环境。  
优先使用：

- 本地打包前端
- 本地导出 API 镜像
- 服务器只负责运行

如果服务器只有部署产物，没有完整源码，优先使用：

- `deploy/docker-compose.server.yml`

不要在服务器直接执行带 `build:` 的 compose 文件，以免线上误触发构建。

### 3. 每一步都记日志

关键动作发生后，必须同步更新：

- `docs/deployment/http-deployment-log.md`

至少记录：

- 时间
- 执行位置
- 动作
- 实际命令
- 结果

不要等部署完再补日志。

### 4. 不跳过验收

部署完成不代表上线完成。  
最少要验证：

- 首页访问
- 后台访问
- `/healthz`
- `/api` 请求
- `/uploads` 访问

## 标准阶段

执行时按下面顺序推进。

## 阶段 P0：预检查

目标：确认这次部署是否具备开始条件。

检查项：

1. 项目能否本地构建
2. 服务器是否可 SSH 访问
3. 服务器系统、磁盘、端口、Nginx 状态
4. 服务器是否已安装 `docker`
5. 服务器是否已安装 `docker compose`
6. 当前 `80` 端口是否被 Nginx 占用

输出要求：

- 明确说明“能否开始部署”
- 如果不能开始，指出阻塞项

## 阶段 P1：服务器准备

目标：让服务器具备承载当前项目的能力。

标准动作：

1. 安装 `docker`
2. 安装 `docker compose`
3. 创建目录：
   - `/opt/voidlab/`
   - `/opt/voidlab/deploy/`
   - `/opt/voidlab/packages/`
   - `/opt/voidlab/data/sqlite/`
   - `/opt/voidlab/data/uploads/`
   - `/opt/voidlab/data/minio/`
   - `/var/www/voidlab/web/`
   - `/var/www/voidlab/admin/`
4. 备份现有 `nginx` 配置

输出要求：

- 报告安装结果
- 报告目录创建结果
- 报告任何兼容性风险

## 阶段 P2：部署配置整理

目标：把仓库配置收敛到可上线状态。

标准动作：

1. 检查 `docker-compose` 环境变量
2. 确认 `PUBLIC_BASE_URL`
3. 补齐 `/uploads/` 的 Nginx 反代
4. 确认后台域名下 API 访问方式
5. 确认静态目录和容器挂载目录

输出要求：

- 给出需要修改的配置文件清单
- 给出修改原因

关键规则：

- `nginx` 代理 `/api/` 时，`proxy_pass` 不要写成 `http://127.0.0.1:8080/`
- 优先使用：`proxy_pass http://127.0.0.1:8080`
- 否则 `/api/v1/...` 很容易被错误转成 `/v1/...`

## 阶段 P3：本地构建与打包

目标：在本地生成可发布产物。

标准动作：

1. 执行：
   - `bash scripts/package-frontends.sh`
   - `bash scripts/package-api-image.sh`
2. 检查产物是否存在：
   - `dist-packages/web-dist.tar.gz`
   - `dist-packages/admin-dist.tar.gz`
   - `dist-packages/voidlab-api.tar`

输出要求：

- 报告是否打包成功
- 报告产物路径和大小

## 阶段 P4：上传与发布

目标：把产物落到服务器并切成线上版本。

标准动作：

1. 上传压缩包和部署文件
2. 解压静态资源到：
   - `/var/www/voidlab/web/`
   - `/var/www/voidlab/admin/`
3. 导入 API 镜像
4. 执行 `docker compose up -d`
5. 写入 Nginx 配置
6. 执行：
   - `nginx -t`
   - `systemctl reload nginx`

输出要求：

- 报告发布是否成功
- 报告容器状态
- 报告 Nginx 状态

如果当前没有正式域名，可优先采用：

- 官网：`80`
- 后台：`8088`

也就是：

- `http://SERVER_IP/`
- `http://SERVER_IP:8088/`

## 阶段 P5：上线验收

目标：确认业务真的可用。

验收项：

1. 官网首页访问正常
2. 后台首页访问正常
3. `GET /healthz` 正常
4. 公共 API 可读
5. 后台接口可用
6. `/uploads/` 无 404

输出要求：

- 逐项给出通过或失败
- 如果失败，给出最可能原因

补充规则：

- API 健康检查至少验证一次直连：`http://127.0.0.1:8080/healthz`
- 如果页面能开但 API 不通，先检查 `nginx` 代理规则
- 如果容器启动但 API 不可用，优先检查：
  - SQLite 挂载目录权限
  - 容器日志
  - 实际生效的 `PUBLIC_BASE_URL`

## 推荐命令策略

### 本地构建

```bash
bash scripts/package-frontends.sh
bash scripts/package-api-image.sh
```

### 服务器运行

```bash
docker load -i /opt/voidlab/packages/voidlab-api.tar
docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
  -f /opt/voidlab/deploy/docker-compose.server.yml up -d
nginx -t
systemctl reload nginx
```

如果 API 启动时报 SQLite 只读错误，可先执行：

```bash
chmod -R 777 /opt/voidlab/data/sqlite /opt/voidlab/data/uploads
docker compose --env-file /opt/voidlab/deploy/voidlab.deploy.env \
  -f /opt/voidlab/deploy/docker-compose.server.yml restart api
```

这只是当前 HTTP 首发阶段的快速修复方式，后续应改成更严格的目录所有权方案。

## 记录策略

每完成一个阶段，就更新一次：

- `docs/deployment/http-deployment-log.md`

记录时遵循这几个规则：

1. 只写真实执行过的动作
2. 只写真实执行过的命令
3. 失败也要记录
4. 出现临时绕路时，要写明绕路原因

## 输出格式

执行这个 skill 时，结果输出应保持简洁，优先包含：

- 当前阶段
- 已完成动作
- 产物或状态
- 风险或阻塞项
- 下一步动作

示例：

- 当前阶段：`P3 本地构建与打包`
- 已完成：官网和后台压缩包已生成，API 镜像包已导出
- 产物：`web-dist.tar.gz`、`admin-dist.tar.gz`、`voidlab-api.tar`
- 风险：服务器尚未安装 `docker compose`
- 下一步：进入服务器准备阶段

## 边界

这个 skill 当前不负责：

- 申请 HTTPS 证书
- 配置自动续期
- CDN 接入
- 防火墙策略设计
- 多环境自动发布流水线

如果用户开始讨论这些内容，应单独拆出下一阶段，而不是混进这次 HTTP 部署里。
