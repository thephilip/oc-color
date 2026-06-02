package theme

func init() {
	if builtins == nil {
		builtins = map[string]Theme{}
	}
	builtins["catppuccin"] = catppuccin()
}

func catppuccin() Theme {
	return Theme{
		Name: "catppuccin",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#A6E3A1"},
			"warning": {Color: "#F9E2AF"},
			"error":   {Color: "#F38BA8", Bold: true},
			"info":    {Color: "#89DCEB"},
			"accent":  {Color: "#CBA6F7"},
			"dim":     {Color: "#6C7086"},
			"shade":   {Background: "#181825"},
			"header":  {Color: "#89B4FA", Bold: true, Underline: true},
			"key":     {Color: "#F9E2AF"},
			"value":   {Color: "#CDD6F4"},
			"pink":    {Color: "#F38BA8"},
			"orange":  {Color: "#FAB387"},
		},
	}
}
