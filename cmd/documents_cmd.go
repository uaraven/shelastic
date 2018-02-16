package cmd

import (
	ishell "gopkg.in/abiosoft/ishell.v2"
)

// UseIndex selects an index to use with document operations
func UseIndex() *ishell.Cmd {

	return &ishell.Cmd{
		Name: "use",
		Help: "Select index to use for subsequent document operations",
		Func: func(c *ishell.Context) {
			if context == nil {
				errorMsg(c, errNotConnected)
			} else {
				if len(c.Args) < 1 {
					errorMsg(c, "Index name not specified")
					return
				}
				s, err := context.ResolveAndValidateIndex(c.Args[0])
				if err != nil {
					errorMsg(c, err.Error())
					return
				}
				context.ActiveIndex = s
				if s != c.Args[0] {
					cprintln(c, "For alias %s selected index %s", c.Args[0], gr(s))
				} else {
					cprintln(c, "Selected index %s", s)
				}
			}
		},
	}

}
