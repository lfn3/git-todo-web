package main

import (
	"encoding/json"
	"log"
	"os"

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

	m := martini.Classic()

	m.Get("/", func() string {
		output := "<h1> Todo files</h1>"
		for _, repoPath := range config.GitPaths {
			output += "<ul>"

			repo, err := git.OpenRepository(repoPath)
			if err != nil {
				log.Fatal(err)
			}

			index, err := repo.Index()
			if err != nil {
				log.Fatal(err)
			}

			for i := uint(0); i < index.EntryCount(); i++ {
				indexEntry, err := index.EntryByIndex(i)
				if err != nil {
					log.Fatal(err)
				}

				output += "<li>" + indexEntry.Path + "</li>"
			}

			output += "</ul>"
		}

		return output
	})

	m.Run()
}