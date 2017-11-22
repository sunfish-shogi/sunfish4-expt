package shogiserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalRate(t *testing.T) {
	rate, err := UnmarshalRate([]byte(`---` + "\n" +
		`players:` + "\n" +
		`  1:` + "\n" +
		`    player_a:` + "\n" +
		`      name: player_a` + "\n" +
		`      rating_group: 1` + "\n" +
		`      rate: 1200.0` + "\n" +
		`      last_modified: 2017-06-13` + "\n" +
		`      win: 100.0` + "\n" +
		`      loss: 50.0` + "\n" +
		`    player_b:` + "\n" +
		`      name: player_b` + "\n" +
		`      rating_group: 1` + "\n" +
		`      rate: 1000.0` + "\n" +
		`      last_modified: 2017-06-12` + "\n" +
		`      win: 100.0` + "\n" +
		`      loss: 90.0` + "\n" +
		`    player_c:` + "\n" +
		`      name: player_c` + "\n" +
		`      rating_group: 1` + "\n" +
		`      rate: 900.0` + "\n" +
		`      last_modified: 2017-06-17` + "\n" +
		`      win: 40.0` + "\n" +
		`      loss: 90.0` + "\n" +
		`  999: {}` + "\n"))

	require.NoError(t, err)

	require.Len(t, rate.Players, 2)
	require.Len(t, rate.Players[1], 3)

	require.NotNil(t, rate.Players[1]["player_a"])
	require.Equal(t, "player_a", rate.Players[1]["player_a"].Name)
	require.Equal(t, 1, rate.Players[1]["player_a"].RatingGroup)
	require.Equal(t, float64(1200), rate.Players[1]["player_a"].Rate)
	require.Equal(t, "2017-06-13", rate.Players[1]["player_a"].LastModified)
	require.Equal(t, float64(100), rate.Players[1]["player_a"].Win)
	require.Equal(t, float64(50), rate.Players[1]["player_a"].Loss)

	assert.NotNil(t, rate.Players[999])
	assert.Empty(t, rate.Players[999])
}
