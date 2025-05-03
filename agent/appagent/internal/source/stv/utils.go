package sangtacviet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
			titleParts := strings.Split(strings.Trim(parts[1], " "), " ")
			if len(titleParts) < 2 {
				continue
			}

			partIndex := 0
			_, err := strconv.Atoi(titleParts[partIndex])
			if err != nil {
				partIndex += 1
				_, _ = strconv.Atoi(titleParts[partIndex])
			}

			title := strings.TrimSpace(strings.Join(titleParts[partIndex+1:], " "))
			title = strings.Trim(title, " ")
			title = strings.TrimLeft(title, "Thứ")
			title = strings.TrimLeft(title, "thứ")
			title = strings.TrimLeft(title, "chương")
			title = strings.TrimLeft(title, "hồi")
			title = strings.Trim(title, " ")
			title = strings.Replace(title, "-/-unvip", "", -1)
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
	// Create a reader from the HTML content
	reader := strings.NewReader(htmlContent)

	// Parse the HTML
	doc, err := html.Parse(reader)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract text from p tags
	var result []string
	var extractText func(*html.Node)

	extractText = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			// Found a p tag, extract its text content
			var pText strings.Builder

			// Helper function to extract text from the p tag and its children
			var extractPTagText func(*html.Node)
			extractPTagText = func(node *html.Node) {
				if node.Type == html.TextNode {
					// Add text node content
					pText.WriteString(node.Data)
				} else if node.Type == html.ElementNode && node.Data == "i" {
					// For i tags, extract their text content
					for child := node.FirstChild; child != nil; child = child.NextSibling {
						if child.Type == html.TextNode {
							pText.WriteString(child.Data)
						} else {
							// Recursively process children of i tags
							extractPTagText(child)
						}
					}
				} else {
					// Process children for other node types
					for child := node.FirstChild; child != nil; child = child.NextSibling {
						extractPTagText(child)
					}
				}
			}

			// Start extracting text from the p tag
			extractPTagText(n)

			// Add the extracted text to the result if it's not empty
			if text := strings.TrimSpace(pText.String()); text != "" {
				fmt.Printf("Extracted text: %s\n", text)
				result = append(result, text)
			}
		} else {
			// Continue traversing the DOM
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				extractText(c)
			}
		}
	}

	// Start the extraction process
	extractText(doc)

	// Join all paragraphs with new line characters
	return strings.Join(result, "\n"), nil
}

// LoadCharacterMapping loads the character mapping from the specified JSON file
func LoadCharacterMapping(filePath string) (map[string]string, error) {
	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mapping file: %w", err)
	}

	// Parse the JSON data
	var mapping map[string]string
	if err := json.Unmarshal(data, &mapping); err != nil {
		return nil, fmt.Errorf("failed to parse mapping file: %w", err)
	}

	return mapping, nil
}

// MapCharacters replaces Unicode escape sequences with their corresponding characters
func MapCharacters(text string, mapping map[string]string) string {
	var builder strings.Builder
	for _, r := range text {
		ch := string(r)
		if replacement, ok := mapping[ch]; ok {
			builder.WriteString(fmt.Sprintf("[%s]", replacement))
		} else {
			builder.WriteString(ch)
		}
	}
	return builder.String()
}
