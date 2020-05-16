package article

// For sort
type Collections []interface{}

func (v Collections) Len() int      { return len(v) }
func (v Collections) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v Collections) Less(i, j int) bool {
	switch v[i].(type) {
	case ArticleInfo:
		return v[i].(ArticleInfo).DetailDate > v[j].(ArticleInfo).DetailDate
	case Article:
		article1 := v[i].(Article)
		article2 := v[j].(Article)
		if article1.Top && !article2.Top {
			return true
		} else if !article1.Top && article2.Top {
			return false
		} else {
			return article1.Date > article2.Date
		}
	case Archive:
		return v[i].(Archive).Year > v[j].(Archive).Year
	case Tag:
		if v[i].(Tag).Count == v[j].(Tag).Count {
			return v[i].(Tag).Name > v[j].(Tag).Name
		}
		return v[i].(Tag).Count > v[j].(Tag).Count
	}
	return false
}
