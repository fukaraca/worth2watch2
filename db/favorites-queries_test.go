package db_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func testAddContentToFavorites(t *testing.T) {
	c := newContext()

	//happy case for movie Lotr -fantasy
	err := testDB.AddContentToFavorites(c, "tt0120737", "movie", adminSampleString)
	assert.Nil(t, err)

	//happy case for series The Staircase -drama
	err = testDB.AddContentToFavorites(c, "tt11324406", "series", adminSampleString)
	assert.Nil(t, err)

}

func testgetFavoriteContents(t *testing.T) {
	c := newContext()

	mov, ser, err := testDB.GetFavoriteContents(c, 1, 2, adminSampleString) //we just inserted 2 content for admin
	assert.Nil(t, err)
	assert.Equal(t, 1, len(*mov))
	assert.Equal(t, 1, len(*ser))

	mov, ser, err = testDB.GetFavoriteContents(c, 1, 2, uniqueUser.Username) //unique user has no favorite
	assert.Nil(t, err)
	assert.Equal(t, 0, len(*mov))
	assert.Equal(t, 0, len(*ser))
}

func testSearchFavorites(t *testing.T) {
	c := newContext()

	mov, ser, err := testDB.SearchFavorites(c, "Lotr", adminSampleString, []string{"Fantasy"}, 1, 5)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(*mov))
	assert.Equal(t, 0, len(*ser))

	//only check genres
	mov, ser, err = testDB.SearchFavorites(c, "", adminSampleString, []string{"Drama"}, 1, 5)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(*mov))
	assert.Equal(t, 1, len(*ser))

}
