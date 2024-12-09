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

export async function fetchStormReports(url: string, date?: string): Promise<StormReport[]> {
  const response = await axios.get<Readable>(url, { responseType: 'stream' });

  return new Promise((resolve, reject) => {
    const reports: StormReport[] = [];
    response?.data
      .pipe(csv())
      .on('data', (row: StormReport) => {
        const timestamp = date ? date : new Date().getTime().toString();
        console.log(date, "setting to", timestamp)
        row.date = timestamp;
        reports.push(row);
      })
      .on('end', () => resolve(reports))
      .on('error', (err) => reject(err));
  });
}

const MAX_MESSAGE_BYTES = 209715200; // 200 MB

async function publishToKafka(data: StormReport[], topic: string): Promise<void> {
  await producer.connect();
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

const formatTimestampToYYMMDD = (unixTimestamp: string): string => {
  const date = new Date(parseInt(unixTimestamp) * 1000);
  const year = date.getFullYear() % 100;
  const month = date.getMonth() + 1;
  const day = date.getDate();

  // Pad with leading zeroes if necessary
  const formattedYear = year.toString().padStart(2, '0');
  const formattedMonth = month.toString().padStart(2, '0');
  const formattedDay = day.toString().padStart(2, '0');

  return `${formattedYear}${formattedMonth}${formattedDay}`;
}

export async function runProducer(dateString?: string) {
  console.log('runProducer function has been invoked.');
  const topic = 'raw-weather-reports';
  dateString = "1716764227"
  const tornadoURL = dateString ? baseURL + formatTimestampToYYMMDD(dateString) + "_rpts_torn.csv" : baseURL + "today_torn.csv";
  const hailURL = dateString ? baseURL + formatTimestampToYYMMDD(dateString) + "_rpts_hail.csv" : baseURL + "today_hail.csv";
  const windURL = dateString ? baseURL + formatTimestampToYYMMDD(dateString) + "_rpts_wind.csv" : baseURL + "today_wind.csv";
  // const tornadoURL = baseURL + "240526_rpts_torn.csv";
  // const hailURL = baseURL + "240526_rpts_hail.csv";
  // const windURL = baseURL + "240526_rpts_wind.csv";
  try {
    console.log('Fetching storm reports...');
    const [tornados, hail, wind] = await Promise.all([
      fetchStormReports(tornadoURL, dateString),
      fetchStormReports(hailURL, dateString),
      fetchStormReports(windURL, dateString)
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
    if (process.argv[3]) {
      runProducer(process.argv[3]);
    } else {
      runProducer();
    }
  }
}