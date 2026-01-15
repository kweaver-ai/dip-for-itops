import { useEffect, useRef } from 'react';

export interface Opts {
  refreshInterval: number;
  timeRange: number[];
  level: any[];
  status: any[];
  keyword?: string;
  refreshPage: () => Promise<void> | void;
}

export const useRefreshPage = (opts: Opts) => {
  const { refreshInterval, timeRange, level, status, keyword, refreshPage } = opts;
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (refreshInterval <= 0) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }

      return;
    }
    const intervalId = setInterval(refreshPage, refreshInterval * 1000);

    intervalRef.current = intervalId;

    return () => clearInterval(intervalId);
  }, [refreshInterval, timeRange]);

  useEffect(() => {
    refreshPage();
  }, [timeRange]);

  useEffect(() => {
    refreshPage();
  }, [level]);

  useEffect(() => {
    refreshPage();
  }, [status]);

  useEffect(() => {
    if (keyword !== undefined) refreshPage();
  }, [keyword]);
};
