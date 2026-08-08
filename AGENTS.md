# Sub2API 本地定制保护规则

用户的最新明确指令优先于本文件。检查或合并上游更新时，先比较当前分支、本地定制提交和目标上游提交，不要用上游文件整体覆盖本地实现。

## 上游合并禁止回灌

- 登录后的部署/运营/合规确认门禁必须保持禁用。不得恢复 `AdminComplianceDialog` 挂载、`admin-compliance-required` 事件、管理员路由合规状态预取，或后端 `423 ADMIN_COMPLIANCE_ACK_REQUIRED` 阻断；后端 guard 必须继续放行，状态默认 `required=false`。
- 用户端 `/subscriptions` 必须继续重定向到 `/redeem`，管理端 `/admin/subscriptions` 必须继续重定向到 `/admin/redeem`。
- 注册、邮箱验证和后台设置中的优惠码前端流程保持下线；合并验证码更新时可以保留 Turnstile、腾讯和阿里云验证码，但不得恢复 `promo_code` 输入、校验或请求字段。
- 新手引导保持移除，不得恢复 onboarding store、composable、样式、步骤或页面触发逻辑。
- 兑换码批量选择后的复制能力必须保留。
- `deploy/docker-compose.vps04.yml` 是 VPS04 低内存部署配置，合并上游时不得删除或被普通 Compose 配置覆盖。
- 上游新增的利润控制、倍率自动同步和验证码服务商等可选功能只合入代码，默认保持关闭；未经用户明确要求不要顺手启用。

## 验证

- 合并后至少运行 `git diff --check`。
- 在 `backend` 目录运行合规门禁和本次上游更新相关测试；条件允许时运行 `go test ./...`。
- 在 `frontend` 目录运行 `pnpm typecheck`，并运行注册、邮箱验证、API client、侧边栏和兑换码相关测试。
- 部署前必须备份数据库、当前镜像和运行配置；只重建 `sub2api` 应用容器，不无故重启 PostgreSQL、Redis 或其他业务。
