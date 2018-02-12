package cmd

import (
	"shelastic/utils"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

// Snapshot wraps all snapshot functions
func Snapshot() *ishell.Cmd {
	snapshot := &ishell.Cmd{
		Name: "snapshot",
		Help: "Snapshot operations",
	}

	repo := &ishell.Cmd{
		Name: "repo",
		Help: "Repository operations",
	}

	repo.AddCmd(&ishell.Cmd{
		Name: "register",
		Help: "Registers repository. Usage: register <repo-name> <repo-type> [setting value [ setting value[ ...]]]",
		Func: createRepo,
	})

	repo.AddCmd(&ishell.Cmd{
		Name: "verify",
		Help: "Verifies repository. Usage: verify repo-name",
		Func: verifyRepo,
	})

	repo.AddCmd(&ishell.Cmd{
		Name: "list",
		Help: "List snapshot repository config",
		Func: listRepo,
	})

	snapshot.AddCmd(repo)

	snapshot.AddCmd(&ishell.Cmd{
		Name: "create",
		Help: "Creates snapshot of all open indices. Usage: create <repo-name> <snapshot-name>",
		Func: createSnapshot,
	})

	snapshot.AddCmd(&ishell.Cmd{
		Name: "info",
		Help: "Retrieves snapshot information. Usage: info <repo-name> [<snapshot-name>]",
		Func: getSnapshotInfo,
	})

	snapshot.AddCmd(&ishell.Cmd{
		Name: "delete",
		Help: "Deletes snapshot. Usage: delete <repo-name> <snapshot-name>",
		Func: deleteSnapshot,
	})

	snapshot.AddCmd(&ishell.Cmd{
		Name: "restore",
		Help: "Restores snapshot. Usage: restore <repo-name> <snapshot-name>",
		Func: restoreSnapshot,
	})

	return snapshot
}

func verifyRepo(c *ishell.Context) {
	if context != nil {
		if len(c.Args) < 1 {
			errorMsg(c, "Please specify name of repository")
			return
		}
		repoName := c.Args[0]
		err := context.VerifyRepository(repoName)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func createSnapshot(c *ishell.Context) {
	if context != nil {
		if len(c.Args) < 2 {
			errorMsg(c, "Please specify name of repository and snapshot name")
			return
		}
		repoName := c.Args[0]
		snapshotName := c.Args[1]
		err := context.CreateSnapshot(repoName, snapshotName)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func getSnapshotInfo(c *ishell.Context) {
	if context != nil {
		if len(c.Args) < 1 {
			errorMsg(c, "Please specify name of repository")
			return
		}
		repoName := c.Args[0]
		var snapshotName string
		if len(c.Args) >= 2 {
			snapshotName = c.Args[1]
		} else {
			snapshotName = "_all"
		}
		info, err := context.GetSnapshotInfo(repoName, snapshotName)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			result, err := utils.MapToYaml(info)
			if err != nil {
				errorMsg(c, err.Error())
			} else {
				cprintln(c, result)
			}
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func deleteSnapshot(c *ishell.Context) {
	if context != nil {
		if len(c.Args) < 2 {
			errorMsg(c, "Please specify name of repository and name of snapshot")
			return
		}
		repoName := c.Args[0]
		snapshot := c.Args[1]

		err := context.DeleteSnapshot(repoName, snapshot)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func restoreSnapshot(c *ishell.Context) {
	if context != nil {
		if len(c.Args) < 2 {
			errorMsg(c, "Please specify name of repository and name of snapshot")
			return
		}
		repoName := c.Args[0]
		snapshot := c.Args[1]

		err := context.RestoreSnapshot(repoName, snapshot)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func createRepo(c *ishell.Context) {
	if context != nil {
		if len(c.Args) < 2 {
			errorMsg(c, "Please specify name and type for repository")
			return
		}
		repoName := c.Args[0]
		repoType := c.Args[1]

		settings := readSettings(c.Args[2:])

		err := context.RegisterRepository(repoName, repoType, settings)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func listRepo(c *ishell.Context) {
	if context != nil {
		data, err := context.ListRepository()
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			yaml, err := utils.MapToYaml(data)
			if err != nil {
				errorMsg(c, err.Error())
			} else {
				cprintln(c, yaml)
			}
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func readSettings(args []string) map[string]string {
	result := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		var val string
		if i+1 <= len(args) {
			val = args[i+1]
		} else {
			val = ""
		}
		result[key] = val
	}
	return result
}
