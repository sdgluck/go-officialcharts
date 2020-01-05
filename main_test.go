package go_officialcharts

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCharts(t *testing.T) {
	type args struct {
		year int
		month int
		day int
	}

	tests := []struct {
		name string
		args args
		expectLength int
		expectFirstSongTitle string
		expectLastSongTitle string
	}{
		{
			name: "happy: gets chart for Halloween 1992",
			args: args{
				year: 1992,
				month: 10,
				day: 31,
			},
			expectLength: 75,
			expectFirstSongTitle: "END OF THE ROAD",
			expectLastSongTitle: "ROADHOUSE MEDLEY (ANNIVERSARY WALTZ PART 25)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetCharts(tt.args.day, tt.args.month, tt.args.year)
			if err != nil {
				t.Fatal(err)
				return
			}

			a := assert.New(t)

			a.Len(result.Songs, tt.expectLength)
			a.Equal(tt.expectFirstSongTitle, result.Songs[0].Title)
			a.Equal(tt.expectLastSongTitle, result.Songs[tt.expectLength - 1].Title)
		})
	}
}