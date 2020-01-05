package go_officialcharts

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"strconv"
	"strings"
	"time"
)

const officialChartsUrlTmpl = "https://www.officialcharts.com/charts/singles-chart/%d%d%d/"

const (
	PositionMoveUp   PositionMove = 1
	PositionMoveDown PositionMove = 2
	PositionMoveNew  PositionMove = 3
)

type PositionMove int

type Song struct {
	Artist           string
	Title            string
	Position         int
	PositionLastWeek int
	PositionMoved    PositionMove
	WeeksOnChart     int
	RecordLabel      string
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

func processSong(e *goquery.Selection) (*Song, error) {
	pos, err := strconv.Atoi(e.Find(".position").Text())
	if err != nil {
		return nil, errors.Wrap(err, "casting position string to integer")
	}

	var posLastWeek int
	posLastWeekStr := strings.TrimSpace(e.Find(".last-week").Text())
	if posLastWeekStr == "New" {
		posLastWeek = -1
	} else {
		posLastWeek, err = strconv.Atoi(posLastWeekStr)
		if err != nil {
			return nil, errors.Wrap(err, "casting position last week to integer")
		}
	}

	weeksOnChart, err := strconv.Atoi(e.Find("td:nth-child(5)").Text())
	if err != nil {
		return nil, errors.Wrap(err, "casting position string to integer")
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

	return &Song{
		Artist:           e.Find(".title-artist .artist a").Text(),
		Title:            e.Find(".title-artist .title a").Text(),
		Position:         pos,
		PositionLastWeek: posLastWeek,
		PositionMoved:    posMoved,
		WeeksOnChart:     weeksOnChart,
		RecordLabel:      e.Find(".label").Text(),
	}, nil
}

// GetCharts scrapes the officalcharts.com singles chart for the given date.
func GetCharts(day, month, year int) (*Chart, error) {
	if day == 0 || day > 31 {
		return nil, fmt.Errorf("invalid day, expecting value between 1-31 inclusive, got %d", day)
	}
	if month == 0 || month > 12 {
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

	initialised := false

	c.OnHTML("section.chart .chart-positions tbody", func(e *colly.HTMLElement) {
		if initialised {
			return
		}
		initialised = true
		var size int
		e.DOM.Find("tr").Each(func(i int, e *goquery.Selection) {
			if isSongRow(e) {
				size += 1
			}
		})
		chart.Songs = make([]*Song, size)
	})

	c.OnHTML("section.chart .chart-positions tr", func(e *colly.HTMLElement) {
		if !isSongRow(e.DOM) {
			return
		}
		eg.Go(func() error {
			song, err := processSong(e.DOM)
			if song == nil || err != nil {
				return errors.Wrap(err, "failed processing song")
			}
			chart.Songs[song.Position-1] = song
			return nil
		})
	})

	if err := c.Visit(fmt.Sprintf(officialChartsUrlTmpl, year, month, day)); err != nil {
		return nil, errors.Wrap(err, "failed visiting officialcharts.com")
	}

	if err := eg.Wait(); err != nil {
		return nil, errors.Wrap(err, "failed getting charts")
	}

	return chart, nil
}
