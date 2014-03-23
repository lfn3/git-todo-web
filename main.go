package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/codegangsta/martini"
	"github.com/libgit2/git2go"
	"github.com/martini-contrib/render"
)

type Config struct {
	GitPaths []string
}

type TemplateModel struct {
	Route martini.Routes
}

type RenderableRepoList struct {
	TemplateModel
	Repos []RenderableRepo
}

type RenderableRepo struct {
	RepoName string
	Files    []TrackedFile
}

type TrackedFile struct {
	FileName string
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

	m.Use(render.Renderer())

	m.Get("/", func(router martini.Routes, logger *log.Logger, r render.Render) {
		model := RenderableRepoList{
			Repos: make([]RenderableRepo, len(repos)),
			TemplateModel: TemplateModel{ Route: router },
		}
		for repoName, repo := range repos {
			rRepo := RenderableRepo{RepoName: repoName}

			branch, err := repo.LookupBranch("master", git.BranchLocal)
			if err != nil {
				logger.Println(err)
			}
			currentCommit, err := repo.LookupCommit(branch.Target())
			if err != nil {
				logger.Println(err)
			}
			currentTree, err := currentCommit.Tree()
			rRepo.Files = make([]TrackedFile, currentTree.EntryCount())
			for i := uint64(0); i < currentTree.EntryCount(); i++ {
				treeEntry := currentTree.EntryByIndex(i)
				if err != nil {
					logger.Println(err)
				}

				rRepo.Files = append(rRepo.Files, TrackedFile{FileName: treeEntry.Name})
			}

			model.Repos = append(model.Repos, rRepo)
		}

		r.HTML(200, "fileList", model)
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
