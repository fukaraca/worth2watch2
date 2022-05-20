package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/fukaraca/worth2watch2/api/admin"
	"github.com/fukaraca/worth2watch2/model"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"

	"testing"
	"time"
)

type mockCacheService struct {
	mock.Mock
}
type mockRepoService struct {
	mock.Mock
}

type mockUtils struct {
	mock.Mock
}

type mockService struct {
	*mockCacheService
	*mockRepoService
	*mockUtils
}

func newMockService() *mockService {
	return &mockService{
		mockCacheService: &mockCacheService{},
		mockRepoService:  &mockRepoService{},
		mockUtils:        &mockUtils{},
	}
}

func bindMockToService(m *mockService) *service {
	return &service{
		Repository: m.mockRepoService,
		Cache:      m.mockCacheService,
		Utilizer:   m.mockUtils,
	}
}

func newTestContext() (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return w, c
}

func newClient() http.Client {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
		Jar: jar,
	}
	return client
}

func TestAuth(t *testing.T) {
	mockservice := newMockService()
	w, c := newTestContext()
	//since context has no affiliated request, we need to create a request
	req, _ := http.NewRequest("GET", "/", nil)
	c.Request = req //and assing it

	//auth checks with cookies, so, we need to inject cookies into
	sessionToken, _ := uuid.NewV4()
	cookie1 := http.Cookie{Path: "/", Name: "session_token_test", Value: sessionToken.String(), MaxAge: 60, Domain: model.ServerHost, Secure: false, HttpOnly: true}
	cookie2 := http.Cookie{Path: "/", Name: "uid_test", Value: "username", MaxAge: 60, Domain: model.ServerHost, Secure: false, HttpOnly: true}
	c.Request.AddCookie(&cookie2)
	c.Request.AddCookie(&cookie1)

	fnc := func(ctx *gin.Context) {
		//string is the best
		ctx.String(201, "this is a test")
	}
	//succcesfull auth returns same handlerfunc. we can execute it by passing c into
	mockservice.mockCacheService.On("CheckSession", c).Return(true, nil)
	serv := bindMockToService(mockservice)
	serv.auth(fnc)(c)

	assert.Equal(t, 201, w.Code)

	tokenVal, _ := c.Cookie("session_token_test")
	assert.Equal(t, sessionToken.String(), tokenVal, "mismatching session token")
	assert.Equal(t, "this is a test", w.Body.String(), "mismatching body")
	////////////////////unauthorized attempt
	w, c = newTestContext()
	req, _ = http.NewRequest("GET", "/", nil)
	c.Request = req //and assing it
	mockservice.mockCacheService.On("CheckSession", c).Return(false, nil)
	serv.auth(fnc)(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

}

func TestCheckRegistration(t *testing.T) {

	mockservice := newMockService()
	//////////////////////first check for already logged-in user
	w, c := newTestContext()
	req, _ := http.NewRequest("POST", "/register", nil)
	c.Request = req //and assing it
	mockservice.mockCacheService.On("CheckSession", c).Return(true, nil)
	serv := bindMockToService(mockservice)
	serv.checkRegistration(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	r, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"notification\":\"already logged in\"}", string(r), "it's expected to complain as you've already logged in")
	///////////////////////////////////////////////////////////
	///////////////////check for a new user's successful register
	fn := func(user string) {
		w, c = newTestContext()
		req, _ = http.NewRequest("POST", "/register", bytes.NewBuffer([]byte(user)))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
	}
	userinfo := `{
    "username":"fukaraca",
    "password":"password",
    "email":"exa@mple1.com",
    "isAdmin":true
}`
	fn(userinfo)

	uname := "fukaraca"
	email := "exa@mple1.com"
	passw := "password"
	newUser := new(model.User)

	mockservice.mockCacheService.On("CheckSession", c).Return(false, nil)
	mockservice.mockUtils.On("Striper", uname).Return(&uname)
	mockservice.mockUtils.On("Striper", email).Return(&email)
	mockservice.mockUtils.On("Striper", passw).Return(&passw)
	mockservice.mockUtils.On("HashPassword", passw).Return("somePass", nil)
	mockservice.mockRepoService.On("CreateNewUser", c, newUser).Return(nil)

	serv = bindMockToService(mockservice)
	serv.checkRegistration(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	r, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"notification\":\"account created successfully\"}", string(r), "it's expected to return as successful ")
	//////////////////////////////////////////////////////
	///////////////check if username or password is empty
	userinfo = `{
    "username":"",
    "password":"password",
    "email":"exa@mple1.com",
    "isAdmin":false
}`
	fn(userinfo)
	mockservice.mockCacheService.On("CheckSession", c).Return(false, nil)

	serv = bindMockToService(mockservice)
	serv.checkRegistration(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	r, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"notification\":\"username or email or password cannot be empty\"}", string(r), "it's expected to return as bad request ")
	//////////////////////////////////////////////////
	////////////////// check if insertion to db failed
	userinfo = `{
    "username":"fukaraca",
    "password":"password",
    "email":"exa@mple1.com",
    "isAdmin":true
}`
	fn(userinfo)

	mockservice.mockCacheService.On("CheckSession", c).Return(false, nil)
	mockservice.mockUtils.On("Striper", uname).Return(&uname)
	mockservice.mockUtils.On("Striper", email).Return(&email)
	mockservice.mockUtils.On("Striper", passw).Return(&passw)
	mockservice.mockUtils.On("HashPassword", passw).Return("somePass", nil)
	mockservice.mockRepoService.On("CreateNewUser", c, newUser).Return(errors.New("some failure"))

	serv = bindMockToService(mockservice)
	serv.checkRegistration(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	r, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"notification\":\"some failure\"}", string(r), "it's expected to return failure ")
	mock.AssertExpectationsForObjects(t)
}

func TestLogin(t *testing.T) {
	mockservice := newMockService()
	//////////////first successful login
	w, c := newTestContext()
	fn := func(form url.Values) {
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		c.Request = req
	}
	formData := url.Values{
		"logUsername": []string{"user1"},
		"logPassword": []string{"pass1"},
	}
	fn(formData)
	mockservice.mockCacheService.On("CheckSession", c).Return(false, nil)
	mockservice.mockRepoService.On("QueryLogin", c, "user1").Return("hashedpass", nil)
	mockservice.mockUtils.On("CheckPasswordHash", "pass1", "hashedpass").Return(true)
	mockservice.mockRepoService.On("UpdateLastLogin", c, time.Now().Truncate(time.Minute), "user1").Return(nil)
	mockservice.mockCacheService.On("CreateSession", "user1", c).Return()

	serv := bindMockToService(mockservice)
	serv.login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	r, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"notification\":\"user user1 successfully logged in\"}", string(r), "its supposed to be successful entry")
	////////////////check unsuccessful login attempt for non-existed username
	w, c = newTestContext()
	formData = url.Values{
		"logUsername": []string{"userThatNotExist"},
		"logPassword": []string{"pass1"},
	}
	fn(formData)
	mockservice.mockCacheService.On("CheckSession", c).Return(false, nil)
	mockservice.mockRepoService.On("QueryLogin", c, "userThatNotExist").Return("", pgx.ErrNoRows)
	serv.login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	r, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"notification\":\""+pgx.ErrNoRows.Error()+"\"}", string(r), "its supposed to be non-existing username")

	////////////check unsuccessful login attempt for wrong password
	w, c = newTestContext()
	formData = url.Values{
		"logUsername": []string{"user1"},
		"logPassword": []string{"passThatWrong"},
	}
	fn(formData)
	mockservice.mockCacheService.On("CheckSession", c).Return(false, nil)
	mockservice.mockRepoService.On("QueryLogin", c, "user1").Return("hashedpass", nil)
	mockservice.mockUtils.On("CheckPasswordHash", "passThatWrong", "hashedpass").Return(false)

	serv.login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	r, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"notification\":\"password or username is incorrect\"}", string(r), "it's suppsed to be wrong password")
	//mock.AssertExpectationsForObjects(t)
}

