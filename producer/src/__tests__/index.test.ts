import { fetchStormReports } from '../index';
import axios from 'axios';
import { Readable } from 'stream';

jest.mock('axios');

describe('fetchStormReports', () => {
  it('should fetch and parse storm reports', async () => {
    const mockCsvData = `location,date,type\nDenver,2024-12-06,Tornado\nProvo,2024-12-06,Hail\n`;
    const mockStream = new Readable();
    mockStream.push(mockCsvData);
    mockStream.push(null);

    (axios.get as jest.Mock).mockResolvedValue({ data: mockStream });

    const reports = await fetchStormReports('https://mock-noaa-reports.com');
    console.log("reps", reports)
    expect(reports).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ location: 'Denver', type: 'Tornado' }),
        expect.objectContaining({ location: 'Provo', type: 'Hail' }),
      ]),
    );

    // Verify publishToKafka was called (optional)
    // expect(publishToKafka).toHaveBeenCalledTimes(1);
  });

  it('should throw an error if fetching fails', async () => {
    (axios.get as jest.Mock).mockRejectedValue(new Error('Network error'));

    await expect(
      fetchStormReports('https://mock-noaa-reports.com'),
    ).rejects.toThrow('Network error');
  });
});
