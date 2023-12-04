review by {{.Reviewer}}:
| 审视项目编号 | 审视类别 | 审视说明 | 审视意见 | 审视评价 |
| ---- | ---- | ---- | ---- | ---- |
{{- range .CheckItems }}
| {{.Id}} | {{.Name}} | {{.Desc}} | {{.Result}} | {{.Comment}} |
{{- end}}