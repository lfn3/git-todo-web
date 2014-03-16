package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"io/ioutil"

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

			index, err := repo.Index()
			if err != nil {
				logger.Println(err)
			}

			for i := uint(0); i < index.EntryCount(); i++ {
				indexEntry, err := index.EntryByIndex(i)
				if err != nil {
					logger.Println(err)
				}

				output += "<li><a href=\"" + router.URLFor("file", repoName, indexEntry.Path) + "\"> " + indexEntry.Path + "</a></li>"
			}

			output += "</ul>"
		}

		return output
	}).Name("index")

	m.Get("/repo/:repoName/file/:fileName", func(params martini.Params, logger *log.Logger) string {
		repo := repos[params["repoName"]]
		//Might be faster to do some sort of filesystem based search?

		index, err := repo.Index()
		if err != nil {
			logger.Println(err)
		}

		for i := uint(0); i < index.EntryCount(); i++ {
			indexEntry, err := index.EntryByIndex(i)
			if err != nil {
				logger.Println(err)
			}

			if indexEntry.Path == params["fileName"] {
				fileBytes, err := ioutil.ReadFile(filepath.Join(repo.Path(), "..", indexEntry.Path))
				if err != nil {
					logger.Println(err)
				}

				return string(fileBytes)
 			}

		}
		
		return "File not found"

	}).Name("file")

	m.Run()
}