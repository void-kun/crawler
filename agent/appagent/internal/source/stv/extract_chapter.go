package stv

import (
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/zrik/agent/appagent/pkg/spider"
)

func (s *Sangtacviet) ExtractChapter(url string, page *rod.Page, spider spider.TaskSpider) (any, error) {
	_, err := AsHeadSpider(spider)
	if err != nil {
		return nil, fmt.Errorf("spider is not of type *spider.HeadSpider")
	}

	paths := strings.Split(url, "/")
	bookId := paths[len(paths)-3]
	chapterId := paths[len(paths)-2]
	bookHost := paths[len(paths)-5]
	bookSty := paths[len(paths)-4]

	chapterUrl := fmt.Sprintf("%s/index.php?bookid=%s&c=%s&h=%s&ngmar=readc&sajax=readchapter&sty=%s&exts=", s.origin, bookId, chapterId, bookHost, bookSty)
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
		return result, nil
	}

	return nil, fmt.Errorf("failed to extract chapter: %s", result)
}
