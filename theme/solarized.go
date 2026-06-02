package theme

func init() {
	if builtins == nil {
		builtins = map[string]Theme{}
	}
	builtins["solarized"] = solarized()
}

func solarized() Theme {
	return Theme{
		Name: "solarized",
		Tokens: map[string]TokenStyle{
			"success": {Color: "#859900"},
			"warning": {Color: "#B58900"},
			"error":   {Color: "#DC322F", Bold: true},
			"info":    {Color: "#268BD2"},
			"accent":  {Color: "#6C71C4"},
			"dim":     {Color: "#657B83"},
			"shade":   {Background: "#073642"},
			"header":  {Color: "#268BD2", Bold: true, Underline: true},
			"key":     {Color: "#B58900"},
			"value":   {Color: "#93A1A1"},
			"pink":    {Color: "#D33682"},
			"orange":  {Color: "#CB4B16"},
		},
	}
}
