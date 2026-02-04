package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化 GoClaw 身份配置",
	Long:  `通过交互式问卷创建个性化的 IDENTITY.md 和 SOUL.md 文件`,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	color.Cyan("╔════════════════════════════════════════╗")
	color.Cyan("║     GoClaw 身份配置向导               ║")
	color.Cyan("╚════════════════════════════════════════╝")
	color.White("\n这将帮助创建个性化的身份配置文件。\n")

	reader := bufio.NewReader(os.Stdin)

	// 收集用户信息
	identity := make(map[string]string)

	color.Yellow("## 基本信息\n")
	identity["name"] = prompt(reader, "你的名字: ")
	identity["occupation"] = prompt(reader, "你的职业: ")
	identity["location"] = prompt(reader, "你所在的城市/地区: ")
	identity["timezone"] = prompt(reader, "你的时区 (例如: Asia/Shanghai): ")

	color.Yellow("\n## 兴趣与专长\n")
	interests := prompt(reader, "兴趣爱好 (用逗号分隔): ")
	identity["interests"] = interests
	expertise := prompt(reader, "专长技能 (用逗号分隔): ")
	identity["expertise"] = expertise

	color.Yellow("\n## 沟通偏好\n")
	color.White("回复风格选项:")
	color.White("  1. 简洁 - 直接回答，避免冗余")
	color.White("  2. 详细 - 提供完整的解释和背景")
	color.White("  3. 幽默 - 轻松诙谐的表达方式")
	style := prompt(reader, "选择回复风格 (1-3, 默认1): ")
	if style == "" {
		style = "1"
	}
	styleMap := map[string]string{"1": "简洁", "2": "详细", "3": "幽默"}
	identity["style"] = styleMap[style]

	techLevel := prompt(reader, "技术深度偏好 (高/中/低, 默认: 中): ")
	if techLevel == "" {
		techLevel = "中"
	}
	identity["techLevel"] = techLevel

	color.Yellow("\n## 工作习惯\n")
	identity["workHours"] = prompt(reader, "工作时间 (例如: 9:00-18:00): ")
	identity["tools"] = prompt(reader, "常用工具 (用逗号分隔): ")

	color.Yellow("\n## 个性特征 (可选)")
	identity["personality"] = prompt(reader, "个性特点描述 (可选，按回车跳过): ")

	// 确认
	color.Yellow("\n════════════════════════════════════════")
	color.White("配置预览:")
	for key, value := range identity {
		if value != "" {
			color.White("  %s: %s", key, value)
		}
	}

	color.Yellow("\n════════════════════════════════════════")
	confirm := prompt(reader, "\n确认生成配置文件？(y/n): ")
	if strings.ToLower(confirm) != "y" {
		color.Yellow("已取消。")
		return nil
	}

	// 生成文件
	workspaceDir := os.Getenv("HOME") + "/.goclaw/workspace"
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return fmt.Errorf("创建 workspace 目录失败: %w", err)
	}

	// 生成 IDENTITY.md
	identityContent := generateIdentityMD(identity)
	identityPath := filepath.Join(workspaceDir, "IDENTITY.md")
	if err := os.WriteFile(identityPath, []byte(identityContent), 0644); err != nil {
		return fmt.Errorf("写入 IDENTITY.md 失败: %w", err)
	}

	// SOUL.md 保持默认
	// 如果用户提供了个性描述，可以追加到 SOUL.md
	if identity["personality"] != "" {
		soulPath := filepath.Join(workspaceDir, "SOUL.md")
		existingSoul, _ := os.ReadFile(soulPath)
		additional := fmt.Sprintf("\n## 用户自定义个性\n\n%s\n", identity["personality"])
		updatedSoul := string(existingSoul) + additional
		if err := os.WriteFile(soulPath, []byte(updatedSoul), 0644); err != nil {
			return fmt.Errorf("更新 SOUL.md 失败: %w", err)
		}
	}

	color.Green("\n✓ 配置文件已生成！")
	color.White("\n文件位置:")
	color.White("  - %s", identityPath)
	color.White("  - %s", filepath.Join(workspaceDir, "SOUL.md"))
	color.White("\n你可以随时手动编辑这些文件来调整我的个性。")
	color.White("\n现在可以开始对话了：")
	color.Cyan("  goclaw chat\n")

	return nil
}

func prompt(reader *bufio.Reader, text string) string {
	color.Green("%s", text)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func generateIdentityMD(identity map[string]string) string {
	return fmt.Sprintf(`# 身份档案

这是 GoClaw 的身份文件，记录了关于你的重要信息。

## 基本信息

- **姓名**: %s
- **职业**: %s
- **位置**: %s
- **时区**: %s

## 兴趣与专长

- **兴趣**: %s
- **专长**: %s

## 沟通偏好

- **语言**: 中文
- **回复风格**: %s
- **技术深度**: %s - 对技术细节的偏好程度

## 工作习惯

- **工作时间**: %s
- **常用工具**: %s

## 个性特征

%s

---

*此文件由 goclaw init 生成，可以随时手动编辑。*`,
		identity["name"],
		identity["occupation"],
		identity["location"],
		identity["timezone"],
		identity["interests"],
		identity["expertise"],
		identity["style"],
		identity["techLevel"],
		identity["workHours"],
		identity["tools"],
		identity["personality"],
	)
}