func TestGetThisMovie(t *testing.T) {
	mockservice := newMockService()
	///////////////////////////
	//////////initial case for successful create with mock data
	id := "tt0120737"
	w, c := newTestContext()
	c.Params = gin.Params{{"id", id}}
	//prepare expected
	movExp := &model.Movie{}
	movExp.IMDBid = &id
	movExp.ReleaseDate.Set(time.Date(2022, 05, 10, 0, 0, 0, 0, time.UTC))
	//
	mockservice.mockRepoService.On("GetThisMovieFromDB", c, id).Return(movExp, nil)
	serv := bindMockToService(mockservice)
	serv.getThisMovie(c)
	assert.Equal(t, http.StatusOK, w.Code)
	r, _ := ioutil.ReadAll(w.Body)
	mov := model.Movie{}
	json.Unmarshal(r, &mov)

	assert.Equal(t, *mov.IMDBid, id)

	assert.Equal(t, mov.ReleaseDate.Time, time.Date(2022, 05, 10, 0, 0, 0, 0, time.UTC))
	///////////////////////////////////////////
	///////////another case successfull with real data
	w, c = newTestContext()
	c.Params = gin.Params{{"id", "tt0111161"}}
	//
	tmdb_id, err := admin.FindIDWithIMDB(id)
	if err != nil {
		t.Fatalf("movie id couldn't be get for imdb id:%v err: %v", id, err)
	}
	movExp = admin.GetMovie(tmdb_id)
	//
	mockservice.mockRepoService.On("GetThisMovieFromDB", c, "tt0111161").Return(movExp, nil)
	serv.getThisMovie(c)

	assert.Equal(t, http.StatusOK, w.Code)

	jsonAct, _ := ioutil.ReadAll(w.Body)
	jsonExp, _ := json.Marshal(movExp)
	assert.NotNil(t, jsonExp)
	//since directors writers and stars were fetched into map, equality of struct or json can vary by sorting,
	//so instead we will compare only first 50 bytes
	assert.Equal(t, len(jsonExp), len(jsonAct))
	assert.Equal(t, jsonExp[:50], jsonAct[:50])

	//////////////////////////////////
	//////////last case unsuccessful query for non-existed imdb-id
	w, c = newTestContext()
	c.Params = gin.Params{{"id", ""}}
	mockservice.mockRepoService.On("GetThisMovieFromDB", c, "").Return(nil, pgx.ErrNoRows)
	serv.getThisMovie(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	//

}

func TestGetMoviesWithPage(t *testing.T) {
	fn := func(pageK, itemsK, pageV, itemsV string, pageInt, ItemsInt int, retErr error, retMov *[]model.Movie) *httptest.ResponseRecorder {
		mockservice := newMockService()
		w, c := newTestContext()
		v := url.Values{}
		v.Add(pageK, pageV)
		v.Add(itemsK, itemsV)
		req, _ := http.NewRequest("GET", "/movies/list", nil)
		req.URL.RawQuery = v.Encode()
		c.Request = req
		mockservice.mockRepoService.On("GetMoviesListWithPage", c, pageInt, ItemsInt).Return(retMov, retErr)
		serv := bindMockToService(mockservice)
		serv.getMoviesWithPage(c)
		return w
	}
	tests := []struct {
		pK, pV, iK, iV, name string
		pI, iI               int
		rE                   error
		rM                   *[]model.Movie
		expected             int
		msg                  string
	}{
		{
			name:     "case for successfully created",
			pK:       "page",
			pV:       "1",
			iK:       "items",
			iV:       "10",
			pI:       1,
			iI:       10,
			rE:       nil,
			rM:       new([]model.Movie),
			expected: http.StatusOK,
			msg:      "it's supposed to get status ok",
		},
		{
			name:     "case for invalid page value",
			pK:       "page",
			pV:       "1e",
			iK:       "items",
			iV:       "10",
			pI:       1,
			iI:       10,
			rE:       nil,
			rM:       new([]model.Movie),
			expected: http.StatusBadRequest,
			msg:      "it's supposed to get bad request for invalid page format",
		},
		{
			name:     "case for invalid items value",
			pK:       "page",
			pV:       "1",
			iK:       "items",
			iV:       "1f0",
			pI:       1,
			iI:       10,
			rE:       nil,
			rM:       new([]model.Movie),
			expected: http.StatusBadRequest,
			msg:      "it's supposed to get bad request for invalid items format",
		},
		{
			name:     "case for internal serv error",
			pK:       "page",
			pV:       "1",
			iK:       "items",
			iV:       "10",
			pI:       1,
			iI:       10,
			rE:       errors.New("FAIL"),
			rM:       nil,
			expected: http.StatusInternalServerError,
			msg:      "it's supposed to get internal server error",
		},
		{
			name:     "case for end of list or no movie ",
			pK:       "page",
			pV:       "1",
			iK:       "items",
			iV:       "10",
			pI:       1,
			iI:       10,
			rE:       pgx.ErrNoRows,
			rM:       nil,
			expected: http.StatusOK,
			msg:      "it's supposed to get end of list err",
		},
	}

	for _, test := range tests {
		w := fn(test.pK, test.iK, test.pV, test.iV, test.pI, test.iI, test.rE, test.rM)
		assert.Equal(t, test.expected, w.Code, test.msg+" for "+test.name)

	}

}

func TestAddContentWithJSON(t *testing.T) {
	mockservice := newMockService()
	w, c := newTestContext()
	fn := func(content, content_type string) {
		form := url.Values{
			"content-type":     []string{content_type},
			"content-raw-data": []string{content},
		}
		requ, _ := http.NewRequest("POST", "/contentWithJSON", strings.NewReader(form.Encode()))
		requ.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		c.Request = requ
	}
	////////////////
	//////////////// first check role of user!=admin

	contentStr := `{"adult":false,"backdrop_path":"/vRQnzOn4HjIMX4LBq9nHhFXbsSu.jpg","belongs_to_collection":{"id":119,"name":"The Lord of the Rings Collection","poster_path":"/nSNle6UJNNuEbglNvXt67m1a1Yn.jpg","backdrop_path":"/bccR2CGTWVVSZAG0yqmy3DIvhTX.jpg"},"budget":93000000,"genres":[{"id":12,"name":"Adventure"},{"id":14,"name":"Fantasy"},{"id":28,"name":"Action"}],"homepage":"http://www.lordoftherings.net/","id":120,"imdb_id":"tt0120737","original_language":"en","original_title":"The Lord of the Rings: The Fellowship of the Ring","overview":"Young hobbit Frodo Baggins, after inheriting a mysterious ring from his uncle Bilbo, must leave his home in order to keep it from falling into the hands of its evil creator. Along the way, a fellowship is formed to protect the ringbearer and make sure that the ring arrives at its final destination: Mt. Doom, the only place where it can be destroyed.","popularity":118.779,"poster_path":"/6oom5QYQ2yQTMJIbnvbkBL9cHo6.jpg","production_companies":[{"id":12,"logo_path":"/iaYpEp3LQmb8AfAtmTvpqd4149c.png","name":"New Line Cinema","origin_country":"US"},{"id":11,"logo_path":"/6FAuASQHybRkZUk08p9PzSs9ezM.png","name":"WingNut Films","origin_country":"NZ"},{"id":5237,"logo_path":null,"name":"The Saul Zaentz Company","origin_country":"US"}],"production_countries":[{"iso_3166_1":"NZ","name":"New Zealand"},{"iso_3166_1":"US","name":"United States of America"}],"release_date":"2001-12-18","revenue":871368364,"runtime":179,"spoken_languages":[{"english_name":"English","iso_639_1":"en","name":"English"}],"status":"Released","tagline":"One ring to rule them all","title":"The Lord of the Rings: The Fellowship of the Ring","video":false,"vote_average":8.4,"vote_count":20776}`
	fn(contentStr, "movie")
	mockservice.mockCacheService.On("CheckAdminForLoggedIn", c, "").Return(false)

	serv := bindMockToService(mockservice)
	serv.addContentWithJSON(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	r, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"notification\":\"unauthorized attempt\"}", string(r), "it's supposed to be: unauthorized attempt")

	////////////////////////////////////////////
	/////////secondly check for invalid content type
	w, c = newTestContext()
	fn(contentStr, "invalid content")
	mockservice.mockCacheService.On("CheckAdminForLoggedIn", c, "").Return(true)
	serv = bindMockToService(mockservice)
	serv.addContentWithJSON(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	r, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"notification\":\"invalid content-type: invalid content\"}", string(r), "it's supposed to be: unauthorized attempt")
	////////////////////////////////////////////////////////
	//////////finally check for successful attempt
	w, c = newTestContext()
	fn(contentStr, "movie")
	mockservice.mockCacheService.On("CheckAdminForLoggedIn", c, "").Return(true)
	mockservice.mockRepoService.On("AddMovieContentWithStruct", c, new(model.Movie)).Return(nil)
	serv = bindMockToService(mockservice)
	serv.addContentWithJSON(c)
	assert.Equal(t, http.StatusCreated, w.Code)
	r, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"notification\":\"content has been created successfully\"}", string(r), "it's supposed to be: content has been created successfully")

}

