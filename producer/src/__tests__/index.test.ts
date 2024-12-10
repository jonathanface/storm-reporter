import { closeProducer, fetchStormReports } from '../index';
import axios from 'axios';
import { Readable } from 'stream';

jest.mock('axios');

afterEach(() => {
  jest.clearAllMocks(); // Reset mocks to avoid unexpected interactions
});

afterAll(async () => {
  jest.restoreAllMocks(); // Ensure no lingering mocks remain
  await closeProducer();
});

describe('fetchStormReports', () => {
  it('should fetch and parse storm reports', async () => {
    const mockCsvData = `location,date,type\nDenver,2024-12-06,Tornado\nProvo,2024-12-06,Hail\n`;
    const mockStream = new Readable();
    mockStream.push(mockCsvData);
    mockStream.push(null);

    (axios.get as jest.Mock).mockResolvedValue({ data: mockStream });

    try {
      const reports = await fetchStormReports('https://mock-noaa-reports.com');
      expect(reports).toEqual(
        expect.arrayContaining([
          expect.objectContaining({ location: 'Denver', type: 'Tornado' }),
          expect.objectContaining({ location: 'Provo', type: 'Hail' }),
        ]),
      );
    } finally {
      mockStream.destroy(); // Ensure the stream is closed
    }
  });

  it('should throw an error if fetching fails', async () => {
    (axios.get as jest.Mock).mockRejectedValue(new Error('Network error'));

    await expect(
      fetchStormReports('https://mock-noaa-reports.com'),
    ).rejects.toThrow('Network error');
  });
});
