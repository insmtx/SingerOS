# GitHub OAuth 集成

此文档描述如何设置 SingerOS 项目的 GitHub OAuth 集成。

## 说明 
SingerOS 支持通过 OAuth 实现 GitHub 用户认证，这将允许用户登录并与系统进行集成。

## 配置

要在 SingerOS 中启用 GitHub OAuth 功能，请在配置文件中设置以下属性：

```yaml
github:
  app_id: 123456                    # GitHub App 的 App ID
  private_key: "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"  # GitHub App 的私钥
  client_id: "your_github_oauth_client_id"     # GitHub OAuth 应用的 Client ID
  client_secret: "your_github_oauth_client_secret"  # GitHub OAuth 应用的 Client Secret
  webhook_secret: "your_webhook_secret"        # 用于验证 webhook 请求的密钥
  base_url: "https://api.github.com"           # GitHub API 的基础 URL
```

## 使用方法

### 启动认证流
用户可以通过访问以下网址开始 GitHub OAuth 认证流程：
```
GET /github/auth
```

这个请求会重定向到 GitHub 的 OAuth 授权页面.

### 处理回调
GitHub 会将授权码发送到以下回调地址：
```
GET /github/callback
```

这将在成功时返回带有访问令牌和用户信息的 JSON 响应.

## 数据持久化
认证的用户信息将保存在数据库的 `singer_users` 表中，包含以下字段:
- github_id: GitHub 上用户的唯一 ID
- github_login: 用户的 GitHub 登录名
- name: 用户的显示名称
- email: 用户的邮箱
- avatar_url: 用户头像链接
- bio: 个人简介
- company: 所属公司
- location: 位置
- public_repos: 公开仓库数量
- followers: 粉丝数

## 注意事项
- 确保你的 GitHub OAuth 应用已正确配置，且回调 URL 与 SingerOS 服务端点匹配
- OAuth 机制与 GitHub App 机制是不同的集成模式，请根据使用场景选择合适的配置