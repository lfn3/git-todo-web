package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/codegangsta/martini"
	"github.com/libgit2/git2go"
)

type Config struct {
	GitPaths []string
}

func main() {
	file, err := os.Open("./config.json")
	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	file.Close()

	repos := make(map[string]*git.Repository, len(config.GitPaths))

	for _, repoPath := range config.GitPaths {
		repo, err := git.OpenRepository(repoPath)
		if err != nil {
			log.Fatal(err)
		}
		repos[filepath.Base(repoPath)] = repo
	}

	m := martini.Classic()

	m.Get("/", func(router martini.Routes, logger *log.Logger) string {
		output := "<h1> Todo files</h1>"
		for repoName, repo := range repos {
			output += "<h3> " + repoName + "</h3><ul>"

			branch, err := repo.LookupBranch("master", git.BranchLocal)
			if err != nil {
				logger.Println(err)
			}
			currentCommit, err := repo.LookupCommit(branch.Target())
			if err != nil {
				logger.Println(err)
			}
			currentTree, err := currentCommit.Tree()

			for i := uint64(0); i < currentTree.EntryCount(); i++ {
				treeEntry := currentTree.EntryByIndex(i)
				if err != nil {
					logger.Println(err)
				}

				output += "<li><a href=\"" + router.URLFor("file", repoName, treeEntry.Name) + "\"> " + treeEntry.Name + "</a></li>"
			}

			output += "</ul>"
		}

		return output
	}).Name("index")

	m.Get("/repo/:repoName/file/:fileName", func(params martini.Params, logger *log.Logger) string {
		repo := repos[params["repoName"]]

		branch, err := repo.LookupBranch("master", git.BranchLocal)
		if err != nil {
			logger.Println(err)
		}
		currentCommit, err := repo.LookupCommit(branch.Target())
		if err != nil {
			logger.Println(err)
		}
		currentTree, err := currentCommit.Tree()

		for i := uint64(0); i < currentTree.EntryCount(); i++ {
			treeEntry := currentTree.EntryByIndex(i)
			if err != nil {
				logger.Println(err)
			}

			if treeEntry.Name == params["fileName"] {
				blob, err := repo.LookupBlob(treeEntry.Id)
				if err != nil {
					logger.Println(err)
				}

				return string(blob.Contents())
 			}

		}
		
		return "File not found"

	}).Name("file")

	m.Run()
}