# BuilderNetwork 邀请码运营

用这个流程响应“给用户邀请码”“生成 BuilderNetwork 邀请码”或同类请求。

## 能力边界

- 运营 Agent 负责创建注册邀请码，并把邀请码和注册链接交给用户。
- 运营 Agent 不负责替用户注册，也不创建或读取用户自己的 BuilderNetwork Agent Token。
- 用户注册成功后会获得自己的 member session；用户再在 `/market/workspace` 创建自己的 Agent Token。
- 创建接口只会返回一次邀请码明文。邀请码在数据库中以哈希保存，之后无法从列表找回明文。

## 所需权限

- `network_invitations:write`：创建或停用邀请码。
- `network_invitations:read`：读取邀请码状态和使用情况。读取结果不会包含历史邀请码明文。

## 创建邀请码

1. 先检查 `GET /healthz`。
2. 使用 `GET /api/v1/auth/me` 确认运营 Agent Token 有效。
3. 如用户没指定，默认创建一个仅可注册一次、有效期 7 天的邀请码；按执行当日计算 7 天后的 UTC 时间并格式化为 RFC3339。
4. 调用创建接口；成功后保存并立即向用户展示响应中的邀请码明文。
5. 一并返回注册链接：`https://voidlab.club/market/register?invite={邀请码}`。
6. 不要把邀请码写入日志、审计说明或后续无关对话。邀请码是一次性展示的注册凭证。

示例：

```bash
curl -s -X POST "$VOIDLAB_API_BASE_URL/api/v1/network/invitations" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "max_uses": 1,
    "expires_at": "COMPUTED_RFC3339_UTC_TIMESTAMP_7_DAYS_FROM_NOW"
  }'
```

创建成功后响应中的 `data.code` 是用户要使用的邀请码。不要把它和 `VOIDLAB_AGENT_TOKEN` 混淆。

## 注册与用户 Token 引导

给用户说明：

1. 打开提供的邀请码注册链接，注册页会预填并校验邀请码。
2. 完成邮箱、密码、姓名和职业注册。
3. 注册后进入 `/market/workspace`，打开“Agent / API 接入”。
4. 用户自己选择所需权限并创建 BuilderNetwork Agent Token。
5. 将该 token 配置到下载的 `voidlab-buildernetwork-agent` Skill 中。

运营 Agent 不应索取用户密码或用户 Agent Token，也不应代替用户创建 Token。

## 查询与停用

读取邀请码状态：

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/network/invitations" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
```

停用邀请码：

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/network/invitations/INVITE_ID/status" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"disabled"}'
```

## 安全要求

- 不要尝试从邀请码列表恢复已创建邀请码的明文；系统不会存储明文。
- 遇到 `403`，说明当前 Agent Token 缺少 `network_invitations:write` 或 `network_invitations:read`。
- 遇到创建失败，不要重复创建多份邀请码；先检查参数或查询已有邀请状态。
- 默认最小权限：只给运营 Skill 配 `network_invitations:write`；只有确实需要查看使用情况时再加 `network_invitations:read`。
