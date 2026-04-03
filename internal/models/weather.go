package models

type WeatherResponse struct {
	Coord      Coord              `json:"coord"`
	Weather    []WeatherCondition `json:"weather"`
	Main       MainData           `json:"main"`
	Visibility int                `json:"visibility"`
	Wind       WindData           `json:"wind"`
	Clouds     CloudData          `json:"clouds"`
	Rain       *PrecipData        `json:"rain,omitempty"`
	Snow       *PrecipData        `json:"snow,omitempty"`
	Sys        SysData            `json:"sys"`
	Name       string             `json:"name"`
	Timezone   int                `json:"timezone"`
	Dt         int64              `json:"dt"`
}

type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type WeatherCondition struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type MainData struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int     `json:"pressure"`
	Humidity  int     `json:"humidity"`
	SeaLevel  int     `json:"sea_level,omitempty"`
	GrndLevel int     `json:"grnd_level,omitempty"`
}

type WindData struct {
	Speed float64 `json:"speed"`
	Deg   int     `json:"deg"`
	Gust  float64 `json:"gust,omitempty"`
}

type CloudData struct {
	All int `json:"all"`
}

type PrecipData struct {
	OneHour   float64 `json:"1h,omitempty"`
	ThreeHour float64 `json:"3h,omitempty"`
}

type SysData struct {
	Country string `json:"country"`
	Sunrise int64  `json:"sunrise"`
	Sunset  int64  `json:"sunset"`
}
