---
title: "API 与登录权限最小闭环"
slug: "fde-0to1-api-auth-minimum"
section_name: "第二部分：基础补全"
public_summary: "把前端和后端真正连起来：接口设计、错误处理、登录与权限。"
estimated_read_minutes: 16
status: "published"
---

## 一句话结论

只要项目进入真实世界，登录、权限和错误处理就一定会出现。

## 这一章要讲什么

- 前后端如何对齐接口：请求/响应结构、错误码与边界
- 登录态怎么维护：Cookie vs Token（至少理解差异与风险）
- 权限如何落地：谁能做什么、前端怎么展示、后端怎么拦

## 交付物

- 一份接口契约：`api-contract.md`
- 一份权限矩阵：`roles-and-permissions.md`

## 验收标准

- 你能说清：前端如何判断登录态、如何处理 401/403、如何安全存 token

## 常见坑

- 只在前端做权限控制，后端没有校验（等同于没权限）
- 登录态处理只考虑“成功”，不考虑过期、刷新、并发与退出
- 报错只弹 toast，没有可定位的信息（请求 id / 状态码 / 关键字段）

## 继续阅读

下一篇：`loop/20-minimal-project-loop.md`
