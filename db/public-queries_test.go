package db_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func testGetThisMovieFromDB(t *testing.T) {
	c := newContext()
	mov, err := testDB.GetThisMovieFromDB(c, "tt3501632")
	assert.Nil(t, err)
	tt3501632.MovieID = mov.MovieID
	assert.Equal(t, tt3501632, mov)
}

func testGetThisSeriesFromDB(t *testing.T) {
	c := newContext()
	series, seasons, err := testDB.GetThisSeriesFromDB(c, "tt16350094")
	assert.Nil(t, err)
	tt16350094.SerieID = series.SerieID
	assert.Equal(t, tt16350094, series)
	for _, s := range *seasons {
		assert.Equal(t, tt16350094.IMDBid, s.IMDBid)
	}

	_, _, err = testDB.GetThisSeriesFromDB(c, "tt1635333330094")
	assert.NotNil(t, err)
}

func testGetEpisodesForaSeasonFromDB(t *testing.T) {
	c := newContext()
	episodes, err := testDB.GetEpisodesForaSeasonFromDB(c, "tt16350094", "1")
	assert.Nil(t, err)
	assert.True(t, len(*episodes) > 0)

	//invalid season number
	_, err = testDB.GetEpisodesForaSeasonFromDB(c, "tt16350094", "1e")
	assert.NotNil(t, err)

	//non-existed series
	_, err = testDB.GetEpisodesForaSeasonFromDB(c, "tt16333333350094", "1")
	assert.NotNil(t, err)

}

func testGetMoviesListWithPage(t *testing.T) {
	c := newContext()
	for _, s := range movie_list {
		testDB.AddMovieContentWithID(s)
	}
	movies, err := testDB.GetMoviesListWithPage(c, 1, 5)
	assert.Nil(t, err)
	assert.True(t, len(*movies) == 5)
}

func testGetSeriesListWithPage(t *testing.T) {
	c := newContext()
	for _, s := range series_list {
		testDB.AddSeriesContentWithID(s)
	}

	series, err := testDB.GetSeriesListWithPage(c, 1, 3)
	assert.Nil(t, err)
	assert.True(t, len(*series) == 3)
}

func testSearchContent(t *testing.T) {
	c := newContext()

	mov, ser, err := testDB.SearchContent(c, "Lord of", []string{"Fantasy"}, 1, 3)
	assert.Nil(t, err)
	assert.True(t, len(*mov) == 3) //there must be 3 items i guess

	for _, movie := range *mov {
		ok := false
		for _, g := range movie.Genres {
			if g == "Fantasy" {
				ok = true
				break
			}
		}
		assert.True(t, ok)
	}

	for _, serie := range *ser {
		ok := false
		for _, g := range serie.Genres {
			if g == "Fantasy" {
				ok = true
				break
			}
		}
		assert.True(t, ok)
	}

}

func testFindSimilarContent(t *testing.T) {
	c := newContext()
	_, series, err := testDB.FindSimilarContent(c, "tt16350094", "series")
	assert.Nil(t, err)
	assert.True(t, len(*series) == 2) //it suggests 2 series

	mov, _, err := testDB.FindSimilarContent(c, "tt0120737", "movie")
	assert.Nil(t, err)
	assert.True(t, len(*mov) == 2)
}
