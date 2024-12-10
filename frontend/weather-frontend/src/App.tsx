import React, { useEffect, useRef, useState } from "react";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers";
import TextField from "@mui/material/TextField";
import CircularProgress from "@mui/material/CircularProgress";
import Backdrop from "@mui/material/Backdrop";

export enum StormType {
  TORNADO = 'tornado',
  HAIL = 'hail',
  WIND = 'wind',
}


interface StormData {
  lat: number;
  lon: number;
  size?: number;
  fScale?: string;
  speed?: number;
  location: string;
  type: string;
  time: string;
  date: string;
  comments: string;
}

const generateInfoWindow = (storm: StormData): string => {
  let additionalDetail = "";

  if (storm.type === "hail") {
    additionalDetail = `<strong>Size:</strong> ${storm.size || "UNK"}`;
  } else if (storm.type === "tornado") {
    additionalDetail = `<strong>F-Scale:</strong> ${storm.fScale || "UNK"}`;
  } else if (storm.type === "wind") {
    additionalDetail = `<strong>Speed:</strong> ${storm.speed || "UNK"}`;
  }

  return `
    <div>
    <strong>Time:</strong> ${storm.time || "N/A"}<br />
      <strong>Location:</strong> ${storm.location}<br />
      <strong>Type:</strong> ${storm.type}<br />
      ${additionalDetail ? additionalDetail + "<br />" : "<br />"}
      <strong>Notes:</strong> ${storm.comments || "N/A"}
    </div>
  `;
};

const App: React.FC = () => {
  const mapRef = useRef<HTMLDivElement>(null);
  const [error, setError] = useState("");
  const [selectedDate, setSelectedDate] = useState<Date | null>(new Date());
  const [loading, setLoading] = useState(false);
  const typeColors = {
    tornado: "#FF0000",
    hail: "#0000FF",
    wind: "#00FF00",
  };


  const fetchAndRenderStormData = async (date: Date | null) => {
    if (!mapRef.current || !date) return;

    const map = new google.maps.Map(mapRef.current, {
      center: { lat: 39.8283, lng: -98.5795 },
      zoom: 4,
    });

    try {
      setLoading(true); // Start loading
      const startOfDay = new Date(date);
      startOfDay.setHours(0, 0, 0, 0);
      const unixTimestamp = Math.floor(startOfDay.getTime() / 1000);
      const response = await fetch(`/api/messages?date=${unixTimestamp}`);

      switch (response.status) {
        case 200:
          break;
        case 400:
          throw new Error("Bad Request: The server could not understand your request.");
        case 404:
          throw new Error("Not Found: No storm data available for the given date.");
        case 500:
          throw new Error("Internal Server Error: Please try again later.");
        default:
          throw new Error(`Unexpected Error: ${response.status} ${response.statusText}`);
      }

      const storms: StormData[] = await response.json();
      const bounds = new google.maps.LatLngBounds();

      storms.forEach((storm) => {
        const latitude = storm.lat;
        const longitude = storm.lon;
        bounds.extend(new google.maps.LatLng(latitude, longitude));

        if (!isNaN(latitude) && !isNaN(longitude)) {
          // const marker = new google.maps.Marker({
          //   position: { lat: latitude, lng: longitude },
          //   map: map,
          // });
          const circleColor: string = typeColors[storm.type as StormType] || "#CCCCCC";
          const marker = new google.maps.Circle({
            center: { lat: latitude, lng: longitude },
            radius: 25000,
            strokeColor: "#000000",
            strokeOpacity: 1,
            strokeWeight: 1,
            fillColor: circleColor,
            fillOpacity: 0.55,
            map: map,
          });

          const infoWindow = new google.maps.InfoWindow({
            content: generateInfoWindow(storm),
          });

          marker.addListener("click", () => {
            infoWindow.setPosition({ lat: latitude, lng: longitude });
            infoWindow.open(map);
          });
        }
      });

      map.fitBounds(bounds);
    } catch (error: unknown) {
      console.error("Error fetching or processing storm data:", error);
      setError((error as Error).message);
    } finally {
      setLoading(false); // Stop loading
    }
  };

  useEffect(() => {
    fetchAndRenderStormData(selectedDate);
  }, [selectedDate]);

  const onDateChange = (date: Date | null) => {
    setError("");
    setSelectedDate(date);
  };

  return (
    <div>
      <h1>Storm Data Visualization</h1>
      <LocalizationProvider dateAdapter={AdapterDateFns}>
        <DatePicker
          label="Select Date"
          value={selectedDate}
          onChange={(newDate) => onDateChange(newDate)}
          renderInput={(params) => <TextField {...params} />}
        />
      </LocalizationProvider>
      {error && <div style={{ color: "red" }}>{error}</div>}
      <Backdrop
        sx={{ color: "#fff", zIndex: (theme) => theme.zIndex.drawer + 1 }}
        open={loading}
      >
        <CircularProgress color="inherit" />
      </Backdrop>
      <div
        ref={mapRef}
        style={{ width: "100%", height: "90vh", position: "relative" }}
      ></div>
    </div>
  );
};

export default App;
