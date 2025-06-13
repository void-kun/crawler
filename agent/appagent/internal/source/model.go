package source

type Book struct {
	BookUrl      string
	BookId       string
	BookName     string
	BookImageUrl string
	AuthorName   string
	Chapters     []Chapter
	BookHost     string
}

type Chapter struct {
	ChapterId     string
	ChapterName   string
	ChapterUrl    string
	ChapterNumber int
}