func TestAddToFavorites(t *testing.T) {

	fn := func(form url.Values, status int, expected, expected_msg string, ret error) {
		mockservice := newMockService()
		w, c := newTestContext()
		requ, _ := http.NewRequest("POST", "/favorites", strings.NewReader(form.Encode()))
		requ.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		c.Request = requ
		mockservice.mockRepoService.On("AddContentToFavorites", c, "", "").Return(ret)
		serv := bindMockToService(mockservice)
		serv.addToFavorites(c)
		assert.Equal(t, status, w.Code)
		r, _ := ioutil.ReadAll(w.Body)
		assert.Equal(t, expected, string(r), expected_msg)

	}
	tt := []struct {
		name                   string
		urls                   url.Values
		status                 int
		expected, expected_msg string
		ret                    error
	}{
		{
			name: "case for successful creation",
			urls: url.Values{
				"imdb-id":      []string{"tt0111161"},
				"content-type": []string{"movie"}},
			status:       http.StatusCreated,
			expected:     "{\"notification\":\"content has been added successfully\"}",
			expected_msg: "it's supposed to get succesful creation",
			ret:          nil,
		},
		{
			name: "case for successful creation",
			urls: url.Values{
				"imdb-id":      []string{"tt0111162"},
				"content-type": []string{"series"}},
			status:       http.StatusCreated,
			expected:     "{\"notification\":\"content has been added successfully\"}",
			expected_msg: "it's supposed to get succesful creation",
			ret:          nil,
		},
		{
			name: "case for missing id",
			urls: url.Values{
				"imdb-id":      []string{""},
				"content-type": []string{"movie"}},
			status:       http.StatusBadRequest,
			expected:     "{\"notification\":\"missing imdb-id\"}",
			expected_msg: "it's supposed to get missing imdb-id error",
			ret:          nil,
		},
		{
			name: "case for invalid content type",
			urls: url.Values{
				"imdb-id":      []string{"tt0111162"},
				"content-type": []string{""}},
			status:       http.StatusBadRequest,
			expected:     "{\"notification\":\"invalid content-type: \"}",
			ret:          nil,
			expected_msg: "it's supposed to get missing content-type error",
		},
		{
			name: "case for insertion failure",
			urls: url.Values{
				"imdb-id":      []string{"tt0111162"},
				"content-type": []string{"movie"}},
			status:       http.StatusInternalServerError,
			expected:     "{\"notification\":\"FAIL\"}",
			ret:          errors.New("FAIL"),
			expected_msg: "it's supposed to be gotten insertion error",
		},
	}
	for _, t := range tt {
		fn(t.urls, t.status, t.expected, t.expected_msg, t.ret)
	}

}

