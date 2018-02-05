package commands

import (
	"fmt"

	cmds "gx/ipfs/QmWGgKRz5S24SqaAapF5PPCfYfLT7MexJZewN5M82CQTzs/go-ipfs-cmds"
	cmdkit "gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit"

	"github.com/filecoin-project/go-filecoin/flags"
)

var versionCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "show the version",
	},
	Run: versionRun,
}

func versionRun(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) {
	if err := re.Emit(fmt.Sprintf("commit: %s\n", flags.Commit)); err != nil {
		// TODO: is there a better way to handle errors on emit?
		panic(err)
	}
}