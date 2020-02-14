package go_officialcharts

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const officialChartsUrlTmpl = "https://www.officialcharts.com/charts/singles-chart/%d%d%d/"

const (
	PositionMoveUp   PositionMove = 1
	PositionMoveDown PositionMove = 2
	PositionMoveNew  PositionMove = 3
)

type PositionMove int

type Song struct {
	Artist             string
	Title              string
	PeakPosition       int
	Position           int
	PositionLastWeek   int
	PositionMoved      PositionMove
	PositionReentry    bool
	WeeksOnChart       int
	RecordLabel        string
	CoverImageSmallURL string
	CoverImageLargeURL string
}

type Chart struct {
	Date  time.Time
	Songs []*Song
}

func isSongRow(e *goquery.Selection) bool {
	ad := e.Find(".adspace")
	if ad.Length() > 0 {
		return false
	}
	className, _ := e.Attr("class")
	return className == ""
}

func processSongRow(e *goquery.Selection) (*Song, error) {
	var isReentry bool

	pos, err := strconv.Atoi(e.Find(".position").Text())
	if err != nil {
		return nil, errors.Wrap(err, "converting position string to integer")
	}

	var posLastWeek int
	posLastWeekStr := strings.TrimSpace(e.Find(".last-week").Text())
	if posLastWeekStr == "New" || posLastWeekStr == "Re" {
		isReentry = posLastWeekStr == "Re"
		posLastWeek = -1
	} else {
		posLastWeek, err = strconv.Atoi(posLastWeekStr)
		if err != nil {
			return nil, errors.Wrap(err, "converting position last week to integer")
		}
	}

	peakPos, err := strconv.Atoi(e.Find("td:nth-child(4)").Text())
	if err != nil {

		return nil, errors.Wrap(err, "converting peak position string to integer")
	}

	weeksOnChart, err := strconv.Atoi(e.Find("td:nth-child(5)").Text())
	if err != nil {
		return nil, errors.Wrap(err, "converting weeks on chart string to integer")
	}

	var posMoved PositionMove
	attr, _ := e.Find(".last-week").Attr("class")
	if strings.Contains(attr, "icon-up") {
		posMoved = PositionMoveUp
	} else if strings.Contains(attr, "icon-down") {
		posMoved = PositionMoveDown
	} else {
		posMoved = PositionMoveNew
	}

	coverImageURL, _ := e.Find(".track .cover img").Attr("src")

	return &Song{
		Artist:             e.Find(".title-artist .artist a").Text(),
		Title:              e.Find(".title-artist .title a").Text(),
		PeakPosition:       peakPos,
		Position:           pos,
		PositionLastWeek:   posLastWeek,
		PositionMoved:      posMoved,
		PositionReentry:    isReentry,
		WeeksOnChart:       weeksOnChart,
		RecordLabel:        e.Find(".label").Text(),
		CoverImageSmallURL: coverImageURL,
		CoverImageLargeURL: strings.Replace(coverImageURL, "img/small?", "img/large?", 1),
	}, nil
}

// GetCharts scrapes the officalcharts.com singles chart for the given date.
func GetCharts(day, month, year int) (*Chart, error) {
	if day < 1 || day > 31 {
		return nil, fmt.Errorf("invalid day, expecting value between 1-31 inclusive, got %d", day)
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid month, expecting value between 1-12 inclusive, got %d", month)
	}
	if year < 1952 || year > time.Now().Year() {
		return nil, fmt.Errorf("invalid year, expecting value between 1952 and current year, got %d", year)
	}

	chart := &Chart{
		Date: time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local),
	}

	eg := errgroup.Group{}

	c := colly.NewCollector()

	tBodySelector := "section.chart .chart-positions tbody"

	c.OnHTML(tBodySelector, func(e *colly.HTMLElement) {
		var size int
		e.DOM.Find("tr").Each(func(i int, e *goquery.Selection) {
			if isSongRow(e) {
				size += 1
			}
		})
		chart.Songs = make([]*Song, size)
		c.OnHTMLDetach(tBodySelector)
	})

	c.OnHTML("section.chart .chart-positions tr", func(e *colly.HTMLElement) {
		if chart.Songs == nil {
			eg.Go(func() error {
				return errors.New("songs slice uninitialised")
			})
			return
		}
		if !isSongRow(e.DOM) {
			return
		}
		eg.Go(func() error {
			song, err := processSongRow(e.DOM)
			if song == nil || err != nil {
				return errors.Wrap(err, "processing song")
			}
			chart.Songs[song.Position-1] = song
			return nil
		})
	})

	if err := c.Visit(fmt.Sprintf(officialChartsUrlTmpl, year, month, day)); err != nil {
		return nil, errors.Wrap(err, "visiting officialcharts.com")
	}

	if err := eg.Wait(); err != nil {
		return nil, errors.Wrap(err, "getting charts")
	}

	return chart, nil
}
