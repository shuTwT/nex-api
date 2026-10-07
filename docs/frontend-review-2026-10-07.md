# 前端审查报告（2026-10-07）

审查范围：frontend/src 全部页面与组件。方法：将前端全部 113 个 API 调用点与后端路由注册逐一比对；14 个表单的提交 payload 与后端 `DecodeJSON` 目标结构体（`pkg/domain/model/*.go`）逐一比对；支付/会员/个人中心域的响应字段读取与后端响应结构比对。

结论：发现 **2 个 P0（提交必失败 / 页面不可用）、5 个 P1（流程损坏 / 功能缺失）、若干 P2**。其余表单（MCP、接口、兑换码、订阅计划、广告、分类、定时任务、系统设置、兑换、登录）payload 与后端全部匹配。

---

## P0 — 提交必失败 / 页面不可用

### 1. 支付页 operation id 大小写错误（4 处调用点）

- **前端**：`frontend/src/pages/payment.tsx:34`、`62` 调用 `api.payment_outTradeNo_route_get`；另有轮询调用 `api.payment_outTradeNo_status_route_get`（共 3 处 + 1 处）。
- **对照**：`frontend/src/lib/api.ts:83-84` 注册的是全小写 `payment_outtradeno_route_get` / `payment_outtradeno_status_route_get`。
- **原因**：`api` 是 Proxy，任何属性名都返回函数（`lib/api.ts:190-193`），TypeScript 无法拦截；运行时 `invoke` 找不到注册项直接抛 `Unknown API operation`。
- **后果**：充值 / 订阅下单成功后跳转的收银台页**永远显示「加载支付信息失败」**；3 秒轮询在空转抛未捕获异常。
- **建议**：4 处调用点 operation id 改为小写形式。

### 2. 用户表单 credits 字符串化

- **前端**：`frontend/src/components/user-form.tsx:41`：`credits: String(values.credits ?? 1000)`，payload 类型 `Record<string, string>` 迫使数字转字符串。
- **后端**：`pkg/domain/model/accounts.go:11` `UserCreateReq.Credits int`；`accounts.go:17` `UserUpdateReq.Credits *int`。Go `encoding/json` 无法将字符串解到 int 字段。
- **后果**：创建 / 编辑用户**必报 request validation failed**（与令牌 `expiresAt: ""` 同类问题）。
- **建议**：body 类型改为 `Record<string, string | number>`，`credits: values.credits ?? 1000` 直接传数字。

---

## P1 — 流程损坏 / 功能缺失

### 3. 支付页二维码是假图标

- **前端**：`frontend/src/pages/payment.tsx:187`：`qrcodeUrl` 有值时渲染的是 lucide 的 `<QrCode />` 装饰图标，不是真实二维码。
- **后果**：易支付 V2 `pay_type=qrcode` 渠道（或未来直连微信）返回二维码链接时，用户**无法扫码付款**。
- **建议**：引入 `qrcode.react` 等库，用 `payment.qrcodeUrl` 渲染真实二维码。

### 4. 免费计划无法订阅，按钮文案与行为矛盾

- **前端**：`frontend/src/pages/console/membership.tsx:64`：`plan.price === 0` 时报错「免费计划无需支付」；但 L219-220 按钮文案就是「免费订阅」。
- **后端**：存在 `POST /api/membership/subscribe`（`MembershipSubscribeReq{planId}`），前端 `lib/api.ts` 也已注册 `membership_subscribe_route_post`，但**从未被调用**。
- **建议**：免费计划点击后调用 `membership_subscribe_route_post({ planId })` 直接开通。

### 5. 个人中心改资料 / 改密码功能缺失

- **后端**：存在 `PUT /api/personal/profile`（`ProfileUpdateReq{name,email,...}`）与 `PUT /api/personal/profile/password`。
- **前端**：既未在 `lib/api.ts` 注册 operation，也没有任何 UI 入口；用户无法自助修改资料或密码。
- **建议**：在个人中心补两个表单并注册对应 operation。

### 6. 系统设置保存失败无反馈

- **前端**：`frontend/src/pages/console/settings.tsx:40`：`if (result.success) toast.success(...)`，失败分支静默（`result.error` 被吞）。
- **后果**：保存失败时管理员误以为成功。
- **建议**：补 `else toast.error(result.error || "保存失败")`。

### 7. 支付创建接口的支付方式弹窗（次要合并项）

- 订阅创建实际走 `POST /api/payment/methods`（后端 `internal/handler/payment/handlers.go:34` 确认存在），路径与语义不完全直观（methods 既是 GET 查询列表也是 POST 创建订阅），前端调用本身正确，但易混淆。
- **建议**：后端后续可拆出更明确的创建路由，前端同步（非紧急）。

---

## P2 — 体验 / 一致性问题

