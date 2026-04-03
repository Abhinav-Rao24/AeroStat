package display

import (
	"fmt"
	"io"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"go-weather-cli/internal/models"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"

	cyan    = "\033[36m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	green   = "\033[32m"
	red     = "\033[31m"
	magenta = "\033[35m"
	white   = "\033[97m"
	gray    = "\033[90m"
)

func colorize(color, s string) string { return color + s + reset }
func bold_(s string) string           { return bold + s + reset }
func dim_(s string) string            { return dim + s + reset }

func titleCase(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return strings.ToUpper(string(r)) + strings.ToLower(s[size:])
}

func conditionIcon(id int) string {
	switch {
	case id >= 200 && id < 300:
		return "[Thunder]"
	case id >= 300 && id < 400:
		return "[Drizzle]"
	case id >= 500 && id < 600:
		return "[Rain]"
	case id >= 600 && id < 700:
		return "[Snow]"
	case id >= 700 && id < 800:
		return "[Fog/Mist]"
	case id == 800:
		return "[Clear]"
	case id == 801:
		return "[Few Clouds]"
	case id == 802:
		return "[Scattered]"
	case id >= 803:
		return "[Overcast]"
	default:
		return "[Weather]"
	}
}

func windDirection(deg int) string {
	dirs := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	idx := int(math.Round(float64(deg)/45)) % 8
	return dirs[idx]
}

func tempUnit(units string) string {
	switch units {
	case "imperial":
		return "F"
	case "standard":
		return "K"
	default:
		return "C"
	}
}

func speedUnit(units string) string {
	if units == "imperial" {
		return "mph"
	}
	return "m/s"
}

const boxWidth = 52

func hLine(left, mid, right string) string {
	return left + strings.Repeat(mid, boxWidth) + right
}

func row(label, value, labelColor, valueColor string) string {
	const labelWidth = 18
	paddedLabel := fmt.Sprintf("%-*s", labelWidth, label)
	inner := colorize(labelColor, paddedLabel) + colorize(valueColor, value)
	return "| " + inner + " |"
}

func separator() string {
	return colorize(gray, "|"+strings.Repeat("-", boxWidth)+"|")
}

func topBar() string    { return colorize(cyan, hLine("+", "-", "+")) }
func bottomBar() string { return colorize(cyan, hLine("+", "-", "+")) }
func sideRow(content string) string {
	return colorize(cyan, "|") + " " + content + colorize(cyan, "|")
}

func centreText(s string, color string) string {
	visLen := len(s)
	pad := (boxWidth - visLen) / 2
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) + colorize(color, s)
}

func PrintWeather(w io.Writer, d *models.WeatherResponse, units string) {
	tUnit := tempUnit(units)
	sUnit := speedUnit(units)

	var icon, desc string
	if len(d.Weather) > 0 {
		icon = conditionIcon(d.Weather[0].ID)
		desc = titleCase(d.Weather[0].Description)
	}

	tz := time.FixedZone("local", d.Timezone)
	sunrise := time.Unix(d.Sys.Sunrise, 0).In(tz).Format("15:04")
	sunset := time.Unix(d.Sys.Sunset, 0).In(tz).Format("15:04")
	updated := time.Unix(d.Dt, 0).In(tz).Format("02 Jan 2006, 15:04 MST")

	var precipLine string
	if d.Rain != nil {
		if d.Rain.OneHour > 0 {
			precipLine = fmt.Sprintf("%.1f mm/h", d.Rain.OneHour)
		} else if d.Rain.ThreeHour > 0 {
			precipLine = fmt.Sprintf("%.1f mm/3h", d.Rain.ThreeHour)
		}
	}
	if d.Snow != nil && precipLine == "" {
		if d.Snow.OneHour > 0 {
			precipLine = fmt.Sprintf("%.1f mm/h snow", d.Snow.OneHour)
		}
	}

	var visStr string
	if d.Visibility > 0 {
		if d.Visibility >= 1000 {
			visStr = fmt.Sprintf("%.1f km", float64(d.Visibility)/1000)
		} else {
			visStr = fmt.Sprintf("%d m", d.Visibility)
		}
	} else {
		visStr = "N/A"
	}

	lines := []string{}
	add := func(s string) { lines = append(lines, s) }

	add(topBar())

	header := fmt.Sprintf("%s, %s", bold_(d.Name), d.Sys.Country)
	add(sideRow(centreText(header, white)))

	condLine := icon + " " + desc
	add(sideRow(centreText(condLine, yellow)))

	add(sideRow(centreText(dim_("Updated: "+updated), "")))
	add(colorize(cyan, separator()))

	add(row("Temperature", fmt.Sprintf("%.1f%s  (feels %.1f%s)", d.Main.Temp, tUnit, d.Main.FeelsLike, tUnit), cyan, white))
	add(row("Range", fmt.Sprintf("%.1f%s - %.1f%s", d.Main.TempMin, tUnit, d.Main.TempMax, tUnit), cyan, white))
	add(colorize(gray, separator()))

	add(row("Humidity", fmt.Sprintf("%d%%", d.Main.Humidity), cyan, blue))
	add(row("Pressure", fmt.Sprintf("%d hPa", d.Main.Pressure), cyan, blue))
	add(row("Visibility", visStr, cyan, blue))
	add(row("Cloud Cover", fmt.Sprintf("%d%%", d.Clouds.All), cyan, blue))

	if precipLine != "" {
		add(row("Precipitation", precipLine, cyan, blue))
	}
	add(colorize(gray, separator()))

	gustStr := ""
	if d.Wind.Gust > 0 {
		gustStr = fmt.Sprintf(" (gust %.1f %s)", d.Wind.Gust, sUnit)
	}
	add(row("Wind", fmt.Sprintf("%.1f %s%s", d.Wind.Speed, sUnit, gustStr), cyan, green))
	add(row("Direction", windDirection(d.Wind.Deg), cyan, green))
	add(colorize(gray, separator()))

	add(row("Sunrise", sunrise, cyan, yellow))
	add(row("Sunset", sunset, cyan, yellow))
	add(colorize(gray, separator()))
	add(row("Coordinates", fmt.Sprintf("%.4f, %.4f", d.Coord.Lat, d.Coord.Lon), cyan, magenta))

	add(bottomBar())

	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
}

func PrintError(w io.Writer, err error) {
	fmt.Fprintln(w, colorize(red, bold_("Error: ")+err.Error()))
}

func PrintCached(w io.Writer) {
	fmt.Fprintln(w, colorize(gray, dim_("  served from cache")))
}
