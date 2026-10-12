---
title: "案例二：餐饮预测 Agent（人力与菜品量）"
slug: "fde-0to1-restaurant-forecasting-agent"
section_name: "第四部分：实战案例"
public_summary: "做一个面向餐饮的预测 Agent：预测明日人力和菜品备货量，并给出可解释建议。"
estimated_read_minutes: 18
status: "published"
---

## 一句话结论

预测 Agent 的难点不在“预测模型多高级”，而在“数据是否可信、建议是否可执行、偏差是否能被解释和修正”。

## 这一章要讲什么

- 先把业务闭环跑通：历史数据 → 预测 → 建议 → 次日回填 → 偏差复盘
- 先钉口径，再上模型：你预测的到底是什么、怎么算的
- 把结果做成“可执行建议页”，而不是只做一段对话

![餐饮预测 Agent 业务闭环图](data:image/svg+xml;utf8,%3Csvg%20xmlns%3D%27http%3A//www.w3.org/2000/svg%27%20width%3D%27960%27%20height%3D%27560%27%20viewBox%3D%270%200%20960%20560%27%3E%3Crect%20width%3D%27960%27%20height%3D%27560%27%20fill%3D%27%23f7f7f5%27/%3E%3Crect%20x%3D%2730%27%20y%3D%2730%27%20width%3D%27900%27%20height%3D%27500%27%20rx%3D%2724%27%20fill%3D%27%23ffffff%27%20stroke%3D%27%23e6e6e1%27/%3E%3Ctext%20x%3D%2760%27%20y%3D%2774%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2730%27%20font-weight%3D%27800%27%20fill%3D%27%23111111%27%3E%E9%A4%90%E9%A5%AE%E9%A2%84%E6%B5%8B%20Agent%20%E4%B8%9A%E5%8A%A1%E9%97%AD%E7%8E%AF%3C/text%3E%3Ctext%20x%3D%2760%27%20y%3D%27106%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2715%27%20fill%3D%27%235f5e58%27%3E%E5%8E%86%E5%8F%B2%E6%95%B0%E6%8D%AE%E2%86%92%E6%B8%85%E6%B4%97%E5%8F%A3%E5%BE%84%E2%86%92%E6%98%8E%E6%97%A5%E9%A2%84%E6%B5%8B%E2%86%92%E6%8E%92%E7%8F%AD%E4%B8%8E%E5%A4%87%E8%B4%A7%E5%BB%BA%E8%AE%AE%E2%86%92%E5%AE%9E%E9%99%85%E5%9B%9E%E5%A1%AB%E2%86%92%E5%81%8F%E5%B7%AE%E5%A4%8D%E7%9B%98%E3%80%82%3C/text%3E%3Crect%20x%3D%2760%27%20y%3D%27154%27%20width%3D%27150%27%20height%3D%27118%27%20rx%3D%2718%27%20fill%3D%27%23fbfbfa%27%20stroke%3D%27%23deded8%27/%3E%3Crect%20x%3D%2760%27%20y%3D%27154%27%20width%3D%278%27%20height%3D%27118%27%20fill%3D%27%23c4f000%27/%3E%3Ctext%20x%3D%2784%27%20y%3D%27192%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2720%27%20font-weight%3D%27750%27%20fill%3D%27%23111111%27%3E1.%20%E5%8E%86%E5%8F%B2%E6%95%B0%E6%8D%AE%3C/text%3E%3Ctext%20x%3D%2784%27%20y%3D%27220%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E9%94%80%E9%87%8F%E3%80%81%E5%AE%A2%E6%B5%81%E3%80%81%E5%A4%A9%E6%B0%94%E3%80%81%E6%B4%BB%E5%8A%A8%E3%80%82%3C/text%3E%3Crect%20x%3D%27234%27%20y%3D%27154%27%20width%3D%27150%27%20height%3D%27118%27%20rx%3D%2718%27%20fill%3D%27%23fbfbfa%27%20stroke%3D%27%23deded8%27/%3E%3Crect%20x%3D%27234%27%20y%3D%27154%27%20width%3D%278%27%20height%3D%27118%27%20fill%3D%27%230f7b6c%27/%3E%3Ctext%20x%3D%27258%27%20y%3D%27192%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2720%27%20font-weight%3D%27750%27%20fill%3D%27%23111111%27%3E2.%20%E5%8F%A3%E5%BE%84%E6%B8%85%E6%B4%97%3C/text%3E%3Ctext%20x%3D%27258%27%20y%3D%27220%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E5%8D%95%E5%BA%97%E3%80%81%E6%97%A5%E7%B2%92%E5%BA%A6%E3%80%81%E5%8F%AF%E5%AF%B9%E9%BD%90%E5%AD%97%E6%AE%B5%E3%80%82%3C/text%3E%3Crect%20x%3D%27408%27%20y%3D%27154%27%20width%3D%27150%27%20height%3D%27118%27%20rx%3D%2718%27%20fill%3D%27%23fbfbfa%27%20stroke%3D%27%23deded8%27/%3E%3Crect%20x%3D%27408%27%20y%3D%27154%27%20width%3D%278%27%20height%3D%27118%27%20fill%3D%27%234f46e5%27/%3E%3Ctext%20x%3D%27432%27%20y%3D%27192%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2720%27%20font-weight%3D%27750%27%20fill%3D%27%23111111%27%3E3.%20%E6%98%8E%E6%97%A5%E9%A2%84%E6%B5%8B%3C/text%3E%3Ctext%20x%3D%27432%27%20y%3D%27220%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E4%BA%BA%E5%8A%9B%E3%80%81%E8%8F%9C%E5%93%81%E9%87%8F%E3%80%81%E9%A3%8E%E9%99%A9%E6%B3%A8%E8%AE%B0%E3%80%82%3C/text%3E%3Crect%20x%3D%27582%27%20y%3D%27154%27%20width%3D%27150%27%20height%3D%27118%27%20rx%3D%2718%27%20fill%3D%27%23fbfbfa%27%20stroke%3D%27%23deded8%27/%3E%3Crect%20x%3D%27582%27%20y%3D%27154%27%20width%3D%278%27%20height%3D%27118%27%20fill%3D%27%23dc2626%27/%3E%3Ctext%20x%3D%27606%27%20y%3D%27192%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2720%27%20font-weight%3D%27750%27%20fill%3D%27%23111111%27%3E4.%20%E5%BB%BA%E8%AE%AE%E4%BA%A4%E4%BB%98%3C/text%3E%3Ctext%20x%3D%27606%27%20y%3D%27220%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E6%8E%92%E7%8F%AD%E3%80%81%E5%A4%87%E8%B4%A7%E3%80%81%E5%BC%82%E5%B8%B8%E6%8F%90%E9%86%92%E3%80%82%3C/text%3E%3Crect%20x%3D%27756%27%20y%3D%27154%27%20width%3D%27140%27%20height%3D%27118%27%20rx%3D%2718%27%20fill%3D%27%23fbfbfa%27%20stroke%3D%27%23deded8%27/%3E%3Crect%20x%3D%27756%27%20y%3D%27154%27%20width%3D%278%27%20height%3D%27118%27%20fill%3D%27%236b7280%27/%3E%3Ctext%20x%3D%27780%27%20y%3D%27192%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2720%27%20font-weight%3D%27750%27%20fill%3D%27%23111111%27%3E5.%20%E5%9B%9E%E5%A1%AB%E5%A4%8D%E7%9B%98%3C/text%3E%3Ctext%20x%3D%27780%27%20y%3D%27220%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2714%27%20fill%3D%27%235f5e58%27%3E%E5%AE%9E%E9%99%85%E5%80%BC%E3%80%81%E8%AF%AF%E5%B7%AE%E3%80%81%E5%8E%9F%E5%9B%A0%E3%80%82%3C/text%3E%3Cpath%20d%3D%27M210%20212%20L234%20212%20M384%20212%20L408%20212%20M558%20212%20L582%20212%20M732%20212%20L756%20212%27%20stroke%3D%27%23d2d2cb%27%20stroke-width%3D%274%27%20stroke-linecap%3D%27round%27/%3E%3Cpolygon%20points%3D%27234%2C212%20220%2C204%20220%2C220%27%20fill%3D%27%23d2d2cb%27/%3E%3Cpolygon%20points%3D%27408%2C212%20394%2C204%20394%2C220%27%20fill%3D%27%23d2d2cb%27/%3E%3Cpolygon%20points%3D%27582%2C212%20568%2C204%20568%2C220%27%20fill%3D%27%23d2d2cb%27/%3E%3Cpolygon%20points%3D%27756%2C212%20742%2C204%20742%2C220%27%20fill%3D%27%23d2d2cb%27/%3E%3Crect%20x%3D%2760%27%20y%3D%27326%27%20width%3D%27836%27%20height%3D%27142%27%20rx%3D%2718%27%20fill%3D%27%23fafaf8%27%20stroke%3D%27%23e4e4de%27/%3E%3Ctext%20x%3D%2788%27%20y%3D%27366%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2718%27%20font-weight%3D%27700%27%20fill%3D%27%23111111%27%3E%E7%AC%AC%E4%B8%80%E7%89%88%E5%BB%BA%E8%AE%AE%EF%BC%9A%E5%85%88%E5%81%9A%E5%8D%95%E5%BA%97%20%2B%20%E6%97%A5%E7%B2%92%E5%BA%A6%E3%80%82%3C/text%3E%3Ctext%20x%3D%27348%27%20y%3D%27366%27%20font-family%3D%27Arial%2C%20PingFang%20SC%2C%20Microsoft%20YaHei%27%20font-size%3D%2718%27%20fill%3D%27%235f5e58%27%3E%E5%85%88%E6%8A%8A%E5%8F%A3%E5%BE%84%E3%80%81%E8%A1%A8%E7%BB%93%E6%9E%84%E3%80%81%E6%97%A5%E5%B8%B8%E5%9B%9E%E5%A1%AB%E6%9C%BA%E5%88%B6%E8%B7%91%E9%A1%BA%EF%BC%8C%E5%86%8D%E5%8A%A0%E5%A4%9A%E9%97%A8%E5%BA%97%E5%92%8C%E6%9B%B4%E7%BB%86%E7%B2%92%E5%BA%A6%E3%80%82%3C/text%3E%3C/svg%3E)

## 这个案例要交付什么

- 一个数据录入/导入入口（历史销量、客流、天气、活动等）
- 一个预测结果页（明日人力与菜品量）
- 一份解释（为什么这么预测、关键因子是什么）

## 为什么这个案例对转行特别有价值

它会迫使你把三个能力同时做出来：

- 数据库与数据建模：你必须把业务数据存好、查好、聚合好
- 工程化：你必须能跑定时任务/预测任务，并留下记录
- 产品交付：预测不是论文指标，而是明天要怎么排班、怎么备货

## 第一版为什么先做单店

如果一上来就做多门店、多时段、多品类、多仓库，这个案例会一下子变成一团口径战争。你还没开始写页面和接口，就会先被“同一个字段到底什么意思”拖住。对转行作品来说，这非常危险，因为你会看起来做了很多事，但没有一条链路真正跑通。

所以第一版最稳的做法是：**先做单店、日粒度、次日预测**。这样你先把 3 件事做扎实：

- 数据每天怎么记
- 预测每天怎么跑
- 偏差第二天怎么回填

只要这三件事顺了，后面再扩成多门店、分时段预测，系统会自然长出来。

## 这个 Agent 真正要回答什么问题

这个案例不是做一个“很会分析的聊天机器人”，而是做一个每天都要回答同一类问题的业务系统。它至少要给门店负责人一个可执行答案：

- 明天大概需要多少人
- 明天每个核心菜品大概备多少
- 哪些预测风险最高，应该人工复核

也就是说，这个 Agent 的价值不在“说得多聪明”，而在“明早店长能不能拿它直接安排班次和备货”。

## 先把口径钉死，比先上模型更重要

这一类项目最容易走偏的地方，就是大家急着聊模型：回归、时间序列、特征工程、外部数据、AutoML。但在真实业务里，模型往往不是第一道门，**口径才是第一道门**。

比如“销量”这个词，至少就可能有好几种理解：

- 下单量
- 实付量
- 出餐量
- 核销量
- 退单后净销量

如果这些东西没先说清楚，后面再漂亮的预测结果都不可信。因为你根本不知道你预测的到底是什么。

## 第一版最小数据对象

如果先按单店来做，第一版至少要有这几类数据：

- 门店基础信息
- 菜品基础信息
- 每日营业数据
- 每日菜品销量
- 每日人力排班/实到
- 天气与活动等外部因素
- 每日预测记录
- 每日实际回填记录

你会发现，这里面真正的核心不是“预测表”，而是“历史事实表”。没有稳定的事实表，预测永远只是飘在空中的数字。

## 一个够用的表结构草案

下面这套结构不是唯一答案，但很适合做第一版作品。

### `stores`

记录门店基础信息。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uuid` | 门店主键 |
| `name` | `text` | 门店名称 |
| `city` | `text` | 所在城市 |
| `biz_type` | `text` | 业态，如快餐/正餐/饮品 |
| `status` | `text` | 启用/停用 |

### `menu_items`

记录菜品基础信息。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uuid` | 菜品主键 |
| `store_id` | `uuid` | 所属门店 |
| `name` | `text` | 菜品名称 |
| `category` | `text` | 分类 |
| `unit` | `text` | 单位，如份/杯/份量包 |
| `is_core_item` | `boolean` | 是否核心备货品 |

### `daily_store_metrics`

按“门店 + 日期”存营业总体情况。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uuid` | 主键 |
| `store_id` | `uuid` | 门店 |
| `biz_date` | `date` | 营业日期 |
| `orders_count` | `int` | 订单数 |
| `customers_count` | `int` | 客流或就餐人数 |
| `gmv` | `numeric` | 营业额 |
| `promo_flag` | `boolean` | 是否有营销活动 |
| `holiday_flag` | `boolean` | 是否节假日 |
| `weather_code` | `text` | 天气标签 |
| `temperature_high` | `numeric` | 最高温 |
| `temperature_low` | `numeric` | 最低温 |

### `daily_item_sales`

按“门店 + 日期 + 菜品”存销量。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uuid` | 主键 |
| `store_id` | `uuid` | 门店 |
| `biz_date` | `date` | 营业日期 |
| `menu_item_id` | `uuid` | 菜品 |
| `sold_qty` | `numeric` | 实际销量 |
| `refund_qty` | `numeric` | 退款/退单量 |
| `net_sold_qty` | `numeric` | 净销量 |

### `daily_staffing`

按“门店 + 日期”存排班与实到。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uuid` | 主键 |
| `store_id` | `uuid` | 门店 |
| `biz_date` | `date` | 营业日期 |
| `planned_headcount` | `int` | 计划排班人数 |
| `actual_headcount` | `int` | 实际到岗人数 |
| `labor_hours` | `numeric` | 总工时 |

### `daily_forecasts`

记录 Agent 对“次日”的预测结果。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uuid` | 主键 |
| `store_id` | `uuid` | 门店 |
| `target_date` | `date` | 预测目标日期 |
| `forecast_type` | `text` | `labor` / `item_qty` |
| `target_key` | `text` | 人力时可为空；菜品时填 `menu_item_id` |
| `predicted_value` | `numeric` | 预测值 |
| `confidence_level` | `text` | 高/中/低 |
| `reason_summary` | `text` | 解释摘要 |
| `version` | `text` | 规则或模型版本 |

### `daily_actuals`

记录次日实际值，用于误差复盘。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uuid` | 主键 |
| `store_id` | `uuid` | 门店 |
| `target_date` | `date` | 对应日期 |
| `actual_type` | `text` | `labor` / `item_qty` |
| `target_key` | `text` | 菜品或空 |
| `actual_value` | `numeric` | 实际值 |
| `forecast_id` | `uuid` | 对应预测记录 |
| `abs_error` | `numeric` | 绝对误差 |
| `mape` | `numeric` | 百分比误差 |

## 口径最容易出错的 5 个点

这个案例里，最容易把系统做崩的不是代码，而是下面这些“看起来很像，实际上不是一回事”的字段：

1. `customers_count` 到底是进店人数、下单人数还是核销人数  
2. `sold_qty` 到底算不算退款和赠送  
3. `planned_headcount` 和 `actual_headcount` 是否区分兼职/小时工  
4. `weather_code` 是来自接口原始值还是你自己归一化后的标签  
5. `promo_flag` 是单纯布尔值，还是要进一步区分活动等级

如果这 5 个点你能在第一版里写清楚，后面这个系统就会稳很多。否则你做出来的不是预测系统，而是一个每天都在变定义的报表系统。

## 第一个查询闭环应该怎么设计

先别急着做特别复杂的预测页面。第一版最重要的是先跑通一条查询链：

1. 取近 14 天或 28 天的门店营业数据  
2. 取近 14 天或 28 天的核心菜品销量  
3. 取同周期的人力安排与实际到岗  
4. 合并天气、节假日、活动数据  
5. 输出明日建议值与解释

只要这条链通了，你就已经把“数据库 → 聚合 → 建议 → 页面展示”这条主路打穿了。

## 人力预测和菜品预测为什么要分开看

它们虽然最后都指向“明天怎么准备”，但在业务上不是一回事。

- 人力预测更接近门店总量问题：客流、营业额、节假日、活动强度、营业时长都重要
- 菜品预测更接近品类级问题：单品历史销量、天气、促销、搭配关系、库存风险更重要

所以第一版最好就把它们分成两条输出，而不是强行做成一个总数字。用户最终看到的结果也应该是：

- 明日建议排班人数/工时
- 明日核心菜品备货建议

## 第一版不用先追求“模型炫技”

对转行作品来说，第一版完全可以先用规则 + 简单统计的方式起步，例如：

- 近 7 天均值
- 同星期均值
- 节假日修正
- 天气修正
- 活动修正

只要你的规则写得清楚、解释链条完整、误差每天能回填，这个作品就已经很强了。等第一版闭环稳了，再把更复杂的模型接进去，会更像迭代，而不是赌博。

## 页面应该长什么样

这个案例的页面不应该只是一个输入框加一段对话。它更适合被做成一个“门店经营看板 + Agent 建议页”。

第一版页面建议至少有这 4 块：

- 历史数据概览：最近 7/14/28 天趋势
- 明日预测区：人力与菜品量
- 原因说明区：天气、活动、历史波动、异常提醒
- 偏差复盘区：昨天预测 vs 今天实际

这样用户看到的不是“模型输出了一个数字”，而是“系统给出一个可被理解、可被质疑、可被修正的经营建议”。

## 一个最小 API 设计

第一版完全可以先做下面这些接口：

- `GET /stores/:id/dashboard?days=28`
- `GET /stores/:id/forecast?date=2026-08-08`
- `POST /stores/:id/forecast/run`
- `POST /stores/:id/actuals/fill`
- `GET /stores/:id/forecast/history`

这些接口已经足够支撑作品展示。它们也很适合在面试里讲，因为每一个接口背后都能对应到明确的数据流和业务动作。

## 这个案例怎么验收才算靠谱

这个案例最重要的不是把误差吹得多低，而是让系统真正站住。第一版我会建议你用下面几条来验收：

1. 同一门店的数据口径前后一致  
2. 明日预测页能同时给出人力和核心菜品量  
3. 每个预测结果都能附一段原因摘要  
4. 第二天能回填实际值，并看到偏差  
5. 用户能基于结果直接做排班和备货动作

只要这 5 条成立，这个案例就已经很像真实产品，而不是课堂作业。

## 这个案例在求职里怎么讲

它非常适合拿来讲“我不只会写页面，我能把业务问题做成系统”。因为你可以自然讲出这些关键词：

- 数据建模
- 业务口径
- 聚合查询
- 任务调度
- 可解释预测
- 偏差复盘

这几个词放在一起，会比单纯说“我做了一个 AI 项目”有说服力得多。因为面试官能听出来，你不是只接了个模型接口，而是真的把业务链路想通了。

## 常见坑

- 一上来就追求复杂模型/多门店/分时段，口径没钉死导致数据全乱
- 没有回填机制：预测做完就结束，无法复盘偏差与迭代
- 页面只做“对话”，不做趋势、风险提示、引用数据，结果不可执行

## 继续阅读

下一篇：`_index.md`

如果你准备继续往下把这个案例做实，下一步最值得补的是一份单独的 `data-model.md`，把上面这些表、口径和示例查询整理成正式设计文档。那份文档会是这个案例真正的地基。

## 交付物

- 项目仓库 + 线上地址
- `data-model.md`（数据字段与口径）
- `forecast-notes.md`（模型/规则与误差解释）
- `ops-log.md`（每天预测与实际偏差记录）

## 验收标准

- 预测结果可复盘：你能解释为什么今天会偏高/偏低
- 建议可执行：结果能直接转成排班与备货动作
- 数据口径清楚：不同门店/不同品类的统计口径明确
