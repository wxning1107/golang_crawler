// Copyright 2013-2014 gopm authors.
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

package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/Unknwon/com"
	"github.com/Unknwon/goconfig"
	"github.com/codegangsta/cli"

	"github.com/gpmgo/gopm/doc"
	"github.com/gpmgo/gopm/log"
)

var (
	installRepoPath string // The path of gopm local repository.
	installGopath   string // The first path in the GOPATH.
	//review:in the code it is not used , instead ctx.bool("gopath") is used
	isHasGopath bool // Indicates whether system has GOPATH.

	downloadCache map[string]bool // Saves packages that have been downloaded.
	downloadCount int
	failConut     int
)

var CmdGet = cli.Command{
	Name:  "get",
	Usage: "fetch remote package(s) and dependencies to local repository",
	Description: `Command get fetches a package, and any pakcage that it depents on. 
If the package has a gopmfile, the fetch process will be driven by that.

gopm get
gopm get <import path>@[<tag|commit|branch>:<value>]
gopm get <package name>@[<tag|commit|branch>:<value>]

Can specify one or more: gopm get beego@tag:v0.9.0 github.com/beego/bee

If no version specified and package exists in GOPATH,
it will be skipped unless user enabled '--remote, -r' option 
then all the packages go into gopm local repository.`,
	Action: runGet,
	Flags: []cli.Flag{
		cli.BoolFlag{"gopath, g", "download all pakcages to GOPATH"},
		cli.BoolFlag{"update, u", "update pakcage(s) and dependencies if any"},
		cli.BoolFlag{"example, e", "download dependencies for example folder"},
		cli.BoolFlag{"remote, r", "download all pakcages to gopm local repository"},
		cli.BoolFlag{"verbose, v", "show process details"},
		cli.BoolFlag{"local,l", "download all packages to local gopath"},
	},
}

func init() {
	downloadCache = make(map[string]bool)
}

func runGet(ctx *cli.Context) {
	setup(ctx)
	// Check conflicts.
	if ctx.Bool("gopath") && ctx.Bool("remote") ||
		ctx.Bool("local") && ctx.Bool("remote") ||
		ctx.Bool("gopath") && ctx.Bool("local") {
		e := " "
		if ctx.Bool("gopath") {
			e += "--gopth,-g "
		}
		if ctx.Bool("remote") {
			e += "--remote,-r "
		}
		if ctx.Bool("local") {
			e += "--local,-l "
		}
		log.Error("get", "Command options have conflicts")
		log.Error("", "Following options are not supposed to use at same time:")
		log.Error("", "\t"+e)
		log.Help("Try 'gopm help get' to get more information")
	}

	if !ctx.Bool("remote") {
		// Get GOPATH.
		installGopath = com.GetGOPATHs()[0]
		if com.IsDir(installGopath) {
			isHasGopath = true
			log.Log("Indicated GOPATH: %s", installGopath)
			installGopath += "/src"
		} else {
			if ctx.Bool("gopath") {
				log.Error("get", "Invalid GOPATH path")
				log.Error("", "GOPATH does not exist or is not a directory:")
				log.Error("", "\t"+installGopath)
				log.Help("Try 'go help gopath' to get more information")
			} else {
				// It's OK that no GOPATH setting
				// when user does not specify to use.
				log.Warn("No GOPATH setting available")
			}
		}
		// if flag local set use localPath as GOPATH
		if ctx.Bool("local") {
			if !com.IsExist(".gopmfile") {
				runGen(ctx)
			}
			gf, err := goconfig.LoadConfigFile(".gopmfile")
			if err != nil {
				log.Fatal("get", err.Error())
			} else {
				installGopath = gf.MustValue("project", "localPath")
				if installGopath == "" {
					os.Remove(".gopmfile")
					log.Fatal("get", "unexpected localPath or no localPath exists")
				}
				installGopath += "/src"
			}
		}
	}

	// The gopm local repository.
	installRepoPath = path.Join(doc.HomeDir, "repos")
	log.Log("Local repository path: %s", installRepoPath)

	// Check number of arguments to decide which function to call.
	switch len(ctx.Args()) {
	case 0:
		getByGopmfile(ctx)
	case 1:
		getByPath(ctx)
	default:
		log.Error("get", "too many arguments")
		log.Help("Try 'gopm help get' to get more information")
	}
}

func getByGopmfile(ctx *cli.Context) {
	// Check if gopmfile exists and generate one if not.
	if !com.IsFile(".gopmfile") {
		runGen(ctx)
	}
	gf := doc.NewGopmfile(".")

	targetPath := parseTarget(gf.MustValue("target", "path"))
	// Get dependencies.
	imports := doc.GetAllImports([]string{workDir}, targetPath, ctx.Bool("example"), false)

	nodes := make([]*doc.Node, 0, len(imports))
	for _, p := range imports {
		// TODO: DOING TEST CASES!!!
		p = doc.GetProjectPath(p)
		// Skip subpackage(s) of current project.
		if isSubpackage(p, targetPath) {
			continue
		}
		node := doc.NewNode(p, p, doc.BRANCH, "", true)

		// Check if user specified the version.
		if v, err := gf.GetValue("deps", p); err == nil && len(v) > 0 {
			node.Type, node.Value = validPath(v)
		}
		nodes = append(nodes, node)
	}

	downloadPackages(ctx, nodes)
	//save vcs infromation in the .gopm/data
	doc.SaveLocalNodes()

	log.Log("%d package(s) downloaded, %d failed", downloadCount, failConut)
}

