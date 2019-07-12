// +build go1.1

// Copyright 2013 gopm authors.
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// gopm(Go Package Manager) is a Go package manage tool for searching, installing, updating and sharing your packages in Go.
package main

import (
	"os"
	"runtime"

	"github.com/codegangsta/cli"

	"github.com/gpmgo/gopm/cmd"
)

const APP_VER = "0.6.5.0524"

//		cmd.CmdTest,
//		cmd.CmdSearch,
// 		cmdClean,
// 		cmdDoc,
// 		cmdEnv,
// 		cmdFix,
// 		cmdList,
// 		cmdTool,
// 		cmdVet,

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	app := cli.NewApp()
	app.Name = "gopm"
	app.Usage = "Go Package Manager"
	app.Version = APP_VER
	app.Commands = []cli.Command{
		cmd.CmdGet,
		cmd.CmdBin,
		// cmd.CmdExec,
		cmd.CmdGen,
		cmd.CmdRun,
		cmd.CmdBuild,
		cmd.CmdInstall,
		cmd.CmdUpdate,
		cmd.CmdConfig,
	}
	app.Flags = append(app.Flags, []cli.Flag{
		cli.BoolFlag{"noterm", "disable color output"},
	}...)
	app.Run(os.Args)
}
