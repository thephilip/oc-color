package theme

func init() {
	if builtins == nil {
		builtins = map[string]Theme{}
	}
	builtins["gruvbox"] = gruvbox()
}

func gruvbox() Theme {
	return Theme{
		Name: "gruvbox",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#B8BB26"},
			"warning": {Color: "#FABD2F"},
			"error":   {Color: "#FB4934", Bold: true},
			"info":    {Color: "#83A598"},
			"accent":  {Color: "#D3869B"},
			"dim":     {Color: "#928374"},
			"shade":   {Background: "#3C3836"},
			"header":  {Color: "#83A598", Bold: true, Underline: true},
			"key":     {Color: "#FABD2F"},
			"value":   {Color: "#EBDBB2"},
			"pink":    {Color: "#D3869B"},
			"orange":  {Color: "#FE8019"},
		},
	}
}
