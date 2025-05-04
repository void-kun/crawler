package sangtacviet

import (
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/zrik/agent/appagent/pkg/spider"
)

func (s *Sangtacviet) ExtractChapter(url string, page *rod.Page, hs *spider.HeadSpider) error {
	if strings.HasPrefix(url, hs.SessionPrefix) {
		return nil
	}

	paths := strings.Split(url, "/")
	fmt.Println("\n==================================================================================")
	fmt.Printf("======= Extracting chapter: valid path length %d, current path length %d\n", CHAPTER_URL_LENGTH, len(paths))
	if len(paths) != CHAPTER_URL_LENGTH {
		return nil
	}

	bookId := paths[len(paths)-3]
	chapterId := paths[len(paths)-2]
	bookHost := paths[len(paths)-5]
	bookSty := paths[len(paths)-4]

	chapterUrl := "https://sangtacviet.app/index.php?bookid=%s&c=%s&h=%s&ngmar=readc&sajax=readchapter&sty=%s&exts="
	chapterUrl = fmt.Sprintf(chapterUrl, bookId, chapterId, bookHost, bookSty)
	fmt.Println(chapterUrl)

	// Extract chapters
	result := page.MustEval(`
		async (url) => {
		  function chapterApi(url) {
				return new Promise((resolve, reject) => {
					const xhr = new XMLHttpRequest();
	  			xhr.open("POST", url, true);
					xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');

					xhr.onload  = function () {
						if (xhr.status == 200) {
							if (xhr.responseText == "") {
								resolve("error: response text empty");
							}

							let jsonVal = JSON.parse(xhr.responseText);
							try {
								if (jsonVal.code == 0) {
									if (!jsonVal.data) {
										resolve("error: data is empty");
										return;
									}
									resolve(jsonVal.data);
								} else {
									 resolve("error: code is " + jsonVal.code);
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

			const data = await chapterApi(url);
			return data;
		}
		`, chapterUrl).String()

	if !strings.HasPrefix(result, "error:") {
		result, _ = ExtractTextFromHTML(result)

		SaveTextToFile(result, fmt.Sprintf("%s_%s", bookId, chapterId), "txt")
		fmt.Println("====== Chapter extracted")
		return nil
	}

	return fmt.Errorf("failed to extract chapter: %s", result)
}
