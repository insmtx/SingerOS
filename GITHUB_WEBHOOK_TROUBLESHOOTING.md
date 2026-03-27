# GitHub Webhook 签名验证问题排查指南

## 错误信息
```
[WARN] Invalid GitHub webhook signature for request: POST /github/webhook
```

## 排查步骤

### 1. 检查 webhook secret 配置
确保配置文件中的 `webhook_secret` 与 GitHub 应用设置中的一致：

```yaml
github:
  webhook_secret: "your_webhook_secret_here"
```

**重要说明**：
- 不要在 secret 前后添加额外的空格或换行符
- 确保没有复制到多余的字符
- secret 是大小写敏感的

### 2. 在 GitHub 中验证 Webhook 设置
1. 访问 GitHub 仓库设置 → Webhooks
2. 选择你的 Webhook
3. 检查 "Secret" 字段是否与配置文件一致

### 3. 查看 Recent Deliveries
在 GitHub Webhook 页面底部：
1. 查看 "Recent Deliveries" 部分
2. 选择最近失败的请求
3. 查看：
   - Request 标签页：确认请求已正确发送
   - Response 标签页：查看服务器返回的错误
   - Headers 标签页：检查是否有 `X-Hub-Signature-256` 头

### 4. 临时禁用签名验证（仅用于测试）

**警告**：仅用于开发环境调试，不要在生产环境使用！

修改 `webhook.go`，临时注释掉签名验证：

```go
// if !c.verifySignature(r, payload) {
//     logs.Warnf("Invalid GitHub webhook signature for request: %s %s", r.Method, r.URL.Path)
//     http.Error(w, "invalid signature", 401)
//     return
// }
```

如果禁用验证后 webhook 能正常工作，说明问题确实在 secret 配置上。

### 5. 添加调试日志

在 `verifySignature` 函数中添加调试信息：

```go
func (c *GitHubConnector) verifySignature(
	r *http.Request,
	payload []byte,
) bool {
	signature := r.Header.Get("X-Hub-Signature-256")
	logs.Debugf("Received signature: %s", signature)
	logs.Debugf("Secret length: %d", len(c.config.WebhookSecret))
	
	mac := hmac.New(sha256.New, []byte(c.config.WebhookSecret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	logs.Debugf("Expected signature: %s", expected)
	
	result := hmac.Equal([]byte(signature), []byte(expected))
	logs.Debugf("Signature match: %v", result)
	
	return result
}
```

### 6. 测试 Secret 有效性

创建一个简单的测试脚本验证 secret 是否正确：

```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func main() {
	secret := "your_webhook_secret" // 替换为你的 secret
	payload := []byte(`{"test": "payload"}`) // 替换为实际 payload
	
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	fmt.Println("Expected signature:", expected)
}
```

### 7. 常见问题

#### 问题 A: Secret 包含特殊字符
如果你的 secret 包含特殊字符，确保在 YAML 中正确转义：

```yaml
# 包含特殊字符时使用引号
webhook_secret: "my-secret-with-!@#$%^&*()"
```

#### 问题 B: 多次复制粘贴
有时复制粘贴会引入不可见字符。尝试直接手动输入 secret。

#### 问题 C: 环境变量问题
如果通过环境变量配置，确保没有额外的空格或换行符。

#### 问题 D: 多个 Webhooks
确保你在 GitHub 上修改的是正确的 webhook。

### 8. 重新生成 Webhook Secret

如果以上步骤都无法解决问题：

1. 在 GitHub Webhook 设置中生成一个新的 secret
2. 更新配置文件中的 `webhook_secret`
3. 重启 SingerOS 服务
4. 在 GitHub 上测试 webhook

## 验证成功后

确认 webhook 工作正常后：
1. 记得移除临时添加的调试日志
2. 恢复签名验证（如果之前禁用了）
3. 确保所有测试通过

## 需要帮助？

如果按照以上步骤仍无法解决问题，请收集以下信息：
- SingerOS 配置文件（隐藏敏感信息）
- GitHub Webhook 设置截图
- Recent Deliveries 中的 Request/Response 详情
- 服务器日志