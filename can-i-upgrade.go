package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func getURLs(file string, oldDist string, newDist string) (map[string]bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "deb") {
			parts := strings.Split(line, " ")
			for i, p := range parts {
				if strings.HasPrefix(p, "http") {
					if len(parts) > i+1 {
						dist := parts[i+1]
						prefix := strings.TrimSuffix(p, "/")
						if dist == "/" {
							out[p+"/Release"] = true
							continue
						} else if dist == oldDist {
							dist = newDist
						}
						url := fmt.Sprintf("%s/dists/%s/Release", prefix, dist)
						out[url] = true
					}
				}
			}
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s current_release new_release", os.Args[0])
	}
	oldRelease, newRelease := os.Args[1], os.Args[2]
	urls := make([]string, 0)
	err := filepath.Walk("/etc/apt/sources.list.d", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("could not read %s: %w", path, err)
		}
		if info.IsDir() {
			return nil
		}
		urlsFromFile, err := getURLs(path, oldRelease, newRelease)
		if err != nil {
			return fmt.Errorf("could not get urls from %s: %w", path, err)
		}
		if len(urlsFromFile) == 0 {
			return nil
		}
		for url := range urlsFromFile {
			urls = append(urls, url)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, url := range urls {
		res, err := http.Get(url)
		if err != nil {
			log.Printf("Could not fetch from %s: %v", url, err)
		}
		if res.StatusCode >= 400 {
			fmt.Printf("%s is not out yet!\n", url)
		}
	}
}
