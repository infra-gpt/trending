package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

func gitAddCommitPush(date, filename string) {
	cmdGitAdd := exec.Command("git", "add", filename)
	cmdGitCommit := exec.Command("git", "commit", "-m", date)
	cmdGitPush := exec.Command("git", "push", "-u", "origin", "master")

	err := cmdGitAdd.Run()
	if err != nil {
		log.Fatalf("git add failed: %v", err)
	}

	err = cmdGitCommit.Run()
	if err != nil {
		log.Fatalf("git commit failed: %v", err)
	}

	err = cmdGitPush.Run()
	if err != nil {
		log.Fatalf("git push failed: %v", err)
	}
}

func createMarkdown(date, filename string) {
	content := fmt.Sprintf("## %s Github Trending\n", date)
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		log.Fatalf("failed to create markdown file: %v", err)
	}
}

func scrape(language, filename string, topk int) {
	client := resty.New()
	url := fmt.Sprintf("https://github.com/trending/%s", language)
	resp, err := client.R().
		SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.7; rv:11.0) Gecko/20100101 Firefox/11.0").
		SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8").
		SetHeader("Accept-Encoding", "gzip,deflate,sdch").
		SetHeader("Accept-Language", "zh-CN,zh;q=0.8").
		Get(url)

	if err != nil {
		log.Fatalf("failed to get trending repos: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		log.Fatalf("failed to parse HTML: %v", err)
	}

	var ds [][]string
	doc.Find("div.Box article.Box-row").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".lh-condensed a").Text()
		description := s.Find("p.col-9").Text()
		url, _ := s.Find(".lh-condensed a").Attr("href")
		url = "https://github.com" + url
		starFork := strings.Fields(s.Find(".f6 a").Text())
		star, fork := starFork[0], starFork[1]
		newStar := strings.Fields(s.Find(".f6 svg.octicon-star").Parent().Text())[1]

		ds = append(ds, []string{title, url, description, star, fork, newStar})
	})

	saveToMd(ds, filename, language, topk)
}

func saveToMd(ds [][]string, filename, language string, topk int) {
	content := fmt.Sprintf("\n### %s\n", language)
	for _, repo := range ds[:topk] {
		title, url, description, star, fork, newStar := repo[0], repo[1], repo[2], repo[3], repo[4], repo[5]
		newTitle := strings.ReplaceAll(strings.TrimSpace(title), "\n", "")
		afterTitle := strings.ReplaceAll(strings.TrimSpace(newTitle), " ", "")
		out := fmt.Sprintf("* [%s](%s): %s ***Star:%s Fork:%s Today stars:%s***\n", strings.TrimSpace(afterTitle) , strings.TrimSpace(url), strings.ReplaceAll(strings.TrimSpace(description), "\n", ""), strings.TrimSpace(star), strings.TrimSpace(fork), strings.TrimSpace(newStar))
		content += out
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open markdown file: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		log.Fatalf("failed to write to markdown file: %v", err)
	}
}

func main() {

	todayStr := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("markdowns/%s.md", todayStr)

	createMarkdown(todayStr, filename)

	scrape("", filename, 10)
	scrape("python", filename, 5)
	scrape("java", filename, 5)
	scrape("javascript", filename, 5)
	scrape("go", filename, 5)

	fmt.Printf("save markdown file to %s\n", filename)
}