func TestSearchContent(t *testing.T) {
	fn := func(pageK, itemsK, pageV, itemsV, nameK, nameV, genreK, shBody string, genreV []string, pageInt, ItemsInt int, retMov *[]model.Movie, retSer *[]model.Series, retErr error) *httptest.ResponseRecorder {
		mockservice := newMockService()
		w, c := newTestContext()
		v := url.Values{}
		v.Add(pageK, pageV)
		v.Add(itemsK, itemsV)
		for _, s := range genreV {
			v.Add(genreK, s)
		}
		req, _ := http.NewRequest("GET", "/search", nil)
		req.URL.RawQuery = v.Encode()
		c.Request = req
		mockservice.mockRepoService.On("SearchContent", c, nameV, genreV, pageInt, ItemsInt).Return(retMov, retSer, retErr)
		serv := bindMockToService(mockservice)
		serv.searchContent(c)
		return w
	}
	tests := []struct {
		pK, pV, iK, iV, nK, nV, gK, name string
		gV                               []string
		pI, iI                           int
		rE                               error
		rM                               *[]model.Movie
		rS                               *[]model.Series
		expected                         int
		shortBody                        string
		msg                              string
	}{
		{
			name:      "case for successful search",
			nK:        "name",
			nV:        "",
			gK:        "genre",
			gV:        []string{},
			pK:        "page",
			pV:        "1",
			iK:        "items",
			iV:        "10",
			pI:        1,
			iI:        10,
			rE:        nil,
			rM:        new([]model.Movie),
			rS:        new([]model.Series),
			expected:  http.StatusOK,
			shortBody: "",
			msg:       "it's supposed to get status ok",
		},
		{
			name:      "case for invalid page format",
			nK:        "name",
			nV:        "",
			gK:        "genre",
			gV:        []string{},
			pK:        "page",
			pV:        "1w",
			iK:        "items",
			iV:        "10",
			pI:        1,
			iI:        10,
			rE:        nil,
			rM:        nil,
			rS:        nil,
			expected:  http.StatusBadRequest,
			shortBody: "{\"notification\":\"invalid page format\"}",
			msg:       "it's supposed to inv page form err",
		},
		{
			name:      "case for invalid items format",
			nK:        "name",
			nV:        "",
			gK:        "genre",
			gV:        []string{},
			pK:        "page",
			pV:        "1",
			iK:        "items",
			iV:        "10zxc",
			pI:        1,
			iI:        10,
			rE:        nil,
			rM:        nil,
			rS:        nil,
			expected:  http.StatusBadRequest,
			shortBody: "{\"notification\":\"invalid items format\"}",
			msg:       "it's supposed to inv items form err",
		},
		{
			name:      "failure of query",
			nK:        "name",
			nV:        "",
			gK:        "genre",
			gV:        []string{},
			pK:        "page",
			pV:        "1",
			iK:        "items",
			iV:        "10",
			pI:        1,
			iI:        10,
			rE:        errors.New("FAIL"),
			rM:        nil,
			rS:        nil,
			expected:  http.StatusInternalServerError,
			shortBody: "{\"notification\":\"en error occurred\"}",
			msg:       "it's supposed to failure of query",
		},
		{
			name:      "case of no movie and series have been found",
			nK:        "name",
			nV:        "",
			gK:        "genre",
			gV:        []string{},
			pK:        "page",
			pV:        "1",
			iK:        "items",
			iV:        "10",
			pI:        1,
			iI:        10,
			rE:        nil,
			rM:        nil,
			rS:        nil,
			expected:  http.StatusOK,
			shortBody: "{\"notification\":\"no content found for given filter\"}",
			msg:       "it's supposed to happen no content found",
		},
	}

	for _, test := range tests {
		w := fn(test.pK, test.iK, test.pV, test.iV, test.nK, test.nV, test.gK, test.shortBody, test.gV, test.pI, test.iI, test.rM, test.rS, test.rE)
		assert.Equal(t, test.expected, w.Code, test.msg+" for "+test.name)
		if test.shortBody != "" {
			r, _ := ioutil.ReadAll(w.Body)
			assert.Equal(t, test.shortBody, string(r), "body err for"+test.name)
		}

	}
}

