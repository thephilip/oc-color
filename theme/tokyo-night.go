package theme

func init() {
	if builtins == nil {
		builtins = map[string]Theme{}
	}
	builtins["tokyo-night"] = tokyoNight()
}

func tokyoNight() Theme {
	return Theme{
		Name: "tokyo-night",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#9ECE6A"},
			"warning": {Color: "#E0AF68"},
			"error":   {Color: "#F7768E", Bold: true},
			"info":    {Color: "#7DCFFF"},
			"accent":  {Color: "#BB9AF7"},
			"dim":     {Color: "#565F89"},
			"shade":   {Background: "#16161E"},
			"header":  {Color: "#7AA2F7", Bold: true, Underline: true},
			"key":     {Color: "#E0AF68"},
			"value":   {Color: "#C0CAF5"},
			"pink":    {Color: "#F7768E"},
			"orange":  {Color: "#FF9E64"},
		},
	}
}