func getByPath(ctx *cli.Context) {
	nodes := make([]*doc.Node, 0, len(ctx.Args()))
	for _, info := range ctx.Args() {
		pkgPath := info
		node := doc.NewNode(pkgPath, pkgPath, doc.BRANCH, "", true)

		if i := strings.Index(info, "@"); i > -1 {
			pkgPath = info[:i]
			tp, ver := validPath(info[i+1:])
			node = doc.NewNode(pkgPath, pkgPath, tp, ver, true)
		}

		// Check package name.
		if !strings.Contains(pkgPath, "/") {
			pkgPath = doc.GetPkgFullPath(pkgPath)
		}

		node.ImportPath = pkgPath
		node.DownloadURL = pkgPath
		nodes = append(nodes, node)
	}

	downloadPackages(ctx, nodes)
	doc.SaveLocalNodes()

	log.Log("%d package(s) downloaded, %d failed", downloadCount, failConut)
}

func copyToGopath(srcPath, destPath string) {
	importPath := strings.TrimPrefix(destPath, installGopath+"/")
	if len(getVcsName(destPath)) > 0 {
		log.Warn("Package in GOPATH has version control: %s", importPath)
		return
	}

	os.RemoveAll(destPath)
	err := com.CopyDir(srcPath, destPath)
	if err != nil {
		log.Error("download", "Fail to copy to GOPATH:")
		log.Fatal("", "\t"+err.Error())
	}

	log.Log("Package copied to GOPATH: %s", importPath)
}

// downloadPackages downloads packages with certain commit,
// if the commit is empty string, then it downloads all dependencies,
// otherwise, it only downloada package with specific commit only.
func downloadPackages(ctx *cli.Context, nodes []*doc.Node) {
	// Check all packages, they may be raw packages path.
	for _, n := range nodes {
		// Check if local reference
		if n.Type == doc.LOCAL {
			continue
		}
		// Check if it is a valid remote path or C.
		if n.ImportPath == "C" {
			continue
		} else if !doc.IsValidRemotePath(n.ImportPath) {
			// Invalid import path.
			log.Error("download", "Skipped invalid package: "+fmt.Sprintf("%s@%s:%s",
				n.ImportPath, n.Type, doc.CheckNodeValue(n.Value)))
			failConut++
			continue
		}

		// Valid import path.
		gopathDir := path.Join(installGopath, n.ImportPath)
		//RootPath is projectpath with certainn number set as VCS TYPE
		n.RootPath = doc.GetProjectPath(n.ImportPath)
		// installPath is the local  gopm repository path with certain VCS value
		installPath := path.Join(installRepoPath, n.RootPath) + versionSuffix(n.Value)

		if isSubpackage(n.RootPath, ".") {
			continue
		}

		// Indicates whether need to download package again.
		if n.IsFixed() && com.IsExist(installPath) {
			n.IsGetDepsOnly = true
		}

		if !ctx.Bool("update") {
			// Check if package has been downloaded.
			if (len(n.Value) == 0 && !ctx.Bool("remote") && com.IsExist(gopathDir)) ||
				com.IsExist(installPath) {
				log.Trace("Skipped installed package: %s@%s:%s",
					n.ImportPath, n.Type, doc.CheckNodeValue(n.Value))

				// Only copy when no version control.
				// if local set copy to local gopath
				if (ctx.Bool("gopath") || ctx.Bool("local")) && (com.IsExist(installPath) ||
					len(getVcsName(gopathDir)) == 0) {
					copyToGopath(installPath, gopathDir)
				}
				continue
			} else {
				doc.LocalNodes.SetValue(n.RootPath, "value", "")
			}
		}

		if downloadCache[n.RootPath] {
			log.Trace("Skipped downloaded package: %s@%s:%s",
				n.ImportPath, n.Type, doc.CheckNodeValue(n.Value))
			continue
		}

		// Download package.
		nod, imports := downloadPackage(ctx, n)
		if len(imports) > 0 {
			var gf *goconfig.ConfigFile

			// Check if has gopmfile.
			if com.IsFile(installPath + "/" + doc.GOPM_FILE_NAME) {
				log.Log("Found gopmfile: %s@%s:%s",
					n.ImportPath, n.Type, doc.CheckNodeValue(n.Value))
				gf = doc.NewGopmfile(installPath)
			}

			// Need to download dependencies.
			// Generate temporary nodes.
			nodes := make([]*doc.Node, len(imports))
			for i := range nodes {
				nodes[i] = doc.NewNode(imports[i], imports[i], doc.BRANCH, "", true)

				if gf == nil {
					continue
				}

				// Check if user specified the version.
				if v, err := gf.GetValue("deps", imports[i]); err == nil && len(v) > 0 {
					nodes[i].Type, nodes[i].Value = validPath(v)
				}
			}
			downloadPackages(ctx, nodes)
		}

		// Only save package information with specific commit.
		if nod == nil {
			continue
		}

		// Save record in local nodes.
		log.Success("SUCC", "GET", fmt.Sprintf("%s@%s:%s",
			n.ImportPath, n.Type, doc.CheckNodeValue(n.Value)))
		downloadCount++

		// Only save non-commit node.
		if len(nod.Value) == 0 && len(nod.Revision) > 0 {
			doc.LocalNodes.SetValue(nod.RootPath, "value", nod.Revision)
		}

		//if update set downloadPackage will use VSC tools to download the package
		//else just use puredownload to gopm repos and copy to gopath
		if (ctx.Bool("gopath") || ctx.Bool("local")) && com.IsExist(installPath) && !ctx.Bool("update") &&
			len(getVcsName(path.Join(installGopath, nod.RootPath))) == 0 {
			copyToGopath(installPath, gopathDir)
		}
	}
}

