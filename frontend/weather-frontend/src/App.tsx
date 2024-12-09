import React, { useEffect, useRef } from "react";

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

  useEffect(() => {
    const initMap = async () => {
      if (!mapRef.current) return;

      const map = new google.maps.Map(mapRef.current, {
        center: { lat: 39.8283, lng: -98.5795 },
        zoom: 4,
      });

      try {
        const response = await fetch("/api/messages");
        const storms: StormData[] = await response.json();
        const bounds = new google.maps.LatLngBounds();
        storms.forEach((storm) => {
          console.log("st", storm);
          const latitude = storm.lat;
          const longitude = storm.lon;
          bounds.extend(new google.maps.LatLng(storm.lat, storm.lon));
          const size = storm.size;
          console.log("ltlng", latitude, longitude)
          if (!isNaN(latitude) && !isNaN(longitude)) {
            const marker = new google.maps.Marker({
              position: { lat: latitude, lng: longitude },
              map: map,
            });

            const infoWindow = new google.maps.InfoWindow({
              content: generateInfoWindow(storm)
            });

            marker.addListener("click", () => {
              infoWindow.open(map, marker);
            });
          }
        });
        map.fitBounds(bounds);
      } catch (error) {
        console.error("Error fetching storm data:", error);
      }
    };

    initMap();
  }, []);

  return (
    <div>
      <h1>Storm Data Visualization</h1>
      <div
        ref={mapRef}
        style={{ width: "100%", height: "90vh", position: "relative" }}
      ></div>
    </div>
  );
};

export default App;
