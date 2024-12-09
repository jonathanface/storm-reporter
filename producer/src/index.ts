import axios from 'axios';
import csv from 'csv-parser';
import { Kafka, Producer } from 'kafkajs';
import { Readable } from 'stream';
import { StormType } from './types';

interface StormReport {
  [key: string]: string;
}

const kafka = new Kafka({
  brokers: ['kafka:9092'],
  requestTimeout: 30000,
  logLevel: 1,
  retry: {
    initialRetryTime: 300,
    retries: 10,
  },
});
const producer: Producer = kafka.producer({
  allowAutoTopicCreation: false,
  maxInFlightRequests: 5,
});

export const generateStormsToday = () => {
  const types = [StormType.TORNADO, StormType.HAIL, StormType.WIND];
  const fScales = ["F0", "F1", "F2", "F3", "F4", "F5"];
  const locations = [
    "Dallas, TX",
    "New York, NY",
    "Chicago, IL",
    "Miami, FL",
    "Seattle, WA",
    "Denver, CO",
    "Los Angeles, CA",
    "Atlanta, GA",
    "Phoenix, AZ",
    "Boston, MA",
  ];

  // Helper to generate random numbers within a range
  const getRandomInRange = (min: number, max: number): number =>
    Math.random() * (max - min) + min;

  // Generate between 5 and 10 storm reports
  const stormCount = Math.floor(getRandomInRange(5, 10));
  const storms: StormReport[] = [];

  for (let i = 0; i < stormCount; i++) {
    const type = types[Math.floor(Math.random() * types.length)];
    const location = locations[Math.floor(Math.random() * locations.length)];
    const [city, state] = location.split(", ");

    // Mock latitude and longitude roughly within US bounds
    const lat = getRandomInRange(25, 49.5).toString();
    const lon = getRandomInRange(-125, -67).toString();

    // Unix timestamp for the current time
    const timestamp = Math.floor(Date.now() / 1000);

    // Generate StormData with type-specific fields
    const storm: StormReport = {
      lat,
      lon,
      location: `${city}`,
      type,
      time: "1200",
      date: timestamp.toString(),
      comments: "it's really bad",
      county: "some county",
      state: state
    };

    // Populate type-specific fields
    if (type === StormType.TORNADO) {
      storm.fScale = fScales[Math.floor(Math.random() * fScales.length)];
    } else if (type === StormType.HAIL) {
      storm.size = parseFloat(getRandomInRange(0.5, 4).toFixed(1)).toString(); // Size in inches
    } else if (type === StormType.WIND) {
      storm.speed = Math.floor(getRandomInRange(30, 100)).toString(); // Speed in mph
    }
    storms.push(storm);
  }
  console.log("Storms generated:", storms);
  const topic = 'raw-weather-reports';
  publishToKafka(storms, topic)
}

export const fetchStormReports = async (url: string, date?: string): Promise<StormReport[]> => {
  const response = await axios.get<Readable>(url, { responseType: 'stream' });

  return new Promise((resolve, reject) => {
    const reports: StormReport[] = [];
    response?.data
      .pipe(csv())
      .on('data', (row: StormReport) => {
        const timestamp = date ? date : new Date().getTime().toString();
        row.date = timestamp;
        reports.push(row);
      })
      .on('end', () => resolve(reports))
      .on('error', (err) => reject(err));
  });
}

const MAX_MESSAGE_BYTES = 209715200; // 200 MB

const publishToKafka = async (data: StormReport[], topic: string): Promise<void> => {
  await producer.connect();
  // sending out to kafka one csv row at a time
  for (const report of data) {
    const message = JSON.stringify(report);
    const messageSize = Buffer.byteLength(message, 'utf-8');

    if (messageSize > MAX_MESSAGE_BYTES) {
      console.error(`Message size exceeds limit: ${messageSize} bytes`);
      continue; // Skip oversized messages
    }

    await producer.send({
      topic,
      messages: [{ value: message }],
    });
  }
  await producer.disconnect();
}

const baseURL = "https://www.spc.noaa.gov/climo/reports/"

export const runProducer = async () => {
  console.log('runProducer function has been invoked.');
  const topic = 'raw-weather-reports';
  const tornadoURL = baseURL + "today_torn.csv";
  const hailURL = baseURL + "today_hail.csv";
  const windURL = baseURL + "today_wind.csv";
  try {
    console.log('Fetching storm reports...');
    const [tornados, hail, wind] = await Promise.all([
      fetchStormReports(tornadoURL),
      fetchStormReports(hailURL),
      fetchStormReports(windURL)
    ]);
    const allReports = [
      ...tornados.map(report => ({ ...report, type: StormType.TORNADO })),
      ...hail.map(report => ({ ...report, type: StormType.HAIL })),
      ...wind.map(report => ({ ...report, type: StormType.WIND }))
    ];
    console.log(`Fetched ${allReports.length} reports. Publishing to Kafka...`);

    await publishToKafka(allReports, topic);
    console.log('Published storm reports to Kafka.');
  } catch (error) {
    console.error('Error in producer:', error);
  }
}
console.log('Producer service has started.');
runProducer();
setInterval(runProducer, 24 * 60 * 60 * 1000);

if (require.main === module) {
  const arg = process.argv[2];
  console.log(process.argv)
  if (arg === 'runProducer') {
    runProducer();
  }
  if (arg === 'generateStorms') {
    generateStormsToday();
  }
}