| # | 问题 | 位置 | 说明 |
|---|------|------|------|
| 8 | 订阅订单显示「自定义计划」 | `payment.tsx:159` | 读 `payment.plan?.title`，但后端 `PaymentResp`（`pkg/domain/model/responses.go:9`）无 plan 字段，永远走兜底文案；可在响应中带 plan 标题或读 metadata |
| 9 | pricing 页硬编码套餐 | `pricing.tsx:8` 起 | 免费版/专业版等写死，与后台「订阅计划」管理完全脱节，后台改价/上下架前台不反映 |
| 10 | 支付方式列可能空白 | `payment.tsx:168-171` | 三个条件渲染无 else，未知 method 显示空白 |
| 11 | lib/api.ts 两条死注册 | `lib/api.ts` | `payment_business_recharge_route_post` / `payment_business_subscription_route_post` 指向后端不存在的 `POST /api/payment/business/*`，且无人调用 |
| 12 | membership 类型标注不严谨 | `console/membership.tsx:27-28` | `startDate/endDate: Date` 实际运行时是 string（`new Date(string)` 恰好能解析，不崩溃） |
| 13 | 令牌无法清除过期时间 | 后端语义限制 | `token_service.go` 更新用 `SetNillableExpiresAt(nil)` = 保持原值；已有过期时间的令牌无法改回「永不过期」，需后端补 Clear 语义 |

---

## 正面确认（无问题）

- **路由全量比对通过**：前端 113 个注册路由与后端路由一致（`gateway` 的 `/api/v1/*`、`upload` 的 `/api/upload`、`cron` 的 `/api/cron/sync-stats` 均真实存在，注册文件名非 `routes.go`/`handlers.go` 需单独确认）。
- **operation id 比对通过**：除上述 payment_outTradeNo 系列外，全部使用点与 `lib/api.ts` 注册表匹配（此前怀疑的驼峰问题仅此一处域）。
- **表单 payload 全部匹配**：MCP 服务、接口管理、兑换码、订阅计划、广告、分类、定时任务、系统初始化、兑换码核销、系统设置的枚举值、指针可空性、ISO 时间转换均正确；兑换码表单的「有值才带 `expiresAt`」写法是正确范式。
- **充值防篡改**：后端 `internal/service/payment/service.go:341` 会重算 `credits = int(amount / creditPrice)` 并与前端值比对，前端无法虚报积分。
- **充值弹窗字段匹配**：`recharge-dialog.tsx` 读取的字段与 `PaymentSettingsResp`（`responses.go:32`）完全一致；易支付启用时后端 `Settings()` 将支付宝/微信同时报为可用（`service.go:303`），两个入口都会展示。
- **登录流程正常**：`providers/auth.tsx` fetch 直调 `/api/auth/login` + CSRF 头，与后端 `auth/routes.go:62-65` 匹配。
- **`/price` 路由问题不存在**：全部跳转均为 `/pricing`，且该路由在 `router.tsx` 中已注册。

---

## 修复进展（2026-10-07 同日）

| # | 状态 | 说明 |
|---|------|------|
| P0-1 | ✅ 已修复 | `payment.tsx` / `payment-result.tsx` / `payment-mock.tsx` 共 4 处 operation id 改为小写 |
| P0-2 | ✅ 已修复 | `user-form.tsx` body 类型改为 `Record<string, string \| number>`，credits 直接传数字 |
| P1-3 | ✅ 已修复 | 引入 `qrcode.react`（`QRCodeSVG`）渲染真实二维码；二维码分支不再限定 wechat，alipay 有 `qrcodeUrl` 时同样渲染 |
| P1-4 | ✅ 已修复 | 免费计划点击后调用 `membership_subscribe_route_post({ planId })` 直接开通并刷新数据 |
| P1-6 | ✅ 已修复 | 系统设置保存失败补 `toast.error(result.error \|\| "设置保存失败")` |
| P2-10 | ✅ 已修复 | 支付方式列补兜底：未知 method 显示原始值 |
| P2-11 | ✅ 已修复 | 删除 `lib/api.ts` 两条死注册 `payment_business_*` |
| P2-12 | ✅ 已修复 | `console/membership.tsx` 的 `startDate/endDate` 类型改为 `string` |
| P1-5 / P2-8 / P2-9 / P2-13 | ⏳ 未处理 | 属「规划」项，涉及后端语义变更或新页面，待后续排期 |
| P1-7 | ✅ 已修复 | 创建订阅支付单拆到新路由 `POST /api/payment/orders`（`payment_orders_route_post`），`/api/payment/methods` 只保留 GET 查列表；测试断言 POST methods 返回 405 |

---

## 建议修复顺序

1. **立即**（一两行改动）：P0-1 operation id 改小写；P0-2 credits 去掉 `String()`。
2. **短期**：P1-3 引入二维码库渲染真实二维码；P1-4 免费计划接入 `membership/subscribe`；P1-6 设置保存失败提示。
3. **规划**：P1-5 个人中心资料/密码表单；P2-9 pricing 页接入后端套餐数据；P2-8 订阅订单展示计划名；其余按需。
