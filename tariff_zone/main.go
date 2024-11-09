package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// HTML template for the circular tariff display
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tariff Zones</title>
    <style>
        * { box-sizing: border-box; }
        body {
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            height: 100vh;
            background-color: #f4f6f9;
            font-family: Arial, sans-serif;
            margin: 0;
            color: #333;
        }
        .circle-container {
            position: relative;
            width: 320px;
            height: 320px;
            border-radius: 50%;
            overflow: hidden;
        }
        .segment {
            position: absolute;
            width: 100%;
            height: 100%;
            clip-path: polygon(50% 50%, 100% 0%, 100% 100%);
            transform-origin: 50% 50%;
        }
        .center-circle {
            position: absolute;
            top: 15%;
            left: 15%;
            width: 70%;
            height: 70%;
            background-color: white;
            border-radius: 50%;
            z-index: 4;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            font-size: 1.2em;
        }
        .center-text {
            text-align: center;
            font-weight: bold;
            color: #333;
        }
        .zone-label { font-size: 1em; margin-top: 5px; font-weight: bold; color: #333; }
        .current-zone { font-size: 2em; color: #1976D2; }
        .hour-markers { position: absolute; width: 100%; height: 100%; pointer-events: none; z-index: 4; }
        .divider-lines { position: absolute; width: 100%; height: 100%; pointer-events: none; z-index: 0; }
        .hour { position: absolute; font-size: 0.85em; font-weight: bold; color: #333; transform: translate(-50%, -50%); }
        .line { position: absolute; width: 1px; height: 50%; background-color: #ccc; top: 50%; left: 50%; transform-origin: 0% 100%; }
        .arrow {
            position: absolute;
            width: 3px;
            height: 60px;
            background-color: red;
            top: 50%;
            left: 50%;
            transform-origin: 50% 100%;
            z-index: 0;
            border-radius: 3px;
        }
        .legend { display: flex; align-items: center; margin-top: 20px; font-size: 0.9em; color: #666; }
        .legend-item { display: flex; align-items: center; margin-right: 10px; }
        .legend-color { width: 20px; height: 20px; border-radius: 4px; margin-right: 5px; }
        .remaining-time { margin-top: 10px; font-size: 1em; color: #666; }
        .current-time-indicator {
            position: absolute;
            width: 4px;
            height: 165px;
            background-color: rgba(255, 255, 255, 0.8); /* Transparent white */
            top: 0;
            left: 50%;
            transform-origin: 50% 100%;
            z-index: 3;
            border-radius: 2px;
        }
    </style>
</head>
<body>
    <div class="circle-container" id="circleContainer">
        <div class="center-circle">
            <div class="current-zone" id="currentZoneNumber">0</div>
            <div class="center-text" id="centerText">Loading...</div>
        </div>
        <div class="hour-markers" id="hourMarkers"></div>
        <div class="divider-lines" id="dividerLines"></div>
        <div class="arrow" id="timeArrow"></div>
        <div class="current-time-indicator" id="currentTimeIndicator"></div> <!-- Current time indicator -->
    </div>
    
    <div class="remaining-time" id="remainingTime">Calculating time until next block...</div>
    
    <div class="legend">
        <div class="legend-item">
            <div class="legend-color" style="background-color: #009688;"></div> Block 5
        </div>
        <div class="legend-item">
            <div class="legend-color" style="background-color: #4DB6AC;"></div> Block 4
        </div>
        <div class="legend-item">
            <div class="legend-color" style="background-color: #90A4AE;"></div> Block 3
        </div>
        <div class="legend-item">
            <div class="legend-color" style="background-color: #1976D2;"></div> Block 2
        </div>
        <div class="legend-item">
            <div class="legend-color" style="background-color: #0D47A1;"></div> Block 1
        </div>
    </div>

    <script>
        const colors = {
            zone1: '#0D47A1',
            zone2: '#1976D2',
            zone3: '#90A4AE',
            zone4: '#4DB6AC',
            zone5: '#009688'
        };

        function createSegments() {
            const container = document.getElementById("circleContainer");
            for (let i = 0; i < 24; i++) {
                const segment = document.createElement("div");
                segment.className = "segment";
                segment.style.backgroundColor = getColorForHour(i);
                segment.style.transform = "rotate(" + (i * 15 - 45) + "deg)";
                container.appendChild(segment);
            }
        }

        function getColorForHour(hour) {
            const zoneInfo = getZoneForHour(hour);
            return colors["zone" + zoneInfo.zone];
        }

        function getZoneForHour(hour) {
            const now = new Date();
            const month = now.getMonth();
            const isHighSeason = month === 10 || month === 11 || month === 0 || month === 1;
            const isWeekend = now.getDay() === 0 || now.getDay() === 6;

            if (isHighSeason && isWeekend) {
                if (hour >= 22 || hour < 6) return { zone: 4, label: "Zone 4 (High Season Non-Working Day)" };
                else if (hour >= 6 && hour < 7) return { zone: 3, label: "Zone 3 (High Season Non-Working Day)" };
                else if (hour >= 7 && hour < 14) return { zone: 2, label: "Zone 2 (High Season Non-Working Day)" };
                else if (hour >= 14 && hour < 16) return { zone: 3, label: "Zone 3 (High Season Non-Working Day)" };
                else if (hour >= 16 && hour < 20) return { zone: 2, label: "Zone 2 (High Season Non-Working Day)" };
                else if (hour >= 20 && hour < 22) return { zone: 3, label: "Zone 3 (High Season Non-Working Day)" };
            }

            if (isHighSeason && !isWeekend) {
                if (hour >= 22 || hour < 6) return { zone: 3, label: "Zone 3 (High Season Working Day)" };
                else if (hour >= 6 && hour < 7) return { zone: 2, label: "Zone 2 (High Season Working Day)" };
                else if (hour >= 7 && hour < 14) return { zone: 1, label: "Zone 1 (High Season Working Day)" };
                else if (hour >= 14 && hour < 16) return { zone: 2, label: "Zone 2 (High Season Working Day)" };
                else if (hour >= 16 && hour < 20) return { zone: 1, label: "Zone 1 (High Season Working Day)" };
                else if (hour >= 20 && hour < 22) return { zone: 2, label: "Zone 2 (High Season Working Day)" };
            }

            if (!isHighSeason && !isWeekend) {
                if (hour >= 22 || hour < 6) return { zone: 4, label: "Zone 4 (Low Season Working Day)" };
                else if (hour >= 6 && hour < 7) return { zone: 3, label: "Zone 3 (Low Season Working Day)" };
                else if (hour >= 7 && hour < 14) return { zone: 2, label: "Zone 2 (Low Season Working Day)" };
                else if (hour >= 14 && hour < 16) return { zone: 3, label: "Zone 3 (Low Season Working Day)" };
                else if (hour >= 16 && hour < 20) return { zone: 2, label: "Zone 2 (Low Season Working Day)" };
                else if (hour >= 20 && hour < 22) return { zone: 3, label: "Zone 3 (Low Season Working Day)" };
            }

            if (!isHighSeason && isWeekend) {
                if (hour >= 22 || hour < 6) return { zone: 5, label: "Zone 5 (Low Season Non-Working Day)" };
                else if (hour >= 6 && hour < 7) return { zone: 4, label: "Zone 4 (Low Season Non-Working Day)" };
                else if (hour >= 7 && hour < 14) return { zone: 1, label: "Zone 1 (Low Season Non-Working Day)" };
                else if (hour >= 14 && hour < 16) return { zone: 2, label: "Zone 2 (Low Season Non-Working Day)" };
                else if (hour >= 16 && hour < 20) return { zone: 1, label: "Zone 1 (Low Season Non-Working Day)" };
                else if (hour >= 20 && hour < 22) return { zone: 2, label: "Zone 2 (Low Season Non-Working Day)" };
            }

            return { zone: 3, label: "Zone 3 (Default)" };
        }

        function createHourMarkers() {
            const markersContainer = document.getElementById("hourMarkers");
            const radius = 100;
            for (let i = 0; i < 24; i++) {
                const angle = (i * 15 - 90) * (Math.PI / 180);
                const x = radius * Math.cos(angle) + 160;
                const y = radius * Math.sin(angle) + 160;
                const marker = document.createElement("div");
                marker.className = "hour";
                marker.style.left = x + "px";
                marker.style.top = y + "px";
                marker.innerText = i;
                markersContainer.appendChild(marker);
            }
        }

        function createDividerLines() {
            const linesContainer = document.getElementById("dividerLines");
            for (let i = 0; i < 24; i++) {
                const line = document.createElement("div");
                line.className = "line";
                line.style.transform = "rotate(" + (i * 15) + "deg)";
                linesContainer.appendChild(line);
            }
        }

        function updateArrow() {
            const now = new Date();
            const hours = now.getHours();
            const minutes = now.getMinutes();
            const totalMinutes = hours * 60 + minutes;
            const rotation = (totalMinutes / 1440) * 360 - 90;
            document.getElementById("timeArrow").style.transform = "rotate(" + rotation + "deg)";
        }

        function updateCurrentTimeIndicator() {
            const now = new Date();
            const hours = now.getHours();
            const minutes = now.getMinutes();
            const totalMinutes = hours * 60 + minutes;
            const rotation = (totalMinutes / 1440) * 360;
            document.getElementById("currentTimeIndicator").style.transform = "rotate(" + rotation + "deg)";
        }

        function calculateRemainingTime() {
            const now = new Date();
            const hour = now.getHours();
            const minutes = now.getMinutes();
            const month = now.getMonth();
            const isHighSeason = month === 10 || month === 11 || month === 0 || month === 1;
            const isWeekend = now.getDay() === 0 || now.getDay() === 6;
            let zoneChangeHours = [];
            if (isHighSeason && !isWeekend) {
                zoneChangeHours = [6, 7, 14, 16, 20, 22];
            } else if (isHighSeason && isWeekend) {
                zoneChangeHours = [0, 6, 14, 16, 20, 22];
            } else if (!isHighSeason && !isWeekend) {
                zoneChangeHours = [7, 14, 16, 20, 22];
            } else {
                zoneChangeHours = [0, 6, 14, 16, 20, 22];
            }

            let nextChangeHour = null;
            for (let i = 0; i < zoneChangeHours.length; i++) {
                if (hour < zoneChangeHours[i] || (hour === zoneChangeHours[i] && minutes === 0)) {
                    nextChangeHour = zoneChangeHours[i];
                    break;
                }
            }

            if (nextChangeHour === null) {
                nextChangeHour = zoneChangeHours[0] + 24;
            }

            const remainingMinutes = ((nextChangeHour * 60 - (hour * 60 + minutes)) + 1440) % 1440;
            const hoursUntilNextChange = Math.floor(remainingMinutes / 60);
            const minutesUntilNextChange = remainingMinutes % 60;

            document.getElementById("remainingTime").innerText = 
                "Naslednji blok čez : " + hoursUntilNextChange + "h:" + minutesUntilNextChange + "m";
        }

        function highlightCurrentZone() {
            const now = new Date();
            const hour = now.getHours();
            const zoneInfo = getZoneForHour(hour);
            document.getElementById("currentZoneNumber").innerText = zoneInfo.zone;
            updateCenterText();
            calculateRemainingTime();
            updateCurrentTimeIndicator();
        }

        function updateCenterText() {
            const now = new Date();
            const isWeekend = now.getDay() === 0 || now.getDay() === 6;
            const month = now.getMonth();
            const isHighSeason = month === 10 || month === 11 || month === 0 || month === 1;
            let dayType = isWeekend ? "Dela prosta dan" : "Delovni Dan";
            let seasonType = isHighSeason ? "Višja sezona" : "Nizka Sezona";
            document.getElementById("centerText").innerText = dayType + "\n" + seasonType;
        }

        createSegments();
        createHourMarkers();
        createDividerLines();
        highlightCurrentZone();
        setInterval(highlightCurrentZone, 60000);
    </script>
</body>
</html>
`

type TariffData struct {
	Zone               int    `json:"zone"`
	Label              string `json:"label"`
	RemainingBlockTime string `json:"remaining_block_time"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("circle").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	zone, label, remainingBlockTime := getTariffZone(now)
	data := TariffData{
		Zone:               zone,
		Label:              label,
		RemainingBlockTime: remainingBlockTime,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func getTariffZone(t time.Time) (int, string, string) {
	hour := t.Hour()
	minute := t.Minute()
	isHighSeason := t.Month() == time.November || t.Month() == time.December || t.Month() == time.January || t.Month() == time.February
	isWeekend := t.Weekday() == time.Saturday || t.Weekday() == time.Sunday

	var zone int
	var label string
	if isHighSeason && isWeekend {
		switch {
		case hour >= 22 || hour < 6:
			zone = 4
			label = "Zone 4 (High Season Non-Working Day)"
		case hour >= 6 && hour < 7:
			zone = 3
			label = "Zone 3 (High Season Non-Working Day)"
		case hour >= 7 && hour < 14:
			zone = 2
			label = "Zone 2 (High Season Non-Working Day)"
		case hour >= 14 && hour < 16:
			zone = 3
			label = "Zone 3 (High Season Non-Working Day)"
		case hour >= 16 && hour < 20:
			zone = 2
			label = "Zone 2 (High Season Non-Working Day)"
		case hour >= 20 && hour < 22:
			zone = 3
			label = "Zone 3 (High Season Non-Working Day)"
		}
	} else if isHighSeason && !isWeekend {
		switch {
		case hour >= 22 || hour < 6:
			zone = 3
			label = "Zone 3 (High Season Working Day)"
		case hour >= 6 && hour < 7:
			zone = 2
			label = "Zone 2 (High Season Working Day)"
		case hour >= 7 && hour < 14:
			zone = 1
			label = "Zone 1 (High Season Working Day)"
		case hour >= 14 && hour < 16:
			zone = 2
			label = "Zone 2 (High Season Working Day)"
		case hour >= 16 && hour < 20:
			zone = 1
			label = "Zone 1 (High Season Working Day)"
		case hour >= 20 && hour < 22:
			zone = 2
			label = "Zone 2 (High Season Working Day)"
		}
	} else if !isHighSeason && !isWeekend {
		switch {
		case hour >= 22 || hour < 6:
			zone = 4
			label = "Zone 4 (Low Season Working Day)"
		case hour >= 6 && hour < 7:
			zone = 3
			label = "Zone 3 (Low Season Working Day)"
		case hour >= 7 && hour < 14:
			zone = 2
			label = "Zone 2 (Low Season Working Day)"
		case hour >= 14 && hour < 16:
			zone = 3
			label = "Zone 3 (Low Season Working Day)"
		case hour >= 16 && hour < 20:
			zone = 2
			label = "Zone 2 (Low Season Working Day)"
		case hour >= 20 && hour < 22:
			zone = 3
			label = "Zone 3 (Low Season Working Day)"
		}
	} else {
		switch {
		case hour >= 22 || hour < 6:
			zone = 5
			label = "Zone 5 (Low Season Non-Working Day)"
		case hour >= 6 && hour < 7:
			zone = 4
			label = "Zone 4 (Low Season Non-Working Day)"
		case hour >= 7 && hour < 14:
			zone = 1
			label = "Zone 1 (Low Season Non-Working Day)"
		case hour >= 14 && hour < 16:
			zone = 2
			label = "Zone 2 (Low Season Non-Working Day)"
		case hour >= 16 && hour < 20:
			zone = 1
			label = "Zone 1 (Low Season Non-Working Day)"
		case hour >= 20 && hour < 22:
			zone = 2
			label = "Zone 2 (Low Season Non-Working Day)"
		}
	}

	nextChangeHour := getNextChangeHour(hour, minute, isHighSeason, isWeekend)
	remainingMinutes := ((nextChangeHour*60 - (hour*60 + minute)) + 1440) % 1440
	remainingHours := remainingMinutes / 60
	remainingMinutes = remainingMinutes % 60
	remainingBlockTime := fmt.Sprintf("%dh:%dm", remainingHours, remainingMinutes)

	return zone, label, remainingBlockTime
}

func getNextChangeHour(hour, minute int, isHighSeason, isWeekend bool) int {
	var zoneChangeHours []int
	if isHighSeason && !isWeekend {
		zoneChangeHours = []int{6, 7, 14, 16, 20, 22}
	} else if isHighSeason && isWeekend {
		zoneChangeHours = []int{0, 6, 14, 16, 20, 22}
	} else if !isHighSeason && !isWeekend {
		zoneChangeHours = []int{7, 14, 16, 20, 22}
	} else {
		zoneChangeHours = []int{0, 6, 14, 16, 20, 22}
	}

	for _, changeHour := range zoneChangeHours {
		if hour < changeHour || (hour == changeHour && minute == 0) {
			return changeHour
		}
	}
	return zoneChangeHours[0] + 24
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/api/tariff", apiHandler)
	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
