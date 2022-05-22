package db_test

import (
	"context"
	"github.com/fukaraca/worth2watch2/api/admin"
	"github.com/fukaraca/worth2watch2/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	tt3501632  *model.Movie  //will be added by struct
	tt2250912  *model.Movie  //will be added by id
	tt13650480 *model.Series //will be added by id
	tt16350094 *model.Series //will be add by struct

)

func testAddMovieContent(t *testing.T) {

	//added by id
	testDB.AddMovieContentWithID("tt2250912")
	id, err := admin.FindIDWithIMDB("tt2250912")
	assert.Nil(t, err)
	tt2250912 = admin.GetMovie(id)
	//

	id, err = admin.FindIDWithIMDB("tt3501632")
	assert.Nil(t, err)
	tt3501632 = admin.GetMovie(id)
	err = testDB.AddMovieContentWithStruct(context.Background(), tt3501632)
	assert.Nil(t, err, "we expect only nil as err")

}

func testAddSeriesContent(t *testing.T) {
	//add by id
	testDB.AddSeriesContentWithID("tt13650480")
	id, err := admin.FindIDWithIMDB("tt13650480")
	assert.Nil(t, err)
	tt13650480 = admin.GetSeries(id)

	//add by struct
	id, err = admin.FindIDWithIMDB("tt16350094")
	assert.Nil(t, err)
	tt16350094 = admin.GetSeries(id)
	err = testDB.AddSeriesContentWithStruct(context.Background(), tt16350094)
	assert.Nil(t, err, "we expect only nil as err")
}
