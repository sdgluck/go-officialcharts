package go_officialcharts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCharts(t *testing.T) {
	type args struct {
		year  int
		month int
		day   int
	}

	tests := []struct {
		name                 string
		args                 args
		expectError          string
		expectLength         int
		expectFirstSongTitle string
		expectLastSongTitle  string
	}{
		{
			name: "sad: bad year",
			args: args{
				year:  1950,
				month: 1,
				day:   1,
			},
			expectError: "invalid year, expecting value between 1952 and current year, got 1950",
		},
		{
			name: "sad: bad month",
			args: args{
				year:  1950,
				month: 100,
				day:   1,
			},
			expectError: "invalid month, expecting value between 1-12 inclusive, got 100",
		},
		{
			name: "sad: bad day",
			args: args{
				year:  1952,
				month: 1,
				day:   50,
			},
			expectError: "invalid day, expecting value between 1-31 inclusive, got 50",
		},
		{
			name: "happy: gets chart for Halloween 1992",
			args: args{
				year:  1992,
				month: 10,
				day:   31,
			},
			expectLength:         75,
			expectFirstSongTitle: "END OF THE ROAD",
			expectLastSongTitle:  "ROADHOUSE MEDLEY (ANNIVERSARY WALTZ PART 25)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)

			result, err := GetCharts(tt.args.day, tt.args.month, tt.args.year)

			if tt.expectError != "" {
				a.EqualError(err, tt.expectError)
				return
			} else if err != nil {
				t.Fatal(err)
				return
			}

			a.Len(result.Songs, tt.expectLength)
			a.Equal(tt.expectFirstSongTitle, result.Songs[0].Title)
			a.Equal(tt.expectLastSongTitle, result.Songs[tt.expectLength-1].Title)
		})
	}
}
