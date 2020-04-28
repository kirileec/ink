package main

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	a := "2015-04-02T13:57:35Z08:00"
	ti, err := time.ParseInLocation(DATE_FORMAT_WITH_TIMEZONE, a, time.Now().Location())
	t.Error(err)
	t.Log(ti)
}
