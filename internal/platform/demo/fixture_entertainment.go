package demo

import (
	"net/http"
	"strings"
	"time"
)

// DemoSteamID is the example Steam account whose wishlist the game deals
// widget shows.
const DemoSteamID = "76561198000000042"

func init() {
	handle(hostIs("ll.thespacedevs.com"), func(_ *http.Request, now time.Time) (int, string, string) {
		return jsonOK(launches(now))
	})
	handle(hostIs("api.nasa.gov"), func(_ *http.Request, now time.Time) (int, string, string) {
		return jsonOK(mustJSON(map[string]string{
			"date": now.Format("2006-01-02"), "title": "Aurora over the Wadden Sea", "copyright": "Example Photographer",
			"explanation": "On a clear November night the northern lights reached far south. Green curtains rose over the mud flats, and the lights of a ferry drew a line along the horizon.",
			"media_type":  "image", "url": "https://apod.nasa.gov/apod/image/demo/aurora_1024.jpg",
			"hdurl": "https://apod.nasa.gov/apod/image/demo/aurora_2048.jpg",
		}))
	})
	handle(hostIs("apod.nasa.gov"), func(_ *http.Request, _ time.Time) (int, string, string) {
		return http.StatusOK, "image/png", aurora(640, 400)
	})
	handle(hostIs("store-site-backend-static.ak.epicgames.com"), func(_ *http.Request, now time.Time) (int, string, string) {
		return jsonOK(epicFreeGames(now))
	})
	handle(hostIs("api.steampowered.com"), func(req *http.Request, now time.Time) (int, string, string) {
		if strings.Contains(req.URL.Path, "GetWishlist") {
			return jsonOK(`{"response":{"items":[{"appid":1001},{"appid":1002},{"appid":1003},{"appid":1004}]}}`)
		}
		return jsonOK(steamItems(now))
	})
}

func launches(now time.Time) string {
	type named struct {
		Name string `json:"name"`
	}
	type launch struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Net    string `json:"net"`
		Status struct {
			Abbrev string `json:"abbrev"`
		} `json:"status"`
		Rocket struct {
			Configuration named `json:"configuration"`
		} `json:"rocket"`
		Mission struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"mission"`
		Provider named `json:"launch_service_provider"`
		Pad      struct {
			Name     string `json:"name"`
			Location named  `json:"location"`
		} `json:"pad"`
		WebcastLive bool `json:"webcast_live"`
	}
	mk := func(id, name, rocket, provider, pad, place, mission, kind, status string, in time.Duration) launch {
		var l launch
		l.ID, l.Name, l.Net = id, name, now.Add(in).UTC().Format(time.RFC3339)
		l.Status.Abbrev = status
		l.Rocket.Configuration.Name = rocket
		l.Provider.Name = provider
		l.Pad.Name, l.Pad.Location.Name = pad, place
		l.Mission.Name, l.Mission.Type = mission, kind
		l.Mission.Description = "Example mission of the Waylight demo."
		return l
	}
	results := []launch{
		mk("6a5e1c2f-0000-4000-8000-000000000001", "Falcon 9 | Starlink Group 12-7", "Falcon 9 Block 5", "SpaceX",
			"SLC-40", "Cape Canaveral, FL, USA", "Starlink Group 12-7", "Communications", "Go", 5*time.Hour),
		mk("6a5e1c2f-0000-4000-8000-000000000002", "Ariane 6 | Galileo L14", "Ariane 62", "Arianespace",
			"ELA-4", "Kourou, French Guiana", "Galileo L14", "Navigation", "TBC", 52*time.Hour),
		mk("6a5e1c2f-0000-4000-8000-000000000003", "Electron | Ocean Watch", "Electron", "Rocket Lab",
			"LC-1B", "Mahia Peninsula, New Zealand", "Ocean Watch", "Earth Science", "Go", 4*24*time.Hour),
		mk("6a5e1c2f-0000-4000-8000-000000000004", "Starship | Flight 16", "Starship", "SpaceX",
			"OLP-1", "Starbase, TX, USA", "Flight 16", "Test Flight", "TBD", 11*24*time.Hour),
	}
	return mustJSON(map[string]any{"count": len(results), "results": results})
}

func epicFreeGames(now time.Time) string {
	offer := func(from, to time.Duration) []map[string]any {
		return []map[string]any{{"promotionalOffers": []map[string]any{{
			"startDate": now.Add(from).UTC().Format(time.RFC3339), "endDate": now.Add(to).UTC().Format(time.RFC3339),
			"discountSetting": map[string]int{"discountPercentage": 0},
		}}}}
	}
	el := func(title, slug string, current bool, from, to time.Duration) map[string]any {
		promo := map[string]any{"promotionalOffers": []any{}, "upcomingPromotionalOffers": []any{}}
		if current {
			promo["promotionalOffers"] = offer(from, to)
		} else {
			promo["upcomingPromotionalOffers"] = offer(from, to)
		}
		return map[string]any{"title": title, "productSlug": slug, "promotions": promo}
	}
	day := 24 * time.Hour
	elements := []any{
		el("Lighthouse Builder", "lighthouse-builder", true, -2*day, 5*day),
		el("Tidal Racers", "tidal-racers", true, -2*day, 5*day),
		el("Moorland Mysteries", "moorland-mysteries", false, 5*day, 12*day),
	}
	return mustJSON(map[string]any{"data": map[string]any{"Catalog": map[string]any{"searchStore": map[string]any{"elements": elements}}}})
}

func steamItems(now time.Time) string {
	end := now.Add(3 * 24 * time.Hour).Unix()
	item := func(id int, name, final, finalCents, origCents string, pct int) map[string]any {
		opt := map[string]any{"formatted_final_price": final, "final_price_in_cents": finalCents,
			"original_price_in_cents": origCents, "discount_pct": pct}
		if pct > 0 {
			opt["active_discounts"] = []map[string]int64{{"discount_end_date": end}}
		}
		return map[string]any{"appid": id, "name": name, "success": 1, "best_purchase_option": opt}
	}
	items := []any{
		item(1001, "Harbour Master Simulator", "9,99€", "999", "2499", 60),
		item(1002, "Northern Rails", "14,79€", "1479", "1999", 26),
		item(1003, "Skyline Gardens", "4,49€", "449", "1799", 75),
		item(1004, "Deep Sea Survey", "29,99€", "2999", "2999", 0),
	}
	return mustJSON(map[string]any{"response": map[string]any{"store_items": items}})
}
