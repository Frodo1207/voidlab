---
title: "安装与登录：从 0 到可用"
slug: "workbuddy-install-and-login"
section_name: "入门"
public_summary: "先解决最小可用：下载、安装、登录与版本更新，不在第一步引入多余复杂性。"
cover_url: "http://142.248.136.161/uploads/20260804091630-ec3549d8-ed7c-4a24-80a8-b5a156bc4b08.jpg"
estimated_read_minutes: 7
status: "published"
---

这章只做一件事：把 WorkBuddy 从“没装过”带到“能稳定用起来”。你会完成下载、安装、登录与更新，并顺手把最容易踩坑的点提前避开。

![WorkBuddy 安装与登录](http://142.248.136.161/uploads/20260804091630-ec3549d8-ed7c-4a24-80a8-b5a156bc4b08.jpg)

## 目标

- 确认下载来源可信
- 完成安装并能正常启动
- 登录成功并能进入主界面
- 学会检查更新，保证版本一致

## 一条流程（先照做）

![下载 → 安装 → 启动 → 登录 → 更新](http://142.248.136.161/uploads/20260804091632-2fd887be-9f3b-4bab-8f97-e3c43813b6fd.jpg)

1. 从官方入口下载对应系统版本（Windows / macOS；注意芯片架构）
2. 完成安装并首次启动
3. 选择微信扫码或手机号完成登录
4. 进入个人中心检查更新，保持最新版本

## 下载（只从官方入口）

不要从网盘或不明镜像获取安装包。你要的不是“能装上”，而是“后续可控、可更新、可追责”。

参考：蓝皮书给出的下载入口是 `codebuddy.cn` 的 WorkBuddy 页面。[$TRAE_REF](https://workbuddy.homes/bluebook/%E7%AC%AC%E4%B8%80%E7%AF%87%20%E4%BD%BF%E7%94%A8%E6%89%8B%E5%86%8C%EF%BC%9A%E5%85%88%E6%8A%8A%20WorkBuddy%20%E7%94%A8%E8%B5%B7%E6%9D%A5/%E7%AC%AC%202%20%E7%AB%A0%20WorkBuddy%E7%9A%84%E4%B8%8B%E8%BD%BD%E3%80%81%E5%AE%89%E8%A3%85%E3%80%81%E7%99%BB%E5%BD%95%E4%B8%8E%E6%9B%B4%E6%96%B0/)

![从官方入口下载 WorkBuddy](https://workbuddy.homes/assets/001_image_GGeabJkE2o.DsI_DADN.png)

下载时重点核对两件事：

- 你的系统：Windows / macOS
- 你的架构：Mac 常见是 `ARM64`（Apple Silicon）或 `x64`（Intel）；Windows 常见是 `x64`  

![识别当前设备并选择正确安装包](https://workbuddy.homes/assets/002_image_HaXcbwaJXo.CUrXEWVM.png)

## Windows 安装

1. 双击安装包开始安装
2. 如果系统弹出安全提示：先核对发布者与下载来源，再决定是否继续
3. 按安装向导完成安装并启动

![Windows 中找到下载好的安装包](https://workbuddy.homes/assets/003_image_Ehpebt4Eso.Dw6FFomg.png)

首次启动时，如果你看到“环境准备中”，不用急着重复点击或强制关闭，先等它把运行环境准备完成。

![首次启动：准备运行环境](https://workbuddy.homes/assets/009_image_Q0l3bAkUPo.BiWg74HP.png)

## macOS 安装

1. 打开安装包，将应用拖入「应用程序」
2. 从「应用程序」启动 WorkBuddy

![macOS 中确认下载的是正确的 dmg 包](https://workbuddy.homes/assets/010_image_TmYPbu7Ibo.Ctjv1mqR.png)

![将 WorkBuddy 拖入 Applications](https://workbuddy.homes/assets/011_image_UlJcbVqX7o.ClRSKGHK.png)

![从应用程序中启动 WorkBuddy](https://workbuddy.homes/assets/012_image_LtqPbQ2z6o.CySUnYI7.png)

如果公司电脑禁止安装，不建议绕过终端安全策略；更合理的做法是找 IT 走白名单或企业部署方式。[$TRAE_REF](https://workbuddy.homes/bluebook/%E7%AC%AC%E4%B8%80%E7%AF%87%20%E4%BD%BF%E7%94%A8%E6%89%8B%E5%86%8C%EF%BC%9A%E5%85%88%E6%8A%8A%20WorkBuddy%20%E7%94%A8%E8%B5%B7%E6%9D%A5/%E7%AC%AC%202%20%E7%AB%A0%20WorkBuddy%E7%9A%84%E4%B8%8B%E8%BD%BD%E3%80%81%E5%AE%89%E8%A3%85%E3%80%81%E7%99%BB%E5%BD%95%E4%B8%8E%E6%9B%B4%E6%96%B0/)

## 登录

登录的关键不是选哪种方式，而是确认“登录完成后客户端能正常回到应用”：

1. 点击登录按钮
2. 按提示跳转到网页登录
3. 使用微信扫码或手机号登录
4. 成功后回到客户端，即可开始使用

![客户端登录入口](https://workbuddy.homes/assets/013_image_MMIXbZJafo.B9_LWfGr.png)

![网页登录与授权确认](https://workbuddy.homes/assets/014_image_MdmYbB2Avo.CkBZpKs-.png)

![微信扫码或手机号登录](https://workbuddy.homes/assets/015_image_WZrBbbWono.C1LiVhfP.png)

第一次登录，建议你顺手确认两件事：

- 默认浏览器是否能正常回跳到客户端
- 登录完成后是否真的回到了 WorkBuddy 主界面，而不是停在网页侧

## 更新

保持版本一致是避免“同一句指令、不同电脑结果不一样”的最低成本手段。

一般入口在左下角个人中心，点击「检查更新」即可。[$TRAE_REF](https://workbuddy.homes/bluebook/%E7%AC%AC%E4%B8%80%E7%AF%87%20%E4%BD%BF%E7%94%A8%E6%89%8B%E5%86%8C%EF%BC%9A%E5%85%88%E6%8A%8A%20WorkBuddy%20%E7%94%A8%E8%B5%B7%E6%9D%A5/%E7%AC%AC%202%20%E7%AB%A0%20WorkBuddy%E7%9A%84%E4%B8%8B%E8%BD%BD%E3%80%81%E5%AE%89%E8%A3%85%E3%80%81%E7%99%BB%E5%BD%95%E4%B8%8E%E6%9B%B4%E6%96%B0/)

这里最实用的习惯不是“天天查”，而是：

- 第一次安装完成后查一次
- 看到明显行为差异时查一次
- 团队里多人一起用时，尽量保持版本一致

## 常见问题（先看这里再折腾）

### 安装包打不开或提示损坏

- 删除安装包并从官方入口重新下载
- 核对系统与芯片版本是否匹配
- 保留系统版本、安装包名、报错截图，通过官方渠道反馈  

不要一上来就关闭系统安全机制，那会把后续风险扩大。[$TRAE_REF](https://workbuddy.homes/bluebook/%E7%AC%AC%E4%B8%80%E7%AF%87%20%E4%BD%BF%E7%94%A8%E6%89%8B%E5%86%8C%EF%BC%9A%E5%85%88%E6%8A%8A%20WorkBuddy%20%E7%94%A8%E8%B5%B7%E6%9D%A5/%E7%AC%AC%202%20%E7%AB%A0%20WorkBuddy%E7%9A%84%E4%B8%8B%E8%BD%BD%E3%80%81%E5%AE%89%E8%A3%85%E3%80%81%E7%99%BB%E5%BD%95%E4%B8%8E%E6%9B%B4%E6%96%B0/)

### 登录后没有反应

- 检查默认浏览器是否拦截登录回跳
- 检查网络代理是否影响认证
- 检查系统时间是否准确
- 退出应用后重试，并保留日志与截图  

### 无法读取或写入文件

- 先确认任务选择的工作目录是否正确
- 确认系统是否授予该目录的权限
- 确认文件是否被其他程序锁定  

建议先用一个空白文件做测试，不要拿重要文件反复试错。[$TRAE_REF](https://workbuddy.homes/bluebook/%E7%AC%AC%E4%B8%80%E7%AF%87%20%E4%BD%BF%E7%94%A8%E6%89%8B%E5%86%8C%EF%BC%9A%E5%85%88%E6%8A%8A%20WorkBuddy%20%E7%94%A8%E8%B5%B7%E6%9D%A5/%E7%AC%AC%202%20%E7%AB%A0%20WorkBuddy%E7%9A%84%E4%B8%8B%E8%BD%BD%E3%80%81%E5%AE%89%E8%A3%85%E3%80%81%E7%99%BB%E5%BD%95%E4%B8%8E%E6%9B%B4%E6%96%B0/)

### 更新前要不要备份

更新本身通常不应该改动你的工作文件，但只要你把 WorkBuddy 用进长期项目，就建议把这些东西纳入备份/版本管理：

- 输入资料（原始文件、邮件、会议记录）
- 输出产物（文档/表格/PPT）
- 配置与自定义 Skill

## 这一步的验收

做到以下三点，就算“从 0 到可用”完成：

- 你能进入主界面并发起一次任务
- 你能清楚工作区在哪、默认权限是什么
- 你知道在哪里检查更新
