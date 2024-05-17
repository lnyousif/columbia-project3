package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

func translateText(targetLanguage, text string) (string, error) {
	ctx := context.Background()

	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", fmt.Errorf("language.Parse: %w", err)
	}

	client, err := translate.NewClient(ctx)
	if err != nil {
		fmt.Printf("err %s\n", err)
		return "", err
	}
	defer client.Close()

	resp, err := client.Translate(ctx, []string{text}, lang, nil)
	if err != nil {
		return "", fmt.Errorf("Translate: %w", err)
	}
	if len(resp) == 0 {
		return "", fmt.Errorf("Translate returned empty response to text: %s", text)
	}

	return resp[0].Text, nil
}

func get_title(content string, lang string) string {
	re_date := regexp.MustCompile(`date = (.+)`)
	re_title := regexp.MustCompile(`title = '([\w\s]+)'`)

	source_date := re_date.FindStringSubmatch(content)[1]
	source_title := re_title.FindStringSubmatch(content)[1]

	translation, err := translateText(lang, source_title)
	if err != nil {
		fmt.Printf("title error: %s", err)
	}

	return fmt.Sprintf("+++\ntitle = '%s'\ndate = %s\n+++", translation, source_date)
}

func get_body(content string, lang string) string {
	re_body := regexp.MustCompile(`\+{3}(?s).*\+{3}((?s).*)`)

	source_body := re_body.FindStringSubmatch(content)[1]

	paragraphs := strings.Split(source_body, "\n")
	translated_paragraphs := make([]string, len(paragraphs))

	for i, p := range paragraphs {
		re := regexp.MustCompile(`\(([^\)]+)\)`)
		hrefs := re.FindAllStringSubmatch(p, -1)

		p = re.ReplaceAllString(p, "()")

		translation, err := translateText(lang, p)
		if err != nil {
			fmt.Printf("title error: %s", err)
		} else if len(translation) > 0 {
			for _, ref := range hrefs {
				r := regexp.MustCompile(`\(\)`)
				translation = r.ReplaceAllString(translation, ref[0])
			}

			translated_paragraphs[i] = translation
		}
	}

	return strings.Join(translated_paragraphs, "\n")
}

func main() {
	base_filepath := "/Users/adellehousker/fun/ai/Columbia/project3/columbia-project3/Adelle/blog-main/content"
	source_filepath := fmt.Sprintf("%s/%s", base_filepath, "en/posts")
	source_filename := os.Args[1]

	content, err := os.ReadFile(fmt.Sprintf("%s/%s", source_filepath, source_filename))
	if err != nil {
		fmt.Printf("err reading file %s\n", err)
	}
	source_content := string(content)

	for _, lang := range []string{"fr", "es", "no", "ar"} {
		fmt.Printf("Translating %s\n", lang)
		title := get_title(source_content, lang)
		body := get_body(source_content, lang)

		post := fmt.Sprintf("%s%s", title, body)
		filepath := fmt.Sprintf("%s/%s/%s/%s", base_filepath, lang, "posts", source_filename)

		if err := os.WriteFile(filepath, []byte(post), 0666); err != nil {
			log.Fatal(err)
		}
	}
}