// downloadPackage downloads package either use version control tools or not.
func downloadPackage(ctx *cli.Context, nod *doc.Node) (*doc.Node, []string) {
	log.Message("Downloading", fmt.Sprintf("package: %s@%s:%s",
		nod.ImportPath, nod.Type, doc.CheckNodeValue(nod.Value)))
	// Mark as donwloaded.
	downloadCache[nod.RootPath] = true

	// Check if only need to use VCS tools.
	var imports []string
	var err error
	gopathDir := path.Join(installGopath, nod.RootPath)
	vcs := getVcsName(gopathDir)
	//if update set and gopath set and VCS tools set,
	//use VCS tools  to download the package
	if ctx.Bool("update") && (ctx.Bool("gopath") || ctx.Bool("local")) && len(vcs) > 0 {
		err = updateByVcs(vcs, gopathDir)
		imports = doc.GetAllImports([]string{gopathDir}, nod.RootPath, false, false)
	} else {
		// If package has revision and exist, then just check dependencies.
		if nod.IsGetDepsOnly {
			return nod, doc.GetAllImports([]string{path.Join(installRepoPath, nod.RootPath) + versionSuffix(nod.Value)},
				nod.RootPath, ctx.Bool("example"), false)
		}
		nod.Revision = doc.LocalNodes.MustValue(nod.RootPath, "value")
		imports, err = doc.PureDownload(nod, installRepoPath, ctx) //CmdGet.Flags)
	}

	if err != nil {
		log.Error("get", "Fail to download pakage: "+nod.ImportPath)
		log.Error("", "\t"+err.Error())
		failConut++
		os.RemoveAll(installRepoPath + "/" + nod.RootPath)
		return nil, nil
	}
	return nod, imports
}

//check whether dirPath has .git .hg .svn else return ""
func getVcsName(dirPath string) string {
	switch {
	case com.IsExist(path.Join(dirPath, ".git")):
		return "git"
	case com.IsExist(path.Join(dirPath, ".hg")):
		return "hg"
	case com.IsExist(path.Join(dirPath, ".svn")):
		return "svn"
	}
	return ""
}

//if vcs has been detected ,  use corresponding command to update dirPath
func updateByVcs(vcs, dirPath string) error {
	err := os.Chdir(dirPath)
	if err != nil {
		log.Error("Update by VCS", "Fail to change work directory:")
		log.Fatal("", "\t"+err.Error())
	}
	defer os.Chdir(workDir)

	switch vcs {
	case "git":
		branch, _, err := com.ExecCmd("git", "rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			log.Error("", "Error occurs when 'git rev-parse --abbrev-ref HEAD'")
			log.Error("", "\t"+err.Error())
		}

		_, _, err = com.ExecCmd("git", "pull", "origin", branch)
		if err != nil {
			log.Error("", "Error occurs when 'git pull origin "+branch+"'")
			log.Error("", "\t"+err.Error())
		}
	case "hg":
		_, stderr, err := com.ExecCmd("hg", "pull")
		if err != nil {
			log.Error("", "Error occurs when 'hg pull'")
			log.Error("", "\t"+err.Error())
		}
		if len(stderr) > 0 {
			log.Error("", "Error: "+stderr)
		}

		_, stderr, err = com.ExecCmd("hg", "up")
		if err != nil {
			log.Error("", "Error occurs when 'hg up'")
			log.Error("", "\t"+err.Error())
		}
		if len(stderr) > 0 {
			log.Error("", "Error: "+stderr)
		}
	case "svn":
		_, stderr, err := com.ExecCmd("svn", "update")
		if err != nil {
			log.Error("", "Error occurs when 'svn update'")
			log.Error("", "\t"+err.Error())
		}
		if len(stderr) > 0 {
			log.Error("", "Error: "+stderr)
		}
	}
	return nil
}
