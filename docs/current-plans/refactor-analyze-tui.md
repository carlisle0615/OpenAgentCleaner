# Refactor Analyze TUI

## Goal

优化 `analyze` 会话面板性能：将会话预览从“进入列表即预加载 / 上下移动即重载”调整为“按 Enter 时按需加载”，降低 Cursor/Codex 等重会话数据下的卡顿。

## Touched Files

- `internal/cleaner/analyze_tui.go`
- `internal/cleaner/analyze_tui_test.go`

## Validation Steps

1. 运行 `go test ./internal/cleaner -run Analyze` 验证 analyze TUI 相关用例。
2. 运行 `make test` 验证全量单元测试。
3. 运行 `make build` 验证可构建性。
4. 手动验证 `oac analyze --assistant cursor` / `oac analyze --assistant codex`：
   - 进入 Conversations 列表时不立即加载预览。
   - 上下移动选择时不触发预览重算。
   - 按 Enter 后才加载并进入预览页。

## Doc-Sync Steps

- 在本计划记录本次“按需加载”改造内容与验证结果。
- 更新 `docs/current-plan.md` 的 Last updated 时间戳。

## Implementation Notes (2026-03-10)

- `reloadSessions()` 不再预加载选中会话预览，改为只清空预览缓存状态。
- 会话列表 `↑/↓` 移动时不再触发 `loadSelectedSessionPreview()`，仅重置缓存。
- 在 `screenSessions` 下按 Enter 时才调用 `loadSelectedSessionPreview()` 并进入预览页。
- 详情面板新增按需加载提示：未加载时显示 “Preview is loaded on demand. Press Enter to load.”
- 新增 `clearSelectedSessionPreview()` / `hasSelectedSessionPreview()` 辅助函数，确保预览与当前选中会话一致。
- 更新测试：`reloadSessions` 不应预加载预览；`activateSelection` 时才加载预览。
