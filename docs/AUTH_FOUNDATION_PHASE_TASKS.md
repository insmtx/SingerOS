# 账户授权底座阶段任务清单

## 目标

先实现“用户自己授权账户，并在运行时被系统复用”的底座能力。

当前阶段不做：

- 数据库表
- 审批流
- 多 provider 全量接入
- 完整 tool runtime
- workflow 编排

当前阶段只做：

- 多 provider 统一授权模型
- 内存版授权存储
- 用户 OAuth 授权接入
- 运行时账户解析
- 第一个 provider 的 client factory

## 阶段拆分

### Phase 1：多 Provider 授权基础模型

目标：

- 建立通用授权对象模型
- 不把模型写死为 GitHub

任务：

1. 定义 `AuthorizedAccount`
2. 定义 `AccountCredential`
3. 定义 `UserProviderBinding`
4. 定义 `OAuthState`
5. 定义 provider / account_type / grant_type 常量

完成标准：

- 后续任意 provider 都可以复用这套模型

### Phase 2：内存版授权存储

目标：

- 在不建表的前提下，把用户授权账户统一放在内存里管理

任务：

1. 增加 `AuthorizedAccountStore` 接口
2. 增加 `InMemoryAuthorizedAccountStore`
3. 支持保存 OAuth state
4. 支持保存账户与凭证
5. 支持设置和读取用户默认账户

完成标准：

- 单进程内可以完成授权接入和默认账户解析

### Phase 3：账户解析能力

目标：

- 后续 tool 可以按 `user + provider` 找到账户

任务：

1. 实现 `AccountResolver`
2. 支持显式 `account_id`
3. 支持按 `user + provider` 找默认账户
4. 支持找第一个可用账户兜底
5. 找不到时返回“需要先授权”

完成标准：

- 运行时能统一解析该用哪个账户

### Phase 4：GitHub 用户授权接入

目标：

- 让使用者自己完成 GitHub OAuth 授权

任务：

1. 新增 GitHub OAuth provider
2. 支持发起授权 URL
3. 支持 callback 处理
4. 使用 token 拉取 GitHub 用户资料
5. 保存授权账户和 credential
6. 首次授权时自动设为默认账户

完成标准：

- 用户可以通过 HTTP 接口完成 GitHub 账户授权

### Phase 5：GitHub Client Factory

目标：

- 把运行时已授权账户转换成 GitHub client

任务：

1. 新增 `GithubClientFactory`
2. 从 resolver 获取账户
3. 从 store 获取 credential
4. 生成 `go-github` client

完成标准：

- 上层不需要传 token，只需要传 `userID`

### Phase 6：接入现有 Connector

目标：

- 让当前 GitHub connector 的 auth 路由使用新授权底座

任务：

1. 在启动时初始化 auth service
2. 将其注入 GitHub connector
3. 替换 `/github/auth`
4. 替换 `/github/callback`

完成标准：

- 现有 GitHub 授权入口变成真正的用户授权入口

### Phase 7：校验与最小验证

目标：

- 保证基础能力可编译、可单测

任务：

1. 为内存 store 写测试
2. 为 resolver 写测试
3. 跑授权相关包测试
4. 跑最小构建检查

完成标准：

- 授权基础模块可稳定被后续 tool runtime 复用

## 当前实现顺序

本轮实际开发先落：

1. Phase 1
2. Phase 2
3. Phase 3
4. Phase 4
5. Phase 5 的 GitHub OAuth user client 路径
6. Phase 6 的最小接入

暂不实现：

- token refresh
- provider 审批策略
- 多 provider client factory 注册中心
- tool runtime 与 workflow engine 的联动
