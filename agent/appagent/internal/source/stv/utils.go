package sangtacviet

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-rod/rod"
	"github.com/zrik/agent/appagent/internal/source"
	"golang.org/x/net/html"
)

func SaveTextToFile(text string, filename, ext string) error {
	// Create output directory if it doesn't exist
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filePath := filepath.Join(outputDir, fmt.Sprintf("%s.%s", filename, ext))

	// Write the content to the file
	if err := os.WriteFile(filePath, []byte(text), 0o644); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

func SaveHTMLToFile(chapterContainer *rod.Element, filename, ext string) error {
	// Get the HTML content of the div#chaptercontainer
	htmlContent, err := chapterContainer.HTML()
	if err != nil {
		return fmt.Errorf("failed to get HTML content: %w", err)
	}

	SaveTextToFile(htmlContent, filename, ext)
	fmt.Printf("HTML content from div#chaptercontainer saved")
	return nil
}

func ExtractBookInfoFromElement(element *rod.Page) (*source.Book, error) {
	book := &source.Book{}

	// Extract the book name
	bookNameElem, err := element.Element("#book_name2")
	if err != nil {
		return nil, fmt.Errorf("failed to find book name element: %w", err)
	}
	bookName, err := bookNameElem.Text()
	if err != nil {
		return nil, fmt.Errorf("failed to get book name text: %w", err)
	}

	// Extract the author name
	authorElem, err := element.Element("i.cap h2")
	if err != nil {
		return nil, fmt.Errorf("failed to find author element: %w", err)
	}
	authorName, err := authorElem.Text()
	if err != nil {
		return nil, fmt.Errorf("failed to get author name text: %w", err)
	}

	// Extract the avatar URL
	avatarElem, err := element.Element("img#thumb-prop")
	if err != nil {
		return nil, fmt.Errorf("failed to find avatar element: %w", err)
	}
	avatarURL, err := avatarElem.Attribute("src")
	if err != nil || avatarURL == nil {
		return nil, fmt.Errorf("failed to get avatar URL: %w", err)
	}

	book.BookName = bookName
	book.AuthorName = authorName
	book.BookImageUrl = *avatarURL
	return book, nil
}

func ExtractChapterInfoFromData(data, bookUrl string) ([]source.Chapter, error) {
	var chapters []source.Chapter

	// Remove start key
	data = strings.Replace(data, "1-/-", "", 1)
	dataSplit := strings.Split(data, "-//-1-/-")

	for _, chapterData := range dataSplit {
		if chapterData == "" {
			continue
		}

		chapter := source.Chapter{}
		parts := strings.Split(chapterData, "-/-")

		if len(parts) >= 2 {
			// Extract ChapterId
			idParts := strings.Split(parts[0], "-/-")
			chapter.ChapterId = idParts[0]

			// Extract ChapterName and ChapterIndex
			title := strings.Trim(parts[1], " ")
			chapter.ChapterName = title

			chapter.ChapterUrl = fmt.Sprintf("%s%s", bookUrl, chapter.ChapterId)
			chapters = append(chapters, chapter)
		}
	}

	return chapters, nil
}

func IsStartWithUpper(val string) bool {
	if len(val) == 0 {
		return false
	}
	return val[0] >= 'A' && val[0] <= 'Z'
}

// ExtractTextFromHTML extracts text content from HTML string, focusing on p tags.
// All text in p tags will be inlined, and each p tag will add a new line character.
func ExtractTextFromHTML(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	var extract func(*html.Node)

	extract = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(strings.Trim(n.Data, "\t"))
		}

		if n.Type == html.ElementNode {
			if n.Data == "p" && n.FirstChild != nil {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					extract(c)
				}
				buf.WriteString("\n") // newline sau mỗi thẻ <p>
				return
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(doc)
	result := strings.ReplaceAll(buf.String(), "\n\n", "\n")
	return result, nil
}