func TestGetThisSeries(t *testing.T) {

	fn := func(idK, idV string, series *model.Series, seasons *[]model.Seasons, fail error) *httptest.ResponseRecorder {
		mockservice := newMockService()
		w, c := newTestContext()
		c.Params = gin.Params{{idK, idV}}
		mockservice.mockRepoService.On("GetThisSeriesFromDB", c, idV).Return(series, seasons, fail)
		serv := bindMockToService(mockservice)
		serv.getThisSeries(c)
		return w
	}

	tests := []struct {
		name, idK, idV string
		series         *model.Series
		seasons        *[]model.Seasons
		err            error
		expStatus      int
		msg, shortBody string
	}{
		{
			name:      "case for missing id",
			idK:       "seriesid",
			idV:       "",
			series:    nil,
			seasons:   nil,
			err:       nil,
			expStatus: http.StatusBadRequest,
			msg:       "it's expected to get bad request=",
			shortBody: "{\"notification\":\"missing serie id\"}",
		},
		{
			name:      "case for no such series",
			idK:       "seriesid",
			idV:       "1",
			series:    nil,
			seasons:   nil,
			err:       pgx.ErrNoRows,
			expStatus: http.StatusNotFound,
			msg:       "it's expected to get 404=",
			shortBody: "{\"notification\":\"no such serie\"}",
		},
		{
			name:      "case for no season for that series",
			idK:       "seriesid",
			idV:       "1",
			series:    new(model.Series),
			seasons:   nil,
			err:       errors.New("there is no season for given series"),
			expStatus: http.StatusOK,
			msg:       "it's expected to get 200=",
			shortBody: "", //body check has been neglected
		},
		{
			name:      "case for another failure with query",
			idK:       "seriesid",
			idV:       "1",
			series:    nil,
			seasons:   nil,
			err:       errors.New("FAIL"),
			expStatus: http.StatusInternalServerError,
			msg:       "it's expected to get internal serv err=",
			shortBody: "{\"notification\":\"FAIL\"}",
		},
		{
			name:      "case for successful response",
			idK:       "seriesid",
			idV:       "1",
			series:    new(model.Series),
			seasons:   new([]model.Seasons),
			err:       nil,
			expStatus: http.StatusOK,
			msg:       "it's expected to get 200=",
			shortBody: "",
		},
	}
	for _, test := range tests {
		w := fn(test.idK, test.idV, test.series, test.seasons, test.err)
		assert.Equal(t, test.expStatus, w.Code, test.msg+" for "+test.name)
		if test.shortBody != "" {
			r, _ := ioutil.ReadAll(w.Body)
			assert.Equal(t, test.shortBody, string(r), "body err for"+test.name)
		}
	}

}

