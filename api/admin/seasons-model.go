package admin

type SeasonsAPI struct {
	IDirrelevant string `json:"_id"`
	AirDate      string `json:"air_date"`
	Episodes     []struct {
		AirDate       string `json:"air_date"`
		EpisodeNumber int    `json:"episode_number"`
		Crew          []struct {
			Department         string  `json:"department"`
			Job                string  `json:"job"`
			CreditID           string  `json:"credit_id"`
			Adult              bool    `json:"adult"`
			Gender             int     `json:"gender"`
			ID                 int     `json:"id"`
			KnownForDepartment string  `json:"known_for_department"`
			Name               string  `json:"name"`
			OriginalName       string  `json:"original_name"`
			Popularity         float64 `json:"popularity"`
			ProfilePath        string  `json:"profile_path"`
		} `json:"crew"`
		GuestStars []struct {
			Character          string  `json:"character"`
			CreditID           string  `json:"credit_id"`
			Order              int     `json:"order"`
			Adult              bool    `json:"adult"`
			Gender             int     `json:"gender"`
			ID                 int     `json:"id"`
			KnownForDepartment string  `json:"known_for_department"`
			Name               string  `json:"name"`
			OriginalName       string  `json:"original_name"`
			Popularity         float64 `json:"popularity"`
			ProfilePath        string  `json:"profile_path"`
		} `json:"guest_stars"`
		ID             int     `json:"id"`
		Name           string  `json:"name"`
		Overview       string  `json:"overview"`
		ProductionCode string  `json:"production_code"`
		SeasonNumber   int     `json:"season_number"`
		StillPath      string  `json:"still_path"`
		VoteAverage    float64 `json:"vote_average"`
		VoteCount      int     `json:"vote_count"`
	} `json:"episodes"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	ID           int    `json:"id"`
	PosterPath   string `json:"poster_path"`
	SeasonNumber int    `json:"season_number"`
}

type EpisodeCastAPI struct {
	Cast []struct {
		Adult              bool    `json:"adult"`
		Gender             int     `json:"gender"`
		ID                 int     `json:"id"`
		KnownForDepartment string  `json:"known_for_department"`
		Name               string  `json:"name"`
		OriginalName       string  `json:"original_name"`
		Popularity         float64 `json:"popularity"`
		ProfilePath        string  `json:"profile_path"`
		Character          string  `json:"character"`
		CreditID           string  `json:"credit_id"`
		Order              int     `json:"order"`
	} `json:"cast"`
	Crew []struct {
		Job                string      `json:"job"`
		Department         string      `json:"department"`
		CreditID           string      `json:"credit_id"`
		Adult              bool        `json:"adult"`
		Gender             int         `json:"gender"`
		ID                 int         `json:"id"`
		KnownForDepartment string      `json:"known_for_department"`
		Name               string      `json:"name"`
		OriginalName       string      `json:"original_name"`
		Popularity         float64     `json:"popularity"`
		ProfilePath        interface{} `json:"profile_path"`
	} `json:"crew"`
	GuestStars []struct {
		Character          string  `json:"character"`
		CreditID           string  `json:"credit_id"`
		Order              int     `json:"order"`
		Adult              bool    `json:"adult"`
		Gender             int     `json:"gender"`
		ID                 int     `json:"id"`
		KnownForDepartment string  `json:"known_for_department"`
		Name               string  `json:"name"`
		OriginalName       string  `json:"original_name"`
		Popularity         float64 `json:"popularity"`
		ProfilePath        string  `json:"profile_path"`
	} `json:"guest_stars"`
	ID int `json:"id"`
}

type TranslationDataOfEpisode struct {
	ID           int `json:"id"`
	Translations []struct {
		Iso31661    string `json:"iso_3166_1"`
		Iso6391     string `json:"iso_639_1"`
		Name        string `json:"name"`
		EnglishName string `json:"english_name"`
		Data        struct {
			Name     string `json:"name"`
			Overview string `json:"overview"`
		} `json:"data"`
	} `json:"translations"`
}
