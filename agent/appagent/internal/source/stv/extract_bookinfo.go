package sangtacviet

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-rod/rod"
)

const BOOK_INFO_URL_LENGTH = 8

func (s *Sangtacviet) ExtractBookInfo(url string, page *rod.Page) error {
	paths := strings.Split(url, "/")
	fmt.Printf("Extracting book info: valid path length %d, current path length %d\n", BOOK_INFO_URL_LENGTH, len(paths))
	if len(paths) != BOOK_INFO_URL_LENGTH {
		return nil
	}

	bookInfo, err := ExtractBookInfoFromElement(page)
	if err != nil {
		return fmt.Errorf("failed to extract book info: %w", err)
	}

	bookInfo.BookUrl = url
	bookInfo.BookId = paths[len(paths)-2]
	bookInfo.BookHost = paths[len(paths)-4]

	chapterListUrl := "https://sangtacviet.app/index.php?ngmar=chapterlist&h=%s&bookid=%s&sajax=getchapterlist"
	chapterListUrl = fmt.Sprintf(chapterListUrl, bookInfo.BookHost, bookInfo.BookId)
	fmt.Println(chapterListUrl)
	// Extract chapters
	result := page.MustEval(`
		async (url) => {
		  function chapterListApi(url) {
				return new Promise((resolve, reject) => {
					const xhr = new XMLHttpRequest();
    			xhr.open("GET", url, true);

					xhr.onload  = function () {
						if (xhr.status == 200) {
							if (xhr.responseText == "") {
								resolve("error: response text empty");
								return;
							}

							let jsonVal = JSON.parse(xhr.responseText);
							try {
								if (jsonVal.code == 1) {
									if (jsonVal.enckey) {
										eval(atob(jsonVal.enckey));
									}
									if (!jsonVal.data) {
										resolve("error: data is empty");
										return;
									}
									resolve(jsonVal.data);
								} else {
									 resolve("error: code not 1");
								}
							} catch(err) {
								resolve("error: " + err);
							}
						} else {
							resolve("error: status not 200"); 
						}
					}

					xhr.send();
				})
			}

			const data = await chapterListApi(url);
			return data;
		}
		`, chapterListUrl).String()

	if strings.HasPrefix(result, "error:") {
		return fmt.Errorf("failed to extract chapters")
	}
	SaveTextToFile(result, bookInfo.BookName, "txt")

	chapters, err := ExtractChapterInfoFromData(result, bookInfo.BookUrl)
	if err != nil {
		return fmt.Errorf("failed to extract chapters: %w", err)
	}

	bookInfo.Chapters = chapters
	bookInfoByte, err := json.Marshal(bookInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal book info: %w", err)
	}
	SaveTextToFile(string(bookInfoByte), bookInfo.BookName, "json")

	return nil
}
