# Refactor Analyze TUI

## Goal

完全重构 `analyze` 命令的 TUI（终端用户界面），通过更现代的配色、布局、边框（RoundedBorder），以及更清晰的块级划分，大幅提升视觉和交互体验。

## Touched Files

- `internal/cleaner/analyze_tui.go`

## Validation Steps

1. 运行 `make build` 确保核心逻辑及语法无误。
2. 运行 `./OpenAgentCleaner analyze` （或直接 `go run cmd/openagentcleaner/main.go analyze`）手动检查界面表现，确保：
   - 界面自适应终端大小，不出现文字截断或错乱。
   - 颜色（Lipgloss Colors）在终端中渲染清晰，无对比度问题。
   - 快捷键提示、弹窗与列表交互反馈明显。

## Doc-Sync Steps

- 更新 `docs/current-plan.md` 添加此计划链接。
- 在本计划中记录步骤。
- 最终完成时，在 `docs/handoffs/refactor-analyze-tui.md` 中记录变更。
