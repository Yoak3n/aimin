package tool

import (
	"fmt"
	"strings"

	"github.com/Yoak3n/aimin/blood/agent/skill"
)

func ComplexTaskForSkill(ctx *Context) string {
	p := ctx.GetPayload()
	if p == "" {
		return "args is empty"
	}
	args := parseArgsN(p, 1)
	p = strings.TrimSpace(firstNonEmpty(args["task"], args["_0"]))
	if p == "..." {
		skill.GlobalSkillHUB().Active = ""
		return "已清除当前激活技能"
	} else {
		skill.GlobalSkillHUB().Active = p
		return fmt.Sprintf("加载技能【%s】完成，已注入到上下文，请继续执行任务", p)
	}
}
