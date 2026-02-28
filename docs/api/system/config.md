# 接口【/system/config】

## 接口概述

| 项目                 | 说明                                                         |
| -------------------- | ------------------------------------------------------------ |
| **完整接口路径**     | `/api/v1/system/config`                                      |
| **请求方式**         | `GET`                                                        |
| **认证要求**         | 无（公开接口）                                               |
| **限流规则**（TODO） | 同一 IP 每分钟最多 60 次                                     |
| **缓存策略**         | `Cache-Control: public, max-age=60`                          |
| **接口用途**         | 前端加载登录页时调用，获取系统基础配置、服务器时间、已开启的登录方式等信息，用于动态渲染登录页 |

## 调用时机

```Plain
用户访问任意页面
      │
      ▼
前端 App 初始化（App.vue onMounted）
      │
      ▼
GET /api/v1/system/config
      │
      ├── 根据 login.methods 渲染登录按钮列表
      ├── 根据 login.captcha_enabled 决定是否展示验证码组件
      ├── 根据 login.sso_force_redirect 决定是否直接跳转 SSO（TODO）
      ├── 根据 meta.server_timestamp_ms 校准本地时钟（TOTP 场景）
      └── 根据 meta.app_version 检测是否需要强制刷新页面
```

## 请求参数

无请求参数，无请求体。

**请求示例：**

```Bash
curl -X GET https://api.company.com/api/v1/system/config \
  -H "Accept: application/json"
```

## 响应字段说明

### 顶层结构

| 字段         | 类型    | 必返回 | 说明         |
| ------------ | ------- | ------ | ------------ |
| `code`       | integer | 是     | 固定为 0     |
| `message`    | string  | 是     | 固定为 "ok"  |
| `data`       | object  | 是     | 配置数据主体 |
| `request_id` | string  | 是     | 链路追踪 ID  |

### data.meta（系统元信息）

| 字段                  | 类型    | 必返回 | 说明                       | 前端用途                            |
| --------------------- | ------- | ------ | -------------------------- | ----------------------------------- |
| `app_version`         | string  | 是     | 业务版本号，格式 `vX.Y.Z`  | 检测版本变化，提示用户刷新页面      |
| `server_timestamp_ms` | integer | 是     | 服务器当前 Unix 毫秒时间戳 | TOTP 时钟校准；短信验证码倒计时同步 |

> **安全说明**：`app_version` 只返回业务版本号，不返回后端框架版本、Go 版本、commit hash 等技术信息，避免暴露攻击面。

### data.system（系统展示信息）

| 字段          | 类型   | 必返回 | 说明                     | 前端用途                     |
| ------------- | ------ | ------ | ------------------------ | ---------------------------- |
| `name`        | string | 是     | 系统名称                 | 登录页标题、浏览器 Tab 标题  |
| `logo_url`    | string | 否     | Logo 图片 URL，可为 null | 登录页顶部 Logo 展示         |
| `favicon_url` | string | 否     | Favicon URL，可为 null   | 浏览器标签页图标             |
| `copyright`   | string | 否     | 版权信息，可为 null      | 登录页底部版权文字           |
| `icp`         | string | 否     | ICP 备案号，可为 null    | 登录页底部备案号（国内合规） |

### data.login（登录策略）

| 字段              | 类型    | 必返回 | 说明                                                         | 前端用途                 |
| ----------------- | ------- | ------ | ------------------------------------------------------------ | ------------------------ |
| `methods`         | array   | 是     | 已启用的登录方式列表，见下方子结构                           | 动态渲染登录按钮         |
| `default_method`  | string  | 是     | 默认选中的登录方式 key                                       | 控制默认激活哪个登录 Tab |
| `captcha_enabled` | boolean | 是     | 是否开启图形验证码                                           | 决定是否渲染验证码组件   |
| `captcha_type`    | string  | 否     | 验证码类型：`image` / `slider`，captcha_enabled=false 时为 null | 决定渲染哪种验证码组件   |

**data.login.methods 子项结构：**

| 字段    | 类型   | 必返回 | 说明                                           |
| ------- | ------ | ------ | ---------------------------------------------- |
| `key`   | string | 是     | 登录方式唯一标识，提交登录时作为 `type` 字段值 |
| `label` | string | 是     | 展示名称（支持多语言）                         |

**method.key 枚举值：**

| key 值        | 说明                      | 登录时 type 字段 |
| ------------- | ------------------------- | ---------------- |
| `local`       | 本地账号（用户名 + 密码） | `local`          |
| `ldap`        | LDAP/AD 域账号            | `ldap`           |
| `wechat`      | 微信公众号 OAuth          | `wechat`         |
| `wechat_work` | 企业微信 OAuth            | `wechat_work`    |
| `dingtalk`    | 钉钉 OAuth                | `dingtalk`       |
| `feishu`      | 飞书 OAuth                | `feishu`         |

### data.security（安全策略，前端感知部分-暂无）