func TestLogout(t *testing.T) {
	fn := func(ok bool, err error) *httptest.ResponseRecorder {
		mockservice := newMockService()
		w, c := newTestContext()
		mockservice.mockCacheService.On("DeleteSession", c).Return(ok, err)
		serv := bindMockToService(mockservice)
		serv.logout(c)
		return w
	}

	tests := []struct {
		name, msg, shortBody string
		err                  error
		expStatus            int
		retCheck             bool
	}{
		{
			name:      "case for failure of session deletion",
			err:       errors.New("FAIL"),
			retCheck:  false,
			expStatus: http.StatusInternalServerError,
			shortBody: "{\"notification\":\"FAIL\"}",
			msg:       "it's expected to get intern serv err",
		},
		{
			name:      "case for success",
			err:       nil,
			retCheck:  true,
			expStatus: http.StatusOK,
			shortBody: "{\"notification\":\"logged out successfully\"}",
			msg:       "it's expected to get success for= ",
		},
	}

	for _, test := range tests {
		w := fn(test.retCheck, test.err)
		assert.Equal(t, test.expStatus, w.Code, test.msg+test.name)
		r, _ := ioutil.ReadAll(w.Body)
		assert.Equal(t, test.shortBody, string(r), test.msg+test.name)
	}
}

