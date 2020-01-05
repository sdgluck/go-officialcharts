package go_officialcharts

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCharts(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "happy: gets chart for Halloween 1992",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetCharts(31, 10, 1992)
			if err != nil {
				t.Fatal(err)
				return
			}

			a := assert.New(t)

			a.Len(result.Songs, 75)
			a.Equal("END OF THE ROAD", result.Songs[0].Title)
			a.Equal("ROADHOUSE MEDLEY (ANNIVERSARY WALTZ PART 25)", result.Songs[74].Title)
		})
	}
}