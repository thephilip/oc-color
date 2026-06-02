package theme

func init() {
	if builtins == nil {
		builtins = make(map[string]Theme)
	}
	builtins["dracula"] = dracula()
}

func dracula() Theme {
	return Theme{
		Name: "dracula",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#50FA7B"},
			"warning": {Color: "#F1FA8C"},
			"error":   {Color: "#FF5555", Bold: true},
			"info":    {Color: "#8BE9FD"},
			"accent":  {Color: "#BD93F9"},
			"pink":    {Color: "#FF79C6"},
			"orange":  {Color: "#FFB86C"},
			"dim":     {Color: "#6272A4"},
			"shade":   {Background: "#2E3040"},
			"header":  {Color: "#BD93F9", Bold: true, Underline: true},
			"key":     {Color: "#F1FA8C"},
			"value":   {Color: "#F8F8F2"},
		},
	}
}