func (s *mockCacheService) InitializeCache() {

}

func (s *mockCacheService) CheckCookie(c *gin.Context, toBeChecked, userId string) (bool, error) {
	return false, nil
}

func (s *mockCacheService) CreateSession(username string, c *gin.Context) {

}

func (s *mockCacheService) CheckSession(c *gin.Context) (bool, error) {
	args := s.Called(c)
	return args.Bool(0), args.Error(1)

}

func (s *mockCacheService) DeleteSession(c *gin.Context) (bool, error) {
	args := s.Called(c)
	return args.Bool(0), args.Error(1)
}

func (s *mockCacheService) CheckAdminForLoggedIn(c *gin.Context, username string) bool {

	args := s.Called(c, username)
	return args.Bool(0)
}

func (s *mockCacheService) CloseCacheConnection() {
}

func (s *mockRepoService) CloseDB() {

}

func (s *mockRepoService) GetThisMovieFromDB(c *gin.Context, id string) (*model.Movie, error) {

	args := s.Called(c, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Movie), args.Error(1)
}

func (s *mockRepoService) GetThisSeriesFromDB(c *gin.Context, id string) (*model.Series, *[]model.Seasons, error) {
	args := s.Called(c, id)
	var ret1 *model.Series
	var ret2 *[]model.Seasons
	if args.Get(0) == nil {
		ret1 = nil
	} else {
		ret1 = args.Get(0).(*model.Series)
	}
	if args.Get(1) == nil {
		ret2 = nil
	} else {
		ret2 = args.Get(1).(*[]model.Seasons)
	}
	return ret1, ret2, args.Error(2)
}

