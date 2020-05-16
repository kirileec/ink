package article

// Data struct
type ArticleInfo struct {
	DetailDate int64
	Date       string
	Title      string
	Link       string
	Top        bool
}

type Archive struct {
	Year     string
	Articles Collections
}

type Tag struct {
	Name     string
	Count    int
	Articles Collections
}
