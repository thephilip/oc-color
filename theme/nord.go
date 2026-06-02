package theme

func init() {
	if builtins == nil {
		builtins = map[string]Theme{}
	}
	builtins["nord"] = nord()
}

func nord() Theme {
	return Theme{
		Name: "nord",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#A3BE8C"},
			"warning": {Color: "#EBCB8B"},
			"error":   {Color: "#BF616A", Bold: true},
			"info":    {Color: "#88C0D0"},
			"accent":  {Color: "#B48EAD"},
			"dim":     {Color: "#4C566A"},
			"shade":   {Background: "#3B4252"},
			"header":  {Color: "#81A1C1", Bold: true, Underline: true},
			"key":     {Color: "#EBCB8B"},
			"value":   {Color: "#ECEFF4"},
			"pink":    {Color: "#BF616A"},
			"orange":  {Color: "#D08770"},
		},
	}
}