func (s *mockRepoService) GetEpisodesForaSeasonFromDB(c *gin.Context, seriesID, sN string) (*[]model.Episodes, error) {
	return nil, nil
}

func (s *mockRepoService) GetMoviesListWithPage(c *gin.Context, page, items int) (*[]model.Movie, error) {
	args := s.Called(c, page, items)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.Movie), args.Error(1)
}

func (s *mockRepoService) GetSeriesListWithPage(c *gin.Context, page, items int) (*[]model.Series, error) {
	return nil, nil
}

func (s *mockRepoService) SearchContent(c *gin.Context, name string, genres []string, page, items int) (*[]model.Movie, *[]model.Series, error) {
	genres = []string{}
	args := s.Called(c, name, genres, page, items)

	//cases for unary null is neglected
	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*[]model.Movie), args.Get(1).(*[]model.Series), args.Error(2)
}

func (s *mockRepoService) FindSimilarContent(c *gin.Context, id, cType string) (*[]model.Movie, *[]model.Series, error) {
	return nil, nil, nil
}

func (s *mockRepoService) AddContentToFavorites(c *gin.Context, IMDB, cType string) error {

	args := s.Called(c, "", "")
	return args.Error(0)
}

func (s *mockRepoService) GetFavoriteContents(c *gin.Context, page, items int) (*[]model.Movie, *[]model.Series, error) {
	return nil, nil, nil
}

func (s *mockRepoService) SearchFavorites(c *gin.Context, name string, genres []string, page, items int) (*[]model.Movie, *[]model.Series, error) {
	return nil, nil, nil
}

func (s *mockRepoService) QueryLogin(c *gin.Context, username string) (string, error) {
	args := s.Called(c, username)
	return args.String(0), args.Error(1)
}

func (s *mockRepoService) CreateNewUser(c *gin.Context, newUser *model.User) error {
	newUser = new(model.User)
	args := s.Called(c, newUser)
	return args.Error(0)

}

func (s *mockRepoService) UpdateLastLogin(c *gin.Context, lastLoginTime time.Time, logUsername string) error {
	lastLoginTime = lastLoginTime.Truncate(time.Minute)
	args := s.Called(c, lastLoginTime, logUsername)
	return args.Error(0)
}

func (s *mockRepoService) UpdateUserInfo(c *gin.Context, firstname, lastname, username string) error {
	return nil
}

func (s *mockRepoService) QueryUserInfo(c *gin.Context, username string) (*model.User, error) {
	return nil, nil
}

func (s *mockRepoService) AddMovieContentWithID(imdb string) {

}

func (s *mockRepoService) AddSeriesContentWithID(imdb string) {

}

func (s *mockRepoService) AddMovieContentWithStruct(ctx context.Context, movie *model.Movie) error {
	movie = new(model.Movie)
	args := s.Called(ctx, movie)
	return args.Error(0)
}

func (s *mockRepoService) AddSeriesContentWithStruct(ctx context.Context, series *model.Series) error {
	series = new(model.Series)
	args := s.Called(ctx, series)
	return args.Error(0)
}

func (s *mockRepoService) DeleteContent(c *gin.Context, id, contentType string) error {
	return nil
}

func (m *mockUtils) Striper(str string) *string {
	arg := m.Called(str)
	if str == "" {
		return nil
	}
	return arg.Get(0).(*string)
}

func (m *mockUtils) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *mockUtils) CheckPasswordHash(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}
