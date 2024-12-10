import axios from 'axios';
import csv from 'csv-parser';
import { Kafka, Producer } from 'kafkajs';
import { Readable } from 'stream';
import { StormType } from './types';
import { exit } from 'process';

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

let producerInitialized = false;

// Initialize producer with a persistent connection
const initProducer = async () => {
  if (!producerInitialized) {
    try {
      console.log('Initializing Kafka producer...');
      await producer.connect();
      producerInitialized = true;
      console.log('Kafka producer connected.');
    } catch (error) {
      console.error('Failed to connect Kafka producer:', error);
      throw error;
    }
  }
};

// Disconnect producer gracefully
export const closeProducer = async () => {
  if (producerInitialized) {
    try {
      console.log('Disconnecting Kafka producer...');
      await producer.disconnect();
      producerInitialized = false;
      console.log('Kafka producer disconnected.');
    } catch (error) {
      console.error('Failed to disconnect Kafka producer:', error);
    }
  }
};

process.on('SIGINT', async () => {
  console.log('SIGINT received. Cleaning up...');
  await closeProducer();
  exit(0);
});

process.on('SIGTERM', async () => {
  console.log('SIGTERM received. Cleaning up...');
  await closeProducer();
  exit(0);
});

// Retry logic with exponential backoff
const retryWithBackoff = async (
  fn: () => Promise<void>,
  retries = 5,
  delay = 1000,
) => {
  try {
    await fn();
  } catch (error) {
    if (retries > 0) {
      console.error(`Retrying after ${delay}ms...`, error);
      await new Promise((res) => setTimeout(res, delay));
      return retryWithBackoff(fn, retries - 1, delay * 2);
    }
    console.error('Max retries reached. Operation failed.');
    throw error;
  }
};

// Message queue for local buffering
const messageQueue: StormReport[] = [];

// Add message to queue
const addToQueue = (message: StormReport) => {
  messageQueue.push(message);
};

// Process queued messages
const processQueue = async (topic: string) => {
  while (messageQueue.length > 0) {
    const message = messageQueue.shift();
    try {
      await producer.send({
        topic,
        messages: [{ value: JSON.stringify(message) }],
      });
    } catch (err) {
      console.error('Failed to send message, re-queuing:', err);
      messageQueue.unshift(message!); // Re-queue the message
      break;
    }
  }
};

// Publish data to Kafka with retry and queue processing
const publishToKafka = async (data: StormReport[], topic: string) => {
  await retryWithBackoff(async () => {
    await initProducer();
    for (const report of data) {
      const message = JSON.stringify(report);
      const messageSize = Buffer.byteLength(message, 'utf-8');

      if (messageSize > 209715200) {
        console.error(`Message size exceeds limit: ${messageSize} bytes`);
        continue; // Skip oversized messages
      }

      addToQueue(report);
    }
    await processQueue(topic);
  });
};

// Fetch storm reports
export const fetchStormReports = async (
  url: string,
  date?: string,
): Promise<StormReport[]> => {
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
};

// Generate storms for testing
export const generateStormsToday = () => {
  const types = [StormType.TORNADO, StormType.HAIL, StormType.WIND];
  const fScales = ['F0', 'F1', 'F2', 'F3', 'F4', 'F5'];
  const locations = [
    'Dallas, TX',
    'New York, NY',
    'Chicago, IL',
    'Miami, FL',
    'Seattle, WA',
    'Denver, CO',
    'Los Angeles, CA',
    'Atlanta, GA',
    'Phoenix, AZ',
    'Boston, MA',
  ];

  const getRandomInRange = (min: number, max: number): number =>
    Math.random() * (max - min) + min;

  const stormCount = Math.floor(getRandomInRange(5, 10));
  const storms: StormReport[] = [];

  for (let i = 0; i < stormCount; i++) {
    const type = types[Math.floor(Math.random() * types.length)];
    const location = locations[Math.floor(Math.random() * locations.length)];
    const [city, state] = location.split(', ');

    const lat = getRandomInRange(25, 49.5).toString();
    const lon = getRandomInRange(-125, -67).toString();

    const timestamp = Math.floor(Date.now() / 1000).toString();

    const storm: StormReport = {
      lat,
      lon,
      location: city,
      type,
      time: '1200',
      date: timestamp,
      comments: 'Generated storm',
      county: 'Some County',
      state: state,
    };

    if (type === StormType.TORNADO) {
      storm.fScale = fScales[Math.floor(Math.random() * fScales.length)];
    } else if (type === StormType.HAIL) {
      storm.size = parseFloat(getRandomInRange(0.5, 4).toFixed(1)).toString();
    } else if (type === StormType.WIND) {
      storm.speed = Math.floor(getRandomInRange(30, 100)).toString();
    }

    storms.push(storm);
  }
  console.log('Generated storms:', storms);
  publishToKafka(storms, 'raw-weather-reports');
};

// Run producer
export const runProducer = async () => {
  const topic = 'raw-weather-reports';
  const baseURL = 'https://www.spc.noaa.gov/climo/reports/';
  const tornadoURL = baseURL + 'today_torn.csv';
  const hailURL = baseURL + 'today_hail.csv';
  const windURL = baseURL + 'today_wind.csv';

  try {
    console.log('Fetching storm reports...');
    const [tornados, hail, wind] = await Promise.all([
      fetchStormReports(tornadoURL),
      fetchStormReports(hailURL),
      fetchStormReports(windURL),
    ]);
    const allReports = [
      ...tornados.map((report) => ({ ...report, type: StormType.TORNADO })),
      ...hail.map((report) => ({ ...report, type: StormType.HAIL })),
      ...wind.map((report) => ({ ...report, type: StormType.WIND })),
    ];
    console.log(`Fetched ${allReports.length} reports. Publishing to Kafka...`);

    await publishToKafka(allReports, topic);
    console.log('Published storm reports to Kafka.');
  } catch (error) {
    console.error('Error in producer:', error);
    await closeProducer();
  }
};

// Interval for daily producer run, skip if running tests
if (process.env.NODE_ENV !== 'test') {
  setInterval(runProducer, 24 * 60 * 60 * 1000);
}

// CLI Arguments
if (require.main === module) {
  const arg = process.argv[2];
  if (arg === 'runProducer') {
    runProducer();
  }
  if (arg === 'generateStorms') {
    generateStormsToday();
  }
}
