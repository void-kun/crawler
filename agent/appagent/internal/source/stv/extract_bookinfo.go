package stv

import (
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/zrik/agent/appagent/pkg/spider"
)

func (s *Sangtacviet) ExtractBookInfo(url string, page *rod.Page, spider spider.TaskSpider) (any, error) {
	_, err := AsHeadSpider(spider)
	if err != nil {
		return nil, fmt.Errorf("spider is not of type *spider.HeadSpider")
	}

	defer page.MustClose()

	paths := strings.Split(url, "/")
	bookInfo, err := ExtractBookInfoFromElement(page)
	if err != nil {
		return nil, fmt.Errorf("failed to extract book info: %w", err)
	}

	bookInfo.BookUrl = url
	bookInfo.BookId = paths[len(paths)-2]
	bookInfo.BookHost = paths[len(paths)-4]

	chapterListUrl := fmt.Sprintf("%s/index.php?ngmar=chapterlist&h=%s&bookid=%s&sajax=getchapterlist", s.origin, bookInfo.BookHost, bookInfo.BookId)
	// Extract chapters
	result := page.MustEval(`
		async (url) => {
		  function chapterListApi(url) {
				return new Promise((resolve, reject) => {
					// Set cookies before sending the request
					const hstamp = Math.floor(Date.now() / 1000);
					document.cookie = 'hstamp=' + hstamp + '; path=/';
					document.cookie = 'lang=vi; path=/';

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

			const data = await chapterListApi(url);
			return data;
		}
		`, chapterListUrl).String()

	if strings.HasPrefix(result, "error:") {
		return nil, fmt.Errorf("failed to extract book info: %s", result)
	}

	chapters, err := ExtractChapterInfoFromData(result, bookInfo.BookUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to extract book info: %+v", err)
	}

	bookInfo.Chapters = chapters
	bookInfoByte, err := ConvertToRawMessage(bookInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal book info: %+v", err)
	}

	return bookInfoByte, nil
}
