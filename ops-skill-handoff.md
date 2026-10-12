# VOIDLAB 运营 Skill 交接

把这个目录整个发给同事：

- `./.trae/skills/voidlab-ops-agent/`

如果他还需要看项目里的另一份镜像文档，也有一份：

- `./voidlab-space/project-management/skills/voidlab-ops-agent/`

建议优先用第一份：`./.trae/skills/voidlab-ops-agent/`

## 需要配置的环境变量

```bash
export VOIDLAB_API_BASE_URL="https://admin.voidlab.club"
export VOIDLAB_AGENT_TOKEN="vla_agent_ef44069ab5316164a128f6ea920dee8dfaf5dcf080e0ee19"
```

## 域名

- Admin API / 后台：`https://admin.voidlab.club`
- 官网活动页：`https://voidlab.club`
- 文件上传访问：`https://files.voidlab.club`

## IP 说明

目前项目里能确认的是反向代理和服务都走域名配置，常用入口就是上面这 3 个域名。

仓库里没有一份适合直接交接给同事的“固定公网 IP 清单”，所以实际给同事时，直接给域名即可，不建议交接裸 IP。

## 这个 token 现在能做什么

当前这个 token 是你刚才实际在用的全权限运营 token，可以覆盖这些模块：

- 文章读写
- 活动读写
- Builder 读写
- 报名信息读写
- 知识库读写
- 知识库令牌读写
- 媒体读取/上传

不能做的事：

- 用户管理
- Agent Token 管理
- 站点配置
- 审计日志
- Leads 管理之外的管理员专属操作

## 第一次连通性检查

同事拿到后先跑这三步：

```bash
curl -s "$VOIDLAB_API_BASE_URL/healthz"

curl -s "$VOIDLAB_API_BASE_URL/api/v1/auth/me" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"

curl -s "$VOIDLAB_API_BASE_URL/api/v1/events" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
```

## 我建议你发给同事的话

你把下面这段直接复制给他就行：

```text
运营 skill 目录给你：./.trae/skills/voidlab-ops-agent/

环境变量：
VOIDLAB_API_BASE_URL=https://admin.voidlab.club
VOIDLAB_AGENT_TOKEN=vla_agent_ef44069ab5316164a128f6ea920dee8dfaf5dcf080e0ee19

官网域名：https://voidlab.club
文件域名：https://files.voidlab.club

先用 /healthz、/api/v1/auth/me、/api/v1/events 这三个接口确认联通。
```

## 安全提醒

这枚 token 已经在实际运营里用过。如果你只是临时给同事，可以直接先用。

如果你后面想更稳一点，建议再单独给同事新建一枚 token，用完后把这枚旧 token 轮换掉。
