# VOIDLAB HTTP 部署计划

这份计划面向当前阶段的正式上线目标：先使用 `HTTP` 部署 `web + admin + api + minio`，暂不引入 `HTTPS`、自动证书续期和更复杂的编排。

## 目标

- 官网可通过主域名正常访问
- 后台可通过管理域名正常访问
- API 可被官网和后台稳定调用
- 上传文件和媒体资源可通过 `/uploads/` 正常访问
- 每一步部署动作都有记录，便于复盘和后续做成一键化 skill

## 当前已知现状

### 项目侧

- `apps/web`、`apps/admin` 已可成功构建
- `apps/api` 已可成功编译
- 仓库已有 `deploy/docker-compose.yml`
- 仓库已有 `deploy/nginx/voidlab.conf.example`
- 仓库已有前端打包和后端镜像导出脚本

### 服务器侧

- 服务器系统：`CentOS 7`
- 已安装并运行 `nginx`
- 服务器当前缺少 `docker`
- 服务器当前缺少 `docker compose`
- 服务器当前缺少 `node` 与 `go`

这意味着当前最合理的方式不是在服务器源码构建，而是：

- 本地打包前端
- 本地导出后端镜像
- 服务器只负责运行 `nginx + docker + compose`

## 部署拓扑

建议采用单机轻量拓扑：

- `主域名` -> `apps/web` 静态产物
- `管理域名` -> `apps/admin` 静态产物
- `/api/` -> 反代到 `127.0.0.1:8080`
- `/uploads/` -> 反代到 `127.0.0.1:8080/uploads/`
- `api` 与 `minio` 通过 `docker compose` 运行
- `sqlite`、`uploads`、`minio` 数据目录挂载到宿主机

## 目录规划

建议服务器目录如下：

- `/opt/voidlab/`
- `/opt/voidlab/deploy/`
- `/opt/voidlab/packages/`
- `/opt/voidlab/data/sqlite/`
- `/opt/voidlab/data/uploads/`
- `/opt/voidlab/data/minio/`
- `/var/www/voidlab/web/`
- `/var/www/voidlab/admin/`

## 分阶段计划

## 阶段 0：部署前确认

目标：冻结本次上线边界，避免部署过程中临时改方案。

步骤：

1. 确认本次先走 `HTTP`
2. 确认官网域名和后台域名
3. 确认是否沿用当前单机 `sqlite + minio`
4. 确认本次不上 CI/CD，不做自动 HTTPS
5. 确认本次部署以“先可用、后加固”为原则

产出：

- 确认后的域名信息
- 本次部署边界说明

## 阶段 1：服务器运行环境准备

目标：让服务器具备运行当前项目的最小能力。

步骤：

1. 安装 `docker`
2. 安装 `docker compose` 插件
3. 创建部署目录与数据目录
4. 检查 `80` 端口占用情况
5. 备份现有 `nginx` 配置
6. 确认 `nginx` reload 流程可用

验收：

- `docker --version`
- `docker compose version`
- `nginx -t`

记录要求：

- 记录安装命令
- 记录目录创建结果
- 记录任何系统兼容性问题，尤其是 `CentOS 7`

## 阶段 2：仓库部署配置整理

目标：把当前仓库从“能部署”整理到“可重复部署”。

步骤：

1. 确认 `docker-compose` 生产环境变量
2. 将关键变量外置到部署环境文件
3. 检查 `PUBLIC_BASE_URL`
4. 补齐 `nginx` 对 `/uploads/` 的反代
5. 检查后台在管理域名下的 API 访问方式
6. 明确静态目录与容器目录映射

关键变量：

- `PORT`
- `APP_ENV`
- `DB_PATH`
- `UPLOADS_DIR`
- `PUBLIC_BASE_URL`
- `MINIO_ENDPOINT`
- `MINIO_BUCKET`
- `MINIO_ROOT_USER`
- `MINIO_ROOT_PASSWORD`

验收：

- `docker compose config` 可通过
- `nginx` 配置可通过语法检查

## 阶段 3：本地构建与打包

目标：避免在服务器做源码构建。

步骤：

1. 本地构建官网前端
2. 本地构建后台前端
3. 本地导出前端压缩包
4. 本地构建 API 镜像
5. 本地导出 API 镜像包
6. 校验产物完整性

使用现有脚本：

- `bash scripts/package-frontends.sh`
- `bash scripts/package-api-image.sh`

预期产物：

- `dist-packages/web-dist.tar.gz`
- `dist-packages/admin-dist.tar.gz`
- `dist-packages/voidlab-api.tar`

记录要求：

- 记录产物生成时间
- 记录产物文件大小
- 记录本次部署对应的提交或版本说明

## 阶段 4：服务器上传与发布

目标：把打包好的产物发布到服务器。

步骤：

1. 上传前端压缩包到 `/opt/voidlab/packages/`
2. 上传 API 镜像包到 `/opt/voidlab/packages/`
3. 上传部署配置文件到 `/opt/voidlab/deploy/`
4. 解压官网静态资源到 `/var/www/voidlab/web/`
5. 解压后台静态资源到 `/var/www/voidlab/admin/`
6. 导入 API 镜像
7. 执行 `docker compose up -d`
8. 安装并启用新的 `nginx` 站点配置
9. `nginx -t && systemctl reload nginx`

验收：

- 容器正常启动
- 站点首页可访问
- 管理后台可访问
- `/healthz` 返回正常

## 阶段 5：上线验收

目标：确认这次不是“服务起了”，而是真的“能用”。

检查项：

1. 官网首页正常加载
2. 官网能读取公共配置
3. 官网活动、Builder、资讯接口正常
4. 后台登录正常
5. 后台列表页能读取真实数据
6. 上传文件后 `/uploads/` 资源可访问
7. `minio` 服务可用
8. 容器重启后数据仍在

建议额外检查：

- 首页资源是否有 404
- 后台接口是否误打到相对路径
- 代理后请求头是否正确

## 阶段 6：部署后收口

目标：让这次部署可以被重复执行，而不是只成功一次。

步骤：

1. 补全本次部署记录
2. 记录最终生效目录和命令
3. 记录本次遇到的问题与修复方式
4. 固化为 skill 的标准输入、标准输出和标准步骤
5. 为后续 HTTPS 改造预留位置

## 记录规范

后续每次操作都必须同步写入：

- `docs/deployment/http-deployment-log.md`

每一步至少记录：

- 时间
- 执行位置：本地或服务器
- 执行动作
- 实际命令
- 结果
- 是否需要回滚

## 本次推荐执行顺序

建议按下面顺序推进：

1. 先补部署配置文件
2. 再装服务器运行环境
3. 然后本地打包
4. 再上传和发布
5. 最后做验收和记录

## 后续 skill 目标

最终会沉淀成一个可复用的部署 skill，负责：

- 检查服务器是否具备部署条件
- 提醒需要的输入项，比如域名和服务器 IP
- 生成或校验部署产物
- 按顺序执行发布步骤
- 每一步都写入部署日志
- 在结尾输出验收结果和待办项

## 当前边界

这份计划当前只覆盖：

- 单机部署
- `HTTP`
- 静态前端 + Docker API/MinIO
- 手动发布

暂不覆盖：

- `HTTPS`
- 自动续期
- 蓝绿发布
- 多机部署
- 自动化回滚
