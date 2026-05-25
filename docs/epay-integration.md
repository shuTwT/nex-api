# 易支付（Z-PAY）对接文档

## 概述

易支付（Z-PAY）是一个第三方支付平台，为商户提供支付宝、微信支付的收款 API 服务。本文档基于 [z-pay.cn](https://z-pay.cn/doc.html) 官方开发文档整理，覆盖完整的支付对接流程。

> **核心接口域名**: `https://zpayz.cn`（也可使用 `https://z-pay.cn`，两者等效）

---

## 目录

1. [页面跳转支付](#1-页面跳转支付)
2. [API 接口支付](#2-api-接口支付)
3. [支付结果通知](#3-支付结果通知)
4. [查询订单](#4-查询订单)
5. [MD5 签名算法](#5-md5-签名算法)
6. [附录：状态码与错误码](#6-附录状态码与错误码)

---

## 1. 页面跳转支付

用户前台直接发起支付，通过 form 表单跳转或拼接 URL 跳转到收银台页面完成付款。

### 1.1 请求 URL

```
https://zpayz.cn/submit.php
```

### 1.2 请求方法

`POST` 或 `GET`（推荐使用 POST，不容易被劫持或屏蔽）

### 1.3 请求参数

| 参数名 | 名称 | 类型 | 必填 | 描述 | 示例 |
|---|---|---|---|---|---|
| `name` | 商品名称 | String | **是** | 需体现出具体售卖的商品，否则容易被封 | `iPhone17苹果手机` |
| `money` | 订单金额 | String | **是** | 最多保留两位小数，单位：元 | `5.67` |
| `type` | 支付方式 | String | **是** | `alipay`：支付宝，`wxpay`：微信支付，`qqpay`：QQ钱包，`tenpay`：财付通 | `alipay` |
| `out_trade_no` | 商户订单号 | String | **是** | 每个订单不可重复，最多 32 位 | `201911914837526544601` |
| `notify_url` | 异步通知地址 | String | **是** | 交易结果回调地址，**不支持带参数** | `https://www.example.com/notify.php` |
| `return_url` | 同步跳转地址 | String | **是** | 交易完成后浏览器跳转地址，**不支持带参数** | `https://www.example.com/return.php` |
| `pid` | 商户唯一标识 | String | **是** | 由平台分配的唯一商户 ID | `201901151314084206659771` |
| `sign` | 签名 | String | **是** | 用于验证信息正确性，MD5 加密 | `28f9583617d9caf66834292b6ab1cc89` |
| `sign_type` | 签名方法 | String | **是** | 固定值 `MD5` | `MD5` |
| `sitename` | 网站名称 | String | 否 | 商户网站名称，用于支付页面展示 | `我的商城` |
| `cid` | 支付渠道 ID | String | 否 | 支持多个，使用 `,` 分隔，不填则随机调用 | `1234` |
| `param` | 附加内容 | String | 否 | 会通过 `notify_url` 原样返回 | `金色 256G` |

### 1.4 用法举例

**GET 请求示例：**

```
https://zpayz.cn/submit.php?name=iPhone17苹果手机&money=599.00&out_trade_no=ORDER20250501001&notify_url=https://www.example.com/notify.php&pid=201901151314084206659771&param=金色 256G&return_url=https://www.example.com/return.php&sign=28f9583617d9caf66834292b6ab1cc89&sign_type=MD5&type=alipay
```

**HTML Form 示例：**

```html
<form action="https://zpayz.cn/submit.php" method="POST">
  <input type="hidden" name="name" value="iPhone17苹果手机">
  <input type="hidden" name="money" value="599.00">
  <input type="hidden" name="type" value="alipay">
  <input type="hidden" name="out_trade_no" value="ORDER20250501001">
  <input type="hidden" name="notify_url" value="https://www.example.com/notify.php">
  <input type="hidden" name="return_url" value="https://www.example.com/return.php">
  <input type="hidden" name="pid" value="201901151314084206659771">
  <input type="hidden" name="sign" value="28f9583617d9caf66834292b6ab1cc89">
  <input type="hidden" name="sign_type" value="MD5">
  <input type="submit" value="去付款">
</form>
```

### 1.5 返回值

| 状态 | 说明 |
|---|---|
| **成功** | 直接跳转到收银台付款页面 |
| **失败** | `{"code":"error","msg":"具体的错误信息"}` |

---

## 2. API 接口支付

服务端至服务端的支付接口，适用于需要在后台创建支付订单的场景。

### 2.1 请求 URL

```
https://zpayz.cn/mapi.php
```

### 2.2 请求方法

`POST`（Content-Type: `application/x-www-form-urlencoded` 或 `multipart/form-data`）

### 2.3 请求参数

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
|---|---|---|---|---|---|
| 商户 ID | `pid` | 是 | String | `20220715225121` | 由平台分配的唯一商户 ID |
| 支付方式 | `type` | 是 | String | `alipay` | `alipay`：支付宝，`wxpay`：微信支付，`qqpay`：QQ钱包，`tenpay`：财付通 |
| 商户订单号 | `out_trade_no` | 是 | String | `20160806151343349` | 每个订单不可重复，最多 32 位 |
| 异步通知地址 | `notify_url` | 是 | String | `https://www.pay.com/notify.php` | 服务器异步通知地址 |
| 商品名称 | `name` | 是 | String | `iPhone17苹果手机` | 需体现出具体售卖的商品 |
| 商品金额 | `money` | 是 | String | `1.00` | 单位：元，最大 2 位小数 |
| 用户 IP 地址 | `clientip` | 是 | String | `192.168.1.100` | 用户发起支付的 IP 地址 |
| 签名字符串 | `sign` | 是 | String | `202cb962ac59075b...` | 签名算法参考[第 5 节](#5-md5-签名算法) |
| 签名类型 | `sign_type` | 是 | String | `MD5` | 固定值 `MD5` |
| 支付渠道 ID | `cid` | 否 | String | `1234` | 多个用 `,` 分隔，不填则随机调用 |
| 设备类型 | `device` | 否 | String | `pc` | 根据用户浏览器 UA 判断，默认为 `pc` |
| 业务扩展参数 | `param` | 否 | String | — | 支付后原样返回 |

### 2.4 返回值（成功时 JSON）

| 字段名 | 变量名 | 类型 | 示例值 | 描述 |
|---|---|---|---|---|
| 返回状态码 | `code` | Int | `1` | `1` 为成功，其他值为失败 |
| 返回信息 | `msg` | String | — | 失败时返回原因 |
| ZPAY 订单号 | `O_id` | String | `123456` | ZPAY 平台订单号 |
| 支付订单号 | `trade_no` | String | `20160806151343349` | 支付订单号 |
| 支付跳转 URL | `payurl` | String | `https://xxx.cn/pay/wxpay/202010903/` | 直接跳转到该 URL 完成支付 |
| 微信 H5 支付 URL | `payurl2` | String | `https://xxx.cn/pay/wxpay/202010903/` | 微信 H5 支付专用链接 |

### 2.5 失败返回值

```json
{
  "code": -1,
  "msg": "签名错误"
}
```

### 2.6 后端调用示例

```typescript
// TypeScript / Node.js 示例
async function createPayment(params: {
  pid: string;
  key: string;
  type: "alipay" | "wxpay";
  outTradeNo: string;
  notifyUrl: string;
  name: string;
  money: string;
  clientip: string;
  param?: string;
}) {
  const { key, ...rest } = params;

  // 生成签名
  const sign = generateSign({ ...rest, key });

  const formData = new URLSearchParams();
  for (const [k, v] of Object.entries(rest)) {
    if (v !== undefined && v !== "") {
      formData.append(k, v);
    }
  }
  formData.append("sign", sign);
  formData.append("sign_type", "MD5");

  const response = await fetch("https://zpayz.cn/mapi.php", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: formData.toString(),
  });

  return response.json();
}
```

---

## 3. 支付结果通知

支付完成后，平台会通过两种方式通知商户支付结果。

### 3.1 通知类型

| 类型 | 说明 |
|---|---|
| **异步通知（`notify_url`）** | 服务器端 POST 通知，商户需返回 `success` 告知平台已收到 |
| **同步跳转（`return_url`）** | 用户浏览器端 GET 跳转，仅作展示用，**不可作为支付成功的唯一依据** |

### 3.2 通知参数

| 参数名 | 名称 | 类型 | 示例值 | 描述 |
|---|---|---|---|---|
| `pid` | 商户 ID | String | `201901151314084206659771` | 平台分配的商户唯一标识 |
| `trade_no` | 平台订单号 | String | `2019011922001418111011411195` | 平台生成的唯一订单号 |
| `out_trade_no` | 商户订单号 | String | `201901191324552185692680` | 商户系统内部的订单号 |
| `type` | 支付方式 | String | `alipay` | `alipay`：支付宝，`wxpay`：微信支付 |
| `name` | 商品名称 | String | `iPhone17苹果手机` | 商品名称（可能被过滤空格） |
| `money` | 订单金额 | String | `599.00` | 实际付款金额（可能跟下单金额不一致） |
| `param` | 附加内容 | String | `金色 256G` | 下单时传入的附加参数 |
| `trade_status` | 支付状态 | String | `TRADE_SUCCESS` | 只有 `TRADE_SUCCESS` 表示支付成功 |
| `sign` | 签名 | String | `ef6e3c5c6ff...` | 用于验证回调真实性 |
| `sign_type` | 签名类型 | String | `MD5` | 固定值 `MD5` |

### 3.3 异步通知处理流程

```
用户支付完成
    │
    ▼
平台 POST 通知 → notify_url
    │
    ▼
商户服务器接收参数
    │
    ├── 验证签名（必须）
    ├── 检查 trade_status === "TRADE_SUCCESS"
    ├── 检查订单金额是否一致
    ├── 处理业务逻辑（更新订单状态等）
    │
    ▼
返回字符串 "success"
```

### 3.4 通知验签示例

```typescript
/**
 * 验证异步通知签名
 * @param params  通知中的所有参数
 * @param key     商户密钥
 * @returns       验签是否通过
 */
function verifyNotifySign(params: Record<string, string>, key: string): boolean {
  const { sign, sign_type, ...rest } = params;

  // 按参数名 ASCII 排序
  const sortedKeys = Object.keys(rest)
    .filter((k) => rest[k] !== "" && rest[k] !== undefined)
    .sort();

  // 拼接字符串
  const signStr = sortedKeys
    .map((k) => `${k}=${rest[k]}`)
    .join("&");

  // MD5(signStr + key)，结果转为小写
  const expectedSign = md5(signStr + key).toLowerCase();

  return expectedSign === sign;
}
```

### 3.5 通知处理示例（Next.js Route Handler）

```typescript
// app/api/payment/notify/route.ts
import { NextRequest, NextResponse } from "next/server";
import crypto from "crypto";

export async function POST(request: NextRequest) {
  const body = await request.text();
  const params = Object.fromEntries(new URLSearchParams(body));
  const key = process.env.EPAY_KEY!;

  // 1. 验证签名
  if (!verifyNotifySign(params, key)) {
    return new NextResponse("sign error", { status: 400 });
  }

  // 2. 检查支付状态
  if (params.trade_status !== "TRADE_SUCCESS") {
    return new NextResponse("success");
  }

  // 3. 处理业务逻辑
  const { out_trade_no, money, trade_no } = params;
  // await updateOrder(out_trade_no, { status: "paid", tradeNo: trade_no });

  // 4. 返回 success 表示已收到通知
  return new NextResponse("success");
}
```

> **重要提示**：
> - 平台对每个订单会最多重试通知，请确保正确处理并返回 `success`
> - 务必使用异步通知（`notify_url`）作为支付确认依据，**切勿依赖同步跳转（`return_url`）**
> - 同步跳转有可能因为用户关闭浏览器等原因无法到达

---

## 4. 查询订单

通过商户订单号查询订单的支付状态。

### 4.1 请求 URL

```
https://zpayz.cn/api.php
```

### 4.2 请求方法

`GET` 或 `POST`

### 4.3 请求参数

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
|---|---|---|---|---|---|
| 操作类型 | `act` | 是 | String | `order` | 此 API 固定值 |
| 商户 ID | `pid` | 是 | String | `1001` | 平台分配的唯一商户 ID |
| 商户密钥 | `key` | 是 | String | `89unJUB8HZ54Hj7x...` | 商户通信密钥 |
| 商户订单号 | `out_trade_no` | 是 | String | `20160806151343349` | 待查询的商户订单号 |

### 4.4 请求示例

```
https://zpayz.cn/api.php?act=order&pid=1001&key=89unJUB8HZ54Hj7x4nUj56HN4nUzUJ8i&out_trade_no=20160806151343349
```

### 4.5 返回值

| 字段名 | 变量名 | 类型 | 示例值 | 描述 |
|---|---|---|---|---|
| 返回状态码 | `code` | Int | `1` | `1` 为成功，其他值为失败 |
| 返回信息 | `msg` | String | `查询订单号成功！` | — |
| 平台订单号 | `trade_no` | String | `2016080622555342651` | 平台内部订单号 |
| 商户订单号 | `out_trade_no` | String | `20160806151343349` | 商户系统内部的订单号 |
| 支付方式 | `type` | String | `alipay` | `alipay`：支付宝，`wxpay`：微信支付 |
| 商户 ID | `pid` | Int | `1001` | 发起支付的商户 ID |
| 创建时间 | `addtime` | String | `2016-08-06 22:55:52` | 订单创建时间 |
| 完成时间 | `endtime` | String | `2016-08-06 22:56:12` | 订单完成时间 |
| 支付状态 | `status` | Int | `1` | `1` 为支付成功，`0` 为未支付 |
| 订单金额 | `money` | String | `1.00` | 订单金额 |
| 商品名称 | `name` | String | `VIP会员` | 商品名称 |

### 4.6 查询示例

```typescript
async function queryOrder(pid: string, key: string, outTradeNo: string) {
  const params = new URLSearchParams({
    act: "order",
    pid,
    key,
    out_trade_no: outTradeNo,
  });

  const response = await fetch(`https://zpayz.cn/api.php?${params.toString()}`);
  const data = await response.json();

  if (data.code === 1) {
    return {
      paid: data.status === 1,
      tradeNo: data.trade_no,
      money: data.money,
      endTime: data.endtime,
    };
  }

  throw new Error(data.msg || "查询订单失败");
}
```

---

## 5. MD5 签名算法

### 5.1 签名步骤

> **核心公式**: `sign = md5( 排序后的参数字符串 + 商户密钥Key )`，结果为**小写**。

#### 步骤详解：

1. **过滤参数**：剔除 `sign`、`sign_type` 以及值为空的参数
2. **按参数名 ASCII 码从小到大排序（a→z）**
3. **拼接字符串**：将排序后的参数按 `key1=value1&key2=value2&...` 格式拼接，**参数值不进行 URL 编码**
4. **追加密钥**：在拼接后的字符串末尾直接拼接商户密钥 `Key`
5. **MD5 加密**：对上一步得到的字符串进行 MD5 运算，结果转为**小写**

### 5.2 签名示例

假设请求参数：

```
name = iPhone17苹果手机
money = 599.00
out_trade_no = ORDER20250501001
notify_url = https://www.example.com/notify.php
pid = 201901151314084206659771
return_url = https://www.example.com/return.php
type = alipay
商户密钥 Key = abc123def456 (假设)
```

**第一步**：过滤（无空值，无 sign/sign_type）

**第二步**：按 ASCII 排序 → `money`, `name`, `notify_url`, `out_trade_no`, `pid`, `return_url`, `type`

**第三步**：拼接字符串
```
money=599.00&name=iPhone17苹果手机&notify_url=https://www.example.com/notify.php&out_trade_no=ORDER20250501001&pid=201901151314084206659771&return_url=https://www.example.com/return.php&type=alipay
```

**第四步**：追加密钥
```
money=599.00&name=iPhone17苹果手机&notify_url=https://www.example.com/notify.php&out_trade_no=ORDER20250501001&pid=201901151314084206659771&return_url=https://www.example.com/return.php&type=alipayabc123def456
```

**第五步**：MD5 运算，结果小写
```
sign = md5(上述字符串).toLowerCase()
```

### 5.3 签名实现（TypeScript）

```typescript
import crypto from "crypto";

/**
 * 生成易支付 MD5 签名
 * @param params  请求参数对象（不含 sign 和 sign_type）
 * @param key     商户密钥
 * @returns       MD5 签名（小写）
 */
export function generateSign(
  params: Record<string, string | number | undefined>,
  key: string
): string {
  // 1. 过滤空值和 sign/sign_type
  const filtered = Object.entries(params).filter(
    ([k, v]) =>
      k !== "sign" && k !== "sign_type" && v !== "" && v !== undefined && v !== null
  );

  // 2. 按参数名 ASCII 排序 (a→z)
  filtered.sort(([a], [b]) => a.localeCompare(b));

  // 3. 拼接字符串
  const signStr = filtered.map(([k, v]) => `${k}=${v}`).join("&");

  // 4. 追加密钥
  const rawStr = signStr + key;

  // 5. MD5 加密，结果小写
  return crypto.createHash("md5").update(rawStr, "utf8").digest("hex").toLowerCase();
}
```

### 5.4 回调验签实现（TypeScript）

```typescript
/**
 * 验证回调通知的签名
 * @param params  回调中接收到的所有参数
 * @param key     商户密钥
 * @returns       验签是否通过
 */
export function verifySign(
  params: Record<string, string>,
  key: string
): boolean {
  const { sign, sign_type, ...rest } = params;
  const expectedSign = generateSign(rest, key);
  return expectedSign === sign;
}
```

---

## 6. 附录：状态码与错误码

### 6.1 接口返回码

| code | 含义 |
|---|---|
| `1` | 请求成功 |
| `-1` | 签名错误 |
| `-2` | 商户 ID 不存在 |
| 其他 | 失败（具体信息见 `msg` 字段） |

### 6.2 支付状态（`trade_status`）

| 值 | 含义 |
|---|---|
| `TRADE_SUCCESS` | 支付成功 |
| 其他值 | 未支付或支付失败 |

### 6.3 查询订单状态（`status`）

| 值 | 含义 |
|---|---|
| `0` | 未支付 |
| `1` | 支付成功 |

### 6.4 支付方式（`type`）

| 值 | 说明 |
|---|---|
| `alipay` | 支付宝 |
| `wxpay` | 微信支付 |
| `qqpay` | QQ 钱包 |
| `tenpay` | 财付通 |

---

---

## 7. 官方 Node.js Demo 参考

以下为 z-pay.cn 官方提供的 Node.js 接入示例代码：

```javascript
const utility = require("utility"); // MD5 第三方库

// 构造订单参数
let data = {
  pid: "你的pid",                         // 商户ID
  money: "金额",                          // 订单金额
  name: "商品名称",                        // 商品名称
  notify_url: "http://xxxxx",            // 异步通知地址
  out_trade_no: "2019050823435494926",   // 订单号，建议 YYYYMMDDHHmmss + 3位随机数
  return_url: "http://xxxx",             // 同步跳转地址
  sitename: "网站名称",                    // 网站名称
  type: "alipay",                        // alipay / wxpay / qqpay / tenpay
};

// 参数排序拼接（签名核心步骤）
function getVerifyParams(params) {
  var sPara = [];
  if (!params) return null;
  for (var key in params) {
    // 跳过空值、sign、sign_type
    if (!params[key] || key == "sign" || key == "sign_type") {
      continue;
    }
    sPara.push([key, params[key]]);
  }
  sPara = sPara.sort();      // 按参数名 ASCII 排序
  var prestr = "";
  for (var i2 = 0; i2 < sPara.length; i2++) {
    var obj = sPara[i2];
    if (i2 == sPara.length - 1) {
      prestr = prestr + obj[0] + "=" + obj[1] + "";
    } else {
      prestr = prestr + obj[0] + "=" + obj[1] + "&";
    }
  }
  return prestr;
}

// 生成签名
let str = getVerifyParams(data);
let key = "你的key";                       // 商户密钥
let sign = utility.md5(str + key);       // MD5(排序字符串 + key)

// 拼接支付跳转 URL
let result = `https://z-pay.cn/submit.php?${str}&sign=${sign}&sign_type=MD5`;
// 前端访问此 URL 即可跳转到收银台支付页面
```

> **注意**：Node.js 内置 `crypto` 模块也可直接替代 `utility` 进行 MD5 运算，无需额外安装依赖。

---

## 接入检查清单

对接完成前，请确认以下事项：

- [ ] 已从平台获取商户 ID（`pid`）和商户密钥（`key`）
- [ ] 实现了签名生成算法（MD5）
- [ ] 实现了异步通知接收端点（`notify_url`），并正确返回 `success`
- [ ] 异步通知处理中**验证了签名**的真实性
- [ ] 异步通知处理中**校验了金额**是否与订单一致
- [ ] 以异步通知为准更新订单状态，**不依赖同步跳转**
- [ ] 订单号生成策略确保全局唯一（建议使用 UUID 或雪花 ID）
- [ ] 记录完整的请求日志用于问题排查

---

> **参考来源**：[z-pay.cn 开发文档](https://z-pay.cn/doc.html)、易支付社区开源实现
