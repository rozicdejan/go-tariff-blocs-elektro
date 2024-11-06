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
        * {
            box-sizing: border-box;
        }
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
            z-index: 2;
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
        .zone-label {
            font-size: 1em;
            margin-top: 5px;
            font-weight: bold;
            color: #333;
        }
        .current-zone {
            font-size: 2em;
            color: #1976D2;
        }
        .hour-markers {
            position: absolute;
            width: 100%;
            height: 100%;
            pointer-events: none;
            z-index: 4;
        }
        .hour {
            position: absolute;
            font-size: 0.85em;
            font-weight: bold;
            color: #333;
            transform: translate(-50%, -50%);
        }
        .divider-lines {
            position: absolute;
            width: 100%;
            height: 100%;
            pointer-events: none;
            z-index: 3;
        }
        .line {
            position: absolute;
            width: 1px;
            height: 50%;
            background-color: #ccc;
            top: 50%;
            left: 50%;
            transform-origin: 0% 100%;
        }
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
        .legend {
            display: flex;
            align-items: center;
            margin-top: 20px;
            font-size: 0.9em;
            color: #666;
        }
        .legend-item {
            display: flex;
            align-items: center;
            margin-right: 10px;
        }
        .legend-color {
            width: 20px;
            height: 20px;
            border-radius: 4px;
            margin-right: 5px;
        }
        .remaining-time {
            margin-top: 10px;
            font-size: 1em;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="circle-container" id="circleContainer">
        <!-- Segments will be generated here by JavaScript -->
        <div class="center-circle">
            <div class="current-zone" id="currentZoneNumber">0</div>
            <div class="center-text" id="centerText">Loading...</div>
        </div>
        <div class="hour-markers" id="hourMarkers">
            <!-- Hour markers will be generated here by JavaScript -->
        </div>
        <div class="divider-lines" id="dividerLines">
            <!-- Divider lines between hours will be generated here by JavaScript -->
        </div>
        <div class="arrow" id="timeArrow"></div> <!-- Arrow indicating the current time -->
    </div>
    
    <div class="remaining-time" id="remainingTime">Calculating time until next block...</div>
    
    <!-- Legend for color-coded zones -->
    <div class="legend">
        <div class="legend-item">
            <div class="legend-color" style="background-color: #009688;"></div> 5 (Lowest)
        </div>
        <div class="legend-item">
            <div class="legend-color" style="background-color: #4DB6AC;"></div> 4
        </div>
        <div class="legend-item">
            <div class="legend-color" style="background-color: #90A4AE;"></div> 3
        </div>
        <div class="legend-item">
            <div class="legend-color" style="background-color: #1976D2;"></div> 2
        </div>
        <div class="legend-item">
            <div class="legend-color" style="background-color: #0D47A1;"></div> 1 (Highest)
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
                segment.style.transform = "rotate(" + (i * 15 - 45) + "deg)"; // Offset by -90° to align 0 at the top
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

            if ((hour >= 7 && hour < 14) || (hour >= 16 && hour < 20)) {
                if (isHighSeason && !isWeekend) {
                    return { zone: 1, label: "High Season, Weekday" };
                }
                return { zone: 1, label: "Zone 1" };
            } else if (hour === 6 || (hour >= 14 && hour < 16) || (hour >= 20 && hour < 22)) {
                return { zone: 2, label: "Zone 2" };
            } else if ((hour >= 0 && hour < 6) || (hour >= 22 && hour < 24)) {
                return { zone: 3, label: "Zone 3" };
            }
            return { zone: 5, label: "Zone 5" };
        }

        function createHourMarkers() {
            const markersContainer = document.getElementById("hourMarkers");
            const radius = 100; // Radius to place markers closer to the center

            for (let i = 0; i < 24; i++) {
                const angle = (i * 15 - 90) * (Math.PI / 180); // Offset -90° for 0 at the top
                const x = radius * Math.cos(angle) + 160; // Center + offset
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
            const rotation = (totalMinutes / 1440) * 360 - 90; // 1440 mins in a day, -90 to align 0 at the top
            document.getElementById("timeArrow").style.transform = "rotate(" + rotation + "deg)";
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
                zoneChangeHours = [0, 6, 14, 22];
            } else if (!isHighSeason && !isWeekend) {
                zoneChangeHours = [7, 14, 16, 20, 22];
            } else {
                zoneChangeHours = [0, 6, 14, 22];
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

            const remainingHours = nextChangeHour - hour;
            const totalMinutesUntilNextChange = remainingHours * 60 - minutes;
            const hoursUntilNextChange = Math.floor(totalMinutesUntilNextChange / 60);
            const minutesUntilNextChange = totalMinutesUntilNextChange % 60;

            document.getElementById("remainingTime").innerText = 
                "Remaining block time : " + hoursUntilNextChange + "h:" + minutesUntilNextChange + "m";
        }

        function highlightCurrentZone() {
            const now = new Date();
            const hour = now.getHours();
            const zoneInfo = getZoneForHour(hour);

           
            document.getElementById("currentZoneNumber").innerText = zoneInfo.zone; // Display the zone number in the center
            updateCenterText();
           
            calculateRemainingTime();
        }

        function updateCenterText() {
            const now = new Date();
            const isWeekend = now.getDay() === 0 || now.getDay() === 6;
            const month = now.getMonth();
            const isHighSeason = month === 10 || month === 11 || month === 0 || month === 1;

            let dayType = isWeekend ? "Free Day" : "Working Day";
            let seasonType = isHighSeason ? "High Season" : "Not High Season";

            document.getElementById("centerText").innerText = dayType + "\n" + seasonType;
        }

        createSegments();
        createHourMarkers();
        
        highlightCurrentZone();
        setInterval(highlightCurrentZone, 60000); // Update every minute
    </script>
</body>
</html>
`

type TariffData struct {
	Zone  int    `json:"zone"`
	Label string `json:"label"`
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
	zone, label := getTariffZone(now)
	data := TariffData{Zone: zone, Label: label}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func getTariffZone(t time.Time) (int, string) {
	hour := t.Hour()
	isHighSeason := t.Month() == time.November || t.Month() == time.December || t.Month() == time.January || t.Month() == time.February
	isWeekend := t.Weekday() == time.Saturday || t.Weekday() == time.Sunday

	switch {
	case (hour >= 7 && hour < 14) || (hour >= 16 && hour < 20):
		if isHighSeason && !isWeekend {
			return 1, "Zone 1 (High Season Weekday)"
		}
		return 1, "Zone 1"
	case hour == 6 || (hour >= 14 && hour < 16) || (hour >= 20 && hour < 22):
		return 2, "Zone 2"
	case (hour >= 0 && hour < 6) || (hour >= 22 && hour < 24):
		return 3, "Zone 3"
	default:
		return 5, "Zone 5"
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/api/tariff", apiHandler)
	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
