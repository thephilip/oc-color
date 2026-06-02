package theme

func init() {
	if builtins == nil {
		builtins = map[string]Theme{}
	}
	builtins["one-dark"] = oneDark()
}

func oneDark() Theme {
	return Theme{
		Name: "one-dark",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#98C379"},
			"warning": {Color: "#E5C07B"},
			"error":   {Color: "#E06C75", Bold: true},
			"info":    {Color: "#61AFEF"},
			"accent":  {Color: "#C678DD"},
			"dim":     {Color: "#5C6370"},
			"shade":   {Background: "#2C313C"},
			"header":  {Color: "#61AFEF", Bold: true, Underline: true},
			"key":     {Color: "#E5C07B"},
			"value":   {Color: "#ABB2BF"},
			"pink":    {Color: "#E06C75"},
			"orange":  {Color: "#D19A66"},
		},
	}
}
