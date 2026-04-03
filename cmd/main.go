package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"go-weather-cli/internal/cache"
	"go-weather-cli/internal/client"
	"go-weather-cli/internal/config"
	"go-weather-cli/internal/display"
	"go-weather-cli/internal/models"
)

const version = "1.0.0"

func main() {
	os.Exit(run())
}

type task struct {
	city   string
	lat    float64
	lon    float64
	byCity bool
}

type result struct {
	t     task
	resp  *models.WeatherResponse
	err   error
	cache bool
}

func run() int {
	cityFlag := flag.String("city", "", `City name or comma-separated list, e.g. "London,New York"`)
	latFlag := flag.Float64("lat", 0, "Latitude (pair with -lon)")
	lonFlag := flag.Float64("lon", 0, "Longitude (pair with -lat)")
	unitsFlag := flag.String("units", "", "Unit system: metric | imperial | standard")
	noCacheFlag := flag.Bool("no-cache", false, "Always fetch fresh data")
	versionFlag := flag.Bool("version", false, "Print version and exit")

	flag.Usage = printUsage
	flag.Parse()

	if *versionFlag {
		fmt.Printf("AeroStat v%s\n", version)
		return 0
	}

	cityVal := strings.TrimSpace(*cityFlag)
	var cities []string
	if cityVal != "" {
		for _, c := range strings.Split(cityVal, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				cities = append(cities, c)
			}
		}
	}
	
	byCity := len(cities) > 0
	byCoords := *latFlag != 0 || *lonFlag != 0

	if !byCity && !byCoords {
		fmt.Fprintln(os.Stderr, "error: supply -city <name(s)> OR -lat <f> -lon <f>")
		flag.Usage()
		return 2
	}
	if byCoords && (*latFlag < -90 || *latFlag > 90 || *lonFlag < -180 || *lonFlag > 180) {
		fmt.Fprintln(os.Stderr, "error: latitude must be [-90,90], longitude [-180,180]")
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		display.PrintError(os.Stderr, err)
		return 1
	}
	if *unitsFlag != "" {
		cfg.Units = *unitsFlag
	}

	ttl := cfg.CacheTTL
	if *noCacheFlag {
		ttl = 0
	}
	weatherCache := cache.New(ttl)
	weatherClient := client.New(cfg.APIKey)

	var tasks []task
	if byCity {
		for _, c := range cities {
			tasks = append(tasks, task{city: c, byCity: true})
		}
	}
	if byCoords {
		tasks = append(tasks, task{lat: *latFlag, lon: *lonFlag, byCity: false})
	}

	numWorkers := runtime.NumCPU()
	if numWorkers > len(tasks) {
		numWorkers = len(tasks)
	}

	jobs := make(chan task, len(tasks))
	results := make(chan result, len(tasks))

	for _, t := range tasks {
		jobs <- t
	}
	close(jobs)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\n[Interrupt] Cancelling pending tasks...")
		cancel()
	}()

	rateLimiter := time.NewTicker(200 * time.Millisecond)
	defer rateLimiter.Stop()

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for w := 0; w < numWorkers; w++ {
		go func() {
			defer wg.Done()
			for t := range jobs {
				var cacheKey string
				if t.byCity {
					cacheKey = fmt.Sprintf("city:%s:%s", strings.ToLower(t.city), cfg.Units)
				} else {
					cacheKey = fmt.Sprintf("coords:%.4f:%.4f:%s", t.lat, t.lon, cfg.Units)
				}

				if hit, ok := weatherCache.Get(cacheKey); ok {
					results <- result{t: t, resp: hit, cache: true}
					continue
				}

				select {
				case <-ctx.Done():
					results <- result{t: t, err: ctx.Err()}
					continue
				case <-rateLimiter.C:
				}

				reqCtx, reqCancel := context.WithTimeout(ctx, cfg.Timeout)
				
				var resp *models.WeatherResponse
				var fetchErr error
				if t.byCity {
					resp, fetchErr = weatherClient.FetchByCity(reqCtx, t.city, cfg.Units)
				} else {
					resp, fetchErr = weatherClient.FetchByCoords(reqCtx, t.lat, t.lon, cfg.Units)
				}
				reqCancel()

				if fetchErr == nil {
					weatherCache.Set(cacheKey, resp)
				}
				
				results <- result{t: t, resp: resp, err: fetchErr}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	hasErrors := false
	for res := range results {
		if res.err != nil {
			var details string
			if res.t.byCity {
				details = fmt.Sprintf("[%s] %s", res.t.city, res.err.Error())
			} else {
				details = fmt.Sprintf("[%.4f, %.4f] %s", res.t.lat, res.t.lon, res.err.Error())
			}
			display.PrintError(os.Stderr, fmt.Errorf(details))
			hasErrors = true
		} else {
			display.PrintWeather(os.Stdout, res.resp, cfg.Units)
			if res.cache {
				display.PrintCached(os.Stdout)
			}
		}
	}

	if hasErrors {
		return 1
	}
	return 0
}

func printUsage() {
	b := func(s string) string { return "\033[1m" + s + "\033[0m" }
	fmt.Fprintf(os.Stderr, `
%s
  Concurrent CLI for live weather data (OpenWeatherMap)

%s
  weather [flags]

%s
  -city      string   City name, or comma-separated list (e.g. "Paris,Tokyo")
  -lat       float    Latitude  (pair with -lon)
  -lon       float    Longitude (pair with -lat)
  -units     string   metric (default) | imperial | standard
  -no-cache           Skip cache
  -version            Print version

%s
  export OWM_API_KEY=your_key_here

  weather -city "New Delhi, London, Tokyo"
  weather -lat 51.5074 -lon -0.1278 -units imperial
  weather -city Tokyo -units standard -no-cache
`,
		b("AeroStat"),
		b("USAGE"),
		b("FLAGS"),
		b("EXAMPLES"),
	)
}