| 字段                                | 类型    | 必返回 | 说明                 | 前端用途                                         |
| ----------------------------------- | ------- | ------ | -------------------- | ------------------------------------------------ |
| `password_policy`                   | object  | 是     | 密码强度规则         | 注册/改密时实时校验密码强度，给用户即时反馈      |
| `password_policy.min_length`        | integer | 是     | 密码最小长度         | 表单校验规则                                     |
| `password_policy.require_uppercase` | boolean | 是     | 需要大写字母         | 表单校验规则                                     |
| `password_policy.require_lowercase` | boolean | 是     | 需要小写字母         | 表单校验规则                                     |
| `password_policy.require_number`    | boolean | 是     | 需要数字             | 表单校验规则                                     |
| `password_policy.require_special`   | boolean | 是     | 需要特殊字符         | 表单校验规则                                     |
| `max_login_attempts`                | integer | 是     | 最大连续失败次数     | 展示"还剩 N 次机会"提示（失败 2 次以上开始提示） |
| `lockout_minutes`                   | integer | 是     | 账号锁定时长（分钟） | 展示"账号已锁定，请 X 分钟后重试"倒计时          |

### data.client（客户端行为配置-暂无）

| 字段                          | 类型    | 必返回 | 说明                        | 前端用途                                         |
| ----------------------------- | ------- | ------ | --------------------------- | ------------------------------------------------ |
| `access_token_expire_seconds` | integer | 是     | access_token 有效期（秒）   | 计算 token 刷新时机（过期前 5 分钟触发静默刷新） |
| `captcha_expire_seconds`      | integer | 是     | 验证码有效期（秒）          | 验证码组件展示倒计时或过期提示                   |
| `otp_expire_seconds`          | integer | 是     | 短信/邮件验证码有效期（秒） | 发送验证码后的倒计时展示                         |

### data.agreement（用户协议-暂无）

| 字段          | 类型   | 必返回 | 说明                        | 前端用途               |
| ------------- | ------ | ------ | --------------------------- | ---------------------- |
| `terms_url`   | string | 否     | 服务条款页面 URL，可为 null | 登录页「服务条款」链接 |
| `privacy_url` | string | 否     | 隐私政策页面 URL，可为 null | 登录页「隐私政策」链接 |

## 响应示例

### 成功响应（200 OK）

```JSON
{
  "code": 0,
  "message": "ok",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "data": {
    "meta": {
      "app_version": "v1.2.0",
      "server_timestamp_ms": 1740567600000
    },
    "system": {
      "name": "企业管理平台",
      "logo_url": "https://cdn.company.com/assets/logo.png",
      "favicon_url": "https://cdn.company.com/assets/favicon.ico",
      "copyright": "© 2026 Company Inc. All rights reserved.",
      "icp": "粤ICP备XXXXXXXX号"
    },
    "login": {
      "methods": [
        {
          "key": "local",
          "label": "账号密码登录",
          "icon": "user",
          "sort": 1
        },
        {
          "key": "ldap",
          "label": "域账号登录",
          "icon": "building",
          "sort": 2
        },
        {
          "key": "wechat_work",
          "label": "企业微信登录",
          "icon": "https://cdn.company.com/icons/wechat-work.svg",
          "sort": 3
        }
      ],
      "default_method": "local",
      "captcha_enabled": true,
      "captcha_type": "image",
      "sso_force_redirect": false,
      "sso_redirect_url": null
    },
    "security": {
      "password_policy": {
        "min_length": 10,
        "require_uppercase": true,
        "require_lowercase": true,
        "require_number": true,
        "require_special": false
      },
      "max_login_attempts": 5,
      "lockout_minutes": 30
    },
    "client": {
      "access_token_expire_seconds": 7200,
      "captcha_expire_seconds": 120,
      "otp_expire_seconds": 300
    },
    "agreement": {
      "terms_url": "/terms",
      "privacy_url": "/privacy"
    }
  }
}
```

### 账号密码错误（401）

```JSON
{
  "code": 4011,
  "message": "认证错误，请检查账号密码是否正确",
  "data": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

### 验证码错误（422 Unprocessable Entity）

```JSON
{
  "code": 4011,
  "message": "验证码错误，请检查后重新提交",
  "data": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

### 服务器错误响应（500）

```JSON
{
  "code": 5001,
  "message": "服务器内部错误",
  "data": null,
  "request_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

### 限流响应（429）

```JSON
{
  "code": 4291,
  "message": "请求过于频繁，请稍后再试",
  "data": {
    "retry_after_seconds": 30
  },
  "request_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

## 安全设计说明

此接口为完全公开接口，设计时需严格遵守以下原则：

| 原则           | 说明                                                         |
| -------------- | ------------------------------------------------------------ |
| 最小暴露       | 只返回前端渲染登录页必须的字段，不返回任何内部配置细节       |
| 不暴露技术栈   | 禁止返回后端框架版本、数据库类型、Go 版本、commit hash 等信息 |
| 不暴露内部 ID  | 禁止返回 provider_id 等数据库主键                            |
| 不暴露密钥     | 禁止返回 AppID、AppSecret、LDAP 地址等认证源配置信息         |
| 不暴露用户数据 | 禁止返回用户总数、注册人数等统计信息                         |

## 缓存策略说明

```Plain
响应 Header：
  Cache-Control: public, max-age=60
  ETag: "{config_version_hash}"

设计理由：
  此接口数据变化频率低（Admin 改配置才变化）
  允许浏览器和 CDN 缓存 60 秒
  命中缓存时前端响应时间约 0ms
  配置变更后通过 ETag 机制让客户端感知变化

前端处理：
  App 初始化时必须请求一次（不能完全依赖缓存）
  后续路由切换可直接使用内存中的配置，无需重复请求
```
