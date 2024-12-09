import React, { useEffect, useRef, useState } from "react";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers";
import TextField from "@mui/material/TextField";
import CircularProgress from "@mui/material/CircularProgress";
import Backdrop from "@mui/material/Backdrop";

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
}

const generateInfoWindow = (storm: StormData): string => {
  let additionalDetail = "";

  if (storm.type === "hail") {
    additionalDetail = `<strong>Size:</strong> ${storm.size}`;
  } else if (storm.type === "tornado") {
    additionalDetail = `<strong>F-Scale:</strong> ${storm.fScale}`;
  } else if (storm.type === "wind") {
    additionalDetail = `<strong>Speed:</strong> ${storm.speed}`;
  }

  return `
    <div>
      <strong>Location:</strong> ${storm.location}<br />
      <strong>Type:</strong> ${storm.type}<br />
      ${additionalDetail ? additionalDetail + "<br />" : ""}
      <strong>Time:</strong> ${storm.time || "N/A"}
    </div>
  `;
};

const App: React.FC = () => {
  const mapRef = useRef<HTMLDivElement>(null);
  const [error, setError] = useState("");
  const [selectedDate, setSelectedDate] = useState<Date | null>(new Date());
  const [loading, setLoading] = useState(false);

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
          const marker = new google.maps.Marker({
            position: { lat: latitude, lng: longitude },
            map: map,
          });

          const infoWindow = new google.maps.InfoWindow({
            content: generateInfoWindow(storm),
          });

          marker.addListener("click", () => {
            infoWindow.open(map, marker);
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
