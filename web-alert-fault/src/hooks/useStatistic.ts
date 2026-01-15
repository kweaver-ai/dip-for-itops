import { useState } from 'react';

export interface Statistic {
  total: number;
  urgent: number;
  important: number;
  common: number;
  expired: number;
  occurred: number;
  recovered: number;
}

export const useStatistic = () => {
  const [level, setLevel] = useState<any[]>([]);
  const [status, setStatus] = useState<any[]>([]);

  const onLevelChange = (value: any[]) => {
    setLevel(value);
  }

  const onStatusChange = (value: any[]) => {
    setStatus(value);
  }

  return {
    level,
    status,
    onLevelChange,
    onStatusChange,
  }
}